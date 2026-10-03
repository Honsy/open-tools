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
	weiboHotURL = "https://weibo.com/ajax/side/hotSearch"
	baiduHotURL = "https://top.baidu.com/api/board?platform=wise&tab=realtime"
	hotLimit    = 10
)

var hotHTTP = &http.Client{Timeout: 6 * time.Second}

type hotSnap struct {
	at    time.Time
	weibo []boardItem
	baidu []boardItem
}

var (
	hotMu       sync.Mutex
	hotFetching bool
	hotState    hotSnap
)

func homeBoards(dbLinks func() (popular, latest []boardItem)) ([]boardView, bool) {
	weibo, baidu := cachedHot()
	boards := make([]boardView, 0, 4)
	if len(weibo) > 0 {
		boards = append(boards, boardView{Key: "weibo", Name: "微博", On: true, Items: weibo})
	}
	if len(baidu) > 0 {
		boards = append(boards, boardView{Key: "baidu", Name: "百度", On: len(boards) == 0, Items: baidu})
	}
	live := len(weibo) > 0 || len(baidu) > 0
	popular, latest := dbLinks()
	boards = append(boards,
		boardView{Key: "hot", Name: "站内热门", On: !live, Items: popular},
		boardView{Key: "new", Name: "刚刚收录", Items: latest},
	)
	return boards, live
}

func cachedHot() ([]boardItem, []boardItem) {
	hotMu.Lock()
	if freshHot() {
		weibo, baidu := hotState.weibo, hotState.baidu
		hotMu.Unlock()
		return weibo, baidu
	}
	if hotFetching {
		weibo, baidu := hotState.weibo, hotState.baidu
		hotMu.Unlock()
		return weibo, baidu
	}
	hotFetching = true
	hotMu.Unlock()

	weibo, baidu := fetchHot()

	hotMu.Lock()
	hotFetching = false
	if len(weibo) > 0 || len(baidu) > 0 || hotState.at.IsZero() {
		hotState = hotSnap{at: time.Now(), weibo: weibo, baidu: baidu}
	} else {
		hotState.at = time.Now()
	}
	weibo, baidu = hotState.weibo, hotState.baidu
	hotMu.Unlock()
	return weibo, baidu
}

func freshHot() bool {
	if hotState.at.IsZero() {
		return false
	}
	ttl := 10 * time.Minute
	if len(hotState.weibo) == 0 && len(hotState.baidu) == 0 {
		ttl = 2 * time.Minute
	}
	return time.Since(hotState.at) < ttl
}

func fetchHot() ([]boardItem, []boardItem) {
	var weibo, baidu []boardItem
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		body, err := getHot(weiboHotURL, "https://weibo.com/")
		if err != nil {
			return
		}
		weibo = parseWeibo(body)
	}()
	go func() {
		defer wg.Done()
		body, err := getHot(baiduHotURL, "https://top.baidu.com/board?tab=realtime")
		if err != nil {
			return
		}
		baidu = parseBaidu(body)
	}()
	wg.Wait()
	return weibo, baidu
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

func heatWan(n int) string {
	if n >= 10000 {
		return strconv.Itoa(n/10000) + "万"
	}
	if n <= 0 {
		return ""
	}
	return formatCount(n)
}
