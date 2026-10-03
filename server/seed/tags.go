package seed

import (
	"strings"

	"opentools/models"

	"gorm.io/gorm"
)

// EnsureTags fills keyword tags only when the field is still empty.
func EnsureTags(db *gorm.DB) error {
	var links []models.Link
	if err := db.Where("tags = ? OR tags IS NULL", "").Find(&links).Error; err != nil {
		return err
	}
	for i := range links {
		tags, ok := siteTags[links[i].Name]
		if !ok || tags == "" {
			continue
		}
		if err := db.Model(&links[i]).Update("tags", tags).Error; err != nil {
			return err
		}
	}
	return nil
}

func SplitTags(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == '、' || r == ';' || r == '；'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

var siteTags = map[string]string{
	"哔哩哔哩":   "弹幕,视频,番剧",
	"爱奇艺":    "电影,剧集",
	"腾讯视频":   "电视剧,综艺",
	"优酷":     "电影,剧集",
	"芒果TV":   "综艺,电视剧",
	"豆瓣电影":   "评分,影片资料",
	"哔哩哔哩番剧": "番剧,新番",
	"AcFun":  "弹幕,动画",
	"腾讯视频动漫": "动画,正版",
	"哔哩哔哩漫画": "漫画,正版",
	"快看漫画":   "条漫,连载",
	"腾讯动漫":   "漫画,正版",
	"网易云音乐":  "歌单,音乐",
	"QQ音乐":   "曲库,音乐",
	"酷狗音乐":   "音乐,播放",
	"起点中文网":  "小说,网文",
	"晋江文学城":  "小说,原创",
	"豆瓣阅读":   "电子书,阅读",
	"哔哩哔哩直播": "直播,游戏",
	"斗鱼":     "直播,游戏",
	"虎牙直播":   "直播,游戏",
	"Unsplash": "照片,免费图",
	"Pexels": "图片,视频",
	"Wallhaven": "壁纸",
	"Steam":  "游戏,商店",
	"TapTap": "手游,社区",
	"4399":   "小游戏",
	"MDN":    "文档,Web",
	"菜鸟教程":   "教程,入门",
	"中国大学MOOC": "公开课,大学",
	"可汗学院":   "课程,基础学科",
	"网易公开课":  "公开课",
	"知乎":     "问答,社区",
	"少数派":    "效率,方法",
	"小众软件":   "软件,工具",
	"Product Hunt": "新产品",
	"Hacker News": "技术,讨论",
	"iconfont": "图标,设计",
	"站酷":     "设计,作品",
	"Dribbble": "界面,设计",
	"GitHub": "代码,托管",
	"Gitee":  "代码,托管",
	"腾讯文档":   "文档,表格",
	"Microsoft": "系统,软件",
	"阿里云盘":   "网盘",
	"百度网盘":   "网盘",
	"夸克网盘":   "网盘",
	"坚果云":    "同步盘,网盘",
	"格式化":    "代码,格式化",
	"加密解密":   "摘要,加解密",
	"进制转换":   "进制",
	"时间戳":    "时间",
	"颜色":     "颜色",
	"计算器":    "计算",
	"Protobuf": "Protobuf",
	"人民币大写":  "金额",
	"百度":     "搜索",
	"中国天气":   "天气,查询",
	"12306":  "火车票,查询",
	"快递100":  "快递,物流",
	"XE 汇率":  "汇率",
	"Kimi":   "对话,AI",
	"ChatGPT": "对话,AI",
	"Claude": "对话,AI",
	"通义千问":   "对话,AI",
	"豆包":     "对话,AI",
	"即梦":     "图像,视频",
	"Midjourney": "图像,AI",
}
