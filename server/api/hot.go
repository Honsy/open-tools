package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	weiboHotURL  = "https://weibo.com/ajax/side/hotSearch"
	baiduHotURL  = "https://top.baidu.com/api/board?platform=wise&tab=realtime"
	zhihuHotURL  = "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true"
	biliHotURL   = "https://api.bilibili.com/x/web-interface/popular?ps=30&pn=1"
	douyinHotURL = "https://www.douyin.com/aweme/v1/web/hot/search/list/"
	hotLimit     = 30
)

var hotHTTP = &http.Client{Timeout: 6 * time.Second}

type hotPack struct {
	key, name string
	items     []boardItem
}

type hotSnap struct {
	at    time.Time
	packs []hotPack
}

var (
	hotMu       sync.Mutex
	hotFetching bool
	hotState    hotSnap
)

func homeBoards(dbLinks func() (popular, latest []boardItem)) ([]boardView, bool) {
	sources := cachedHot()
	boards := make([]boardView, 0, len(sources)+2)
	for _, src := range sources {
		if len(src.items) == 0 {
			continue
		}
		boards = append(boards, boardView{Key: src.key, Name: src.name, On: len(boards) == 0, Items: src.items})
	}
	live := len(boards) > 0
	popular, latest := dbLinks()
	boards = append(boards,
		boardView{Key: "hot", Name: "站内热门", On: !live, Items: popular},
		boardView{Key: "new", Name: "刚刚收录", Items: latest},
	)
	return boards, live
}

func cachedHot() []hotPack {
	hotMu.Lock()
	if freshHot() {
		packs := hotState.packs
		hotMu.Unlock()
		return packs
	}
	if hotFetching {
		packs := hotState.packs
		hotMu.Unlock()
		return packs
	}
	hotFetching = true
	hotMu.Unlock()

	packs := fetchHot()

	hotMu.Lock()
	hotFetching = false
	if hotHasItems(packs) || hotState.at.IsZero() {
		hotState = hotSnap{at: time.Now(), packs: packs}
	} else {
		hotState.at = time.Now()
	}
	packs = hotState.packs
	hotMu.Unlock()
	return packs
}

func hotHasItems(packs []hotPack) bool {
	for _, pack := range packs {
		if len(pack.items) > 0 {
			return true
		}
	}
	return false
}

func freshHot() bool {
	if hotState.at.IsZero() {
		return false
	}
	ttl := 10 * time.Minute
	if !hotHasItems(hotState.packs) {
		ttl = 2 * time.Minute
	}
	return time.Since(hotState.at) < ttl
}

func fetchHot() []hotPack {
	specs := []struct {
		key, name, raw, referer string
		parse                   func([]byte) []boardItem
	}{
		{"baidu", "百度", baiduHotURL, "https://top.baidu.com/board?tab=realtime", parseBaidu},
		{"weibo", "微博", weiboHotURL, "https://weibo.com/", parseWeibo},
		{"zhihu", "知乎", zhihuHotURL, "https://www.zhihu.com/hot", parseZhihu},
		{"bili", "哔哩哔哩", biliHotURL, "https://www.bilibili.com/", parseBili},
		{"douyin", "抖音", douyinHotURL, "https://www.douyin.com/", parseDouyin},
	}
	packs := make([]hotPack, len(specs))
	var wg sync.WaitGroup
	for i, spec := range specs {
		packs[i] = hotPack{key: spec.key, name: spec.name}
		wg.Add(1)
		go func(i int, spec struct {
			key, name, raw, referer string
			parse                   func([]byte) []boardItem
		}) {
			defer wg.Done()
			body, err := getHot(spec.raw, spec.referer)
			if err != nil {
				return
			}
			packs[i].items = spec.parse(body)
		}(i, spec)
	}
	wg.Wait()
	return packs
}

func getHot(raw, referer string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", referer)
	resp, err := hotHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errStatus(resp.StatusCode)
	}
	return body, nil
}

type statusError int

func (s statusError) Error() string { return "hot status " + strconv.Itoa(int(s)) }

func errStatus(code int) error { return statusError(code) }

func parseWeibo(body []byte) []boardItem {
	var payload struct {
		Data struct {
			Realtime []struct {
				Word string  `json:"word"`
				Num  float64 `json:"num"`
			} `json:"realtime"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return nil
	}
	items := make([]boardItem, 0, hotLimit)
	for _, row := range payload.Data.Realtime {
		title := strings.TrimSpace(row.Word)
		if title == "" {
			continue
		}
		items = append(items, boardItem{
			Title:    title,
			Heat:     heatWan(int(row.Num)),
			Href:     "https://s.weibo.com/weibo?q=" + url.QueryEscape(title),
			External: true,
		})
		if len(items) == hotLimit {
			break
		}
	}
	return items
}

func parseBaidu(body []byte) []boardItem {
	var payload struct {
		Data struct {
			Cards []struct {
				Content []struct {
					Word    string `json:"word"`
					HotName string `json:"newHotName"`
					Content []struct {
						Word    string `json:"word"`
						HotName string `json:"newHotName"`
					} `json:"content"`
				} `json:"content"`
			} `json:"cards"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return nil
	}
	items := make([]boardItem, 0, hotLimit)
	seen := map[string]bool{}
	add := func(word, tag string) {
		title := strings.TrimSpace(word)
		if title == "" || seen[title] || len(items) == hotLimit {
			return
		}
		seen[title] = true
		items = append(items, boardItem{
			Title:    title,
			Heat:     strings.TrimSpace(tag),
			Href:     "https://www.baidu.com/s?wd=" + url.QueryEscape(title),
			External: true,
		})
	}
	for _, card := range payload.Data.Cards {
		for _, block := range card.Content {
			add(block.Word, block.HotName)
			for _, row := range block.Content {
				add(row.Word, row.HotName)
			}
		}
	}
	return items
}

func parseZhihu(body []byte) []boardItem {
	var payload struct {
		Data []struct {
			DetailText string `json:"detail_text"`
			Target     struct {
				Title     string `json:"title"`
				URL       string `json:"url"`
				TitleArea struct {
					Text string `json:"text"`
				} `json:"title_area"`
			} `json:"target"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return nil
	}
	items := make([]boardItem, 0, hotLimit)
	seen := map[string]bool{}
	for _, row := range payload.Data {
		title := strings.TrimSpace(row.Target.Title)
		if title == "" {
			title = strings.TrimSpace(row.Target.TitleArea.Text)
		}
		if title == "" || seen[title] {
			continue
		}
		seen[title] = true
		href := strings.TrimSpace(row.Target.URL)
		if !strings.HasPrefix(href, "http") {
			href = "https://www.zhihu.com/search?type=content&q=" + url.QueryEscape(title)
		}
		items = append(items, boardItem{Title: title, Heat: strings.TrimSpace(row.DetailText), Href: href, External: true})
		if len(items) == hotLimit {
			break
		}
	}
	return items
}

func parseBili(body []byte) []boardItem {
	var payload struct {
		Data struct {
			List []struct {
				Title string `json:"title"`
				Bvid  string `json:"bvid"`
				Stat  struct {
					View int `json:"view"`
				} `json:"stat"`
			} `json:"list"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return nil
	}
	items := make([]boardItem, 0, hotLimit)
	for _, row := range payload.Data.List {
		title := strings.TrimSpace(row.Title)
		if title == "" || row.Bvid == "" {
			continue
		}
		items = append(items, boardItem{
			Title:    title,
			Heat:     heatWan(row.Stat.View),
			Href:     "https://www.bilibili.com/video/" + row.Bvid,
			External: true,
		})
		if len(items) == hotLimit {
			break
		}
	}
	return items
}

func parseDouyin(body []byte) []boardItem {
	var payload struct {
		Data struct {
			WordList []struct {
				Word     string `json:"word"`
				HotValue int    `json:"hot_value"`
			} `json:"word_list"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return nil
	}
	items := make([]boardItem, 0, hotLimit)
	for _, row := range payload.Data.WordList {
		title := strings.TrimSpace(row.Word)
		if title == "" {
			continue
		}
		items = append(items, boardItem{
			Title:    title,
			Heat:     heatWan(row.HotValue),
			Href:     "https://www.douyin.com/search/" + url.PathEscape(title),
			External: true,
		})
		if len(items) == hotLimit {
			break
		}
	}
	return items
}

func heatWan(n int) string {
	if n >= 10000 {
		return strconv.Itoa(n/10000) + "万"
	}
	if n <= 0 {
		return ""
	}
	return formatCount(n)
}
