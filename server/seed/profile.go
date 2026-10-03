package seed

import (
	"opentools/models"

	"gorm.io/gorm"
)

type profile struct {
	Alias  string
	Lang   string
	Region string
}

// EnsureProfile fills alias, language and region only when those columns are still empty.
func EnsureProfile(db *gorm.DB) error {
	var links []models.Link
	if err := db.Find(&links).Error; err != nil {
		return err
	}
	for i := range links {
		card, ok := profiles[links[i].Name]
		if !ok {
			continue
		}
		updates := map[string]any{}
		if links[i].Alias == "" && card.Alias != "" {
			updates["alias"] = card.Alias
		}
		if links[i].Lang == "" && card.Lang != "" {
			updates["lang"] = card.Lang
		}
		if links[i].Region == "" && card.Region != "" {
			updates["region"] = card.Region
		}
		if len(updates) == 0 {
			continue
		}
		if err := db.Model(&links[i]).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func p(alias, lang, region string) profile {
	return profile{Alias: alias, Lang: lang, Region: region}
}

var profiles = map[string]profile{
	"哔哩哔哩":     p("B站", "中文", "中国"),
	"爱奇艺":      p("", "中文", "中国"),
	"腾讯视频":     p("", "中文", "中国"),
	"优酷":       p("", "中文", "中国"),
	"芒果TV":     p("", "中文", "中国"),
	"豆瓣电影":     p("", "中文", "中国"),
	"哔哩哔哩番剧":   p("B站番剧", "中文", "中国"),
	"AcFun":    p("A站", "中文", "中国"),
	"腾讯视频动漫":   p("", "中文", "中国"),
	"哔哩哔哩漫画":   p("", "中文", "中国"),
	"快看漫画":     p("", "中文", "中国"),
	"腾讯动漫":     p("", "中文", "中国"),
	"网易云音乐":    p("网易云", "中文", "中国"),
	"QQ音乐":     p("", "中文", "中国"),
	"酷狗音乐":     p("", "中文", "中国"),
	"起点中文网":    p("起点", "中文", "中国"),
	"晋江文学城":    p("晋江", "中文", "中国"),
	"豆瓣阅读":     p("", "中文", "中国"),
	"哔哩哔哩直播":   p("", "中文", "中国"),
	"斗鱼":       p("", "中文", "中国"),
	"虎牙直播":     p("", "中文", "中国"),
	"Unsplash": p("", "英文", "美国"),
	"Pexels":   p("", "英文", "美国"),
	"Wallhaven": p("", "英文", ""),
	"Steam":    p("", "英文", "美国"),
	"TapTap":   p("", "中文", "中国"),
	"4399":     p("", "中文", "中国"),
	"MDN":      p("", "英文", "美国"),
	"菜鸟教程":     p("", "中文", "中国"),
	"中国大学MOOC": p("慕课", "中文", "中国"),
	"可汗学院":     p("", "英文", "美国"),
	"网易公开课":    p("", "中文", "中国"),
	"知乎":       p("", "中文", "中国"),
	"少数派":      p("", "中文", "中国"),
	"小众软件":     p("", "中文", "中国"),
	"Product Hunt": p("", "英文", "美国"),
	"Hacker News":  p("", "英文", "美国"),
	"iconfont": p("", "中文", "中国"),
	"站酷":       p("", "中文", "中国"),
	"Dribbble": p("", "英文", "美国"),
	"GitHub":   p("", "英文", "美国"),
	"Gitee":    p("", "中文", "中国"),
	"腾讯文档":     p("", "中文", "中国"),
	"Microsoft": p("微软", "中文", "美国"),
	"阿里云盘":     p("", "中文", "中国"),
	"百度网盘":     p("", "中文", "中国"),
	"夸克网盘":     p("", "中文", "中国"),
	"坚果云":      p("", "中文", "中国"),
	"格式化":      p("", "中文", "本站"),
	"加密解密":     p("", "中文", "本站"),
	"进制转换":     p("", "中文", "本站"),
	"时间戳":      p("", "中文", "本站"),
	"颜色":       p("", "中文", "本站"),
	"计算器":      p("", "中文", "本站"),
	"Protobuf": p("", "中文", "本站"),
	"人民币大写":    p("", "中文", "本站"),
	"百度":       p("", "中文", "中国"),
	"中国天气":     p("", "中文", "中国"),
	"12306":    p("", "中文", "中国"),
	"快递100":    p("", "中文", "中国"),
	"XE 汇率":    p("XE", "英文", "加拿大"),
	"Kimi":     p("", "中文", "中国"),
	"ChatGPT":  p("", "英文", "美国"),
	"Claude":   p("", "英文", "美国"),
	"通义千问":     p("千问", "中文", "中国"),
	"豆包":       p("", "中文", "中国"),
	"即梦":       p("", "中文", "中国"),
	"Midjourney": p("MJ", "英文", "美国"),
}
