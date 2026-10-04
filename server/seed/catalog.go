package seed

import (
	"strings"
	"time"

	"opentools/models"

	"gorm.io/gorm"
)

type repair struct {
	From   string
	Name   string
	OldURL string
	URL    string
	Desc   string
	Region string
}

// repairs only touch a row while it still has the old address, so a later
// edit in the admin is left alone.
var repairs = []repair{
	{From: "Unsplash", Name: "花瓣", OldURL: "https://unsplash.com", URL: "https://huaban.com", Desc: "图片采集和灵感", Region: "中国"},
	{From: "Pexels", Name: "图虫", OldURL: "https://www.pexels.com", URL: "https://tuchong.com", Desc: "摄影师社区", Region: "中国"},
	{From: "ChatGPT", Name: "DeepSeek", OldURL: "https://chatgpt.com", URL: "https://chat.deepseek.com", Desc: "对话和代码", Region: "中国"},
	{From: "Claude", Name: "文心一言", OldURL: "https://claude.ai", URL: "https://yiyan.baidu.com", Desc: "对话", Region: "中国"},
	{From: "Midjourney", Name: "可灵", OldURL: "https://www.midjourney.com", URL: "https://klingai.com", Desc: "图像和视频", Region: "中国"},
	{From: "格式化", OldURL: "/tools/prettier", URL: "https://www.json.cn", Desc: "在线 JSON 格式化", Region: "中国"},
	{From: "加密解密", OldURL: "/tools/crypto", URL: "https://tool.oschina.net/encrypt", Desc: "MD5 和 Base64", Region: "中国"},
	{From: "进制转换", OldURL: "/tools/hexconvert", URL: "https://tool.oschina.net/hexconvert", Desc: "二、八、十、十六进制", Region: "中国"},
	{From: "时间戳", OldURL: "/tools/moment", URL: "https://www.matools.com/timestamp", Desc: "时间和时间戳互转", Region: "中国"},
	{From: "颜色", OldURL: "/tools/rgb", URL: "https://zhongguose.com", Desc: "中国传统色对照", Region: "中国"},
	{From: "计算器", OldURL: "/tools/calculator", URL: "https://www.zxgj.cn/g/jisuanqi", Desc: "在线计算器", Region: "中国"},
	{From: "Protobuf", Name: "正则", OldURL: "/tools/protobuf", URL: "https://c.runoob.com/front-end/854", Desc: "在线测试正则表达式", Region: "中国"},
	{From: "人民币大写", OldURL: "/tools/rmbconvert", URL: "https://www.zxgj.cn/g/rmbdaxie", Desc: "金额转中文大写", Region: "中国"},
}

// EnsureCatalog rewrites known dead addresses and adds sites that are not
// in the database yet. Seed.Run does not run again after the first boot.
func EnsureCatalog(db *gorm.DB) error {
	if err := db.Model(&models.Category{}).Where("slug = ?", "ai").Update("show_on_home", true).Error; err != nil {
		return err
	}
	for _, row := range categories() {
		var n int64
		if err := db.Model(&models.Category{}).Where("slug = ?", row.Slug).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := db.Create(&row).Error; err != nil {
				return err
			}
		}
	}
	for _, row := range tags() {
		var n int64
		if err := db.Model(&models.Tag{}).Where("`group` = ? AND name = ?", row.Group, row.Name).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := db.Create(&row).Error; err != nil {
				return err
			}
		}
	}
	for _, fix := range repairs {
		var current models.Link
		err := db.Where("name = ? AND url = ?", fix.From, fix.OldURL).First(&current).Error
		if err != nil {
			continue
		}
		name := fix.From
		if fix.Name != "" {
			name = fix.Name
		}
		updates := map[string]any{
			"url":  fix.URL,
			"desc": fix.Desc,
		}
		if fix.Region != "" {
			updates["region"] = fix.Region
		}
		if body, ok := intros[name]; ok {
			if strings.Contains(current.Body, "legacy") || name != fix.From {
				updates["body"] = body
			}
		}
		if name != fix.From {
			var n int64
			if err := db.Model(&models.Link{}).Where("name = ?", name).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				if err := db.Model(&current).Update("status", "off").Error; err != nil {
					return err
				}
				continue
			}
			updates["name"] = name
			if tags, ok := siteTags[name]; ok {
				updates["tags"] = tags
			}
			if card, ok := profiles[name]; ok {
				if card.Alias != "" {
					updates["alias"] = card.Alias
				}
				if card.Lang != "" {
					updates["lang"] = card.Lang
				}
				if card.Region != "" {
					updates["region"] = card.Region
				}
			}
		}
		if err := db.Model(&current).Updates(updates).Error; err != nil {
			return err
		}
	}
	rows := linkRows()
	for i := range rows {
		link := rows[i]
		var n int64
		if err := db.Model(&models.Link{}).Where("name = ?", link.Name).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		link.Clicks = 0
		link.Views = 0
		link.Pinned = false
		link.PinSort = 0
		link.CreatedAt = time.Now().Add(-time.Minute * time.Duration(i))
		if tags, ok := siteTags[link.Name]; ok && link.Tags == "" {
			link.Tags = tags
		}
		if card, ok := profiles[link.Name]; ok {
			if link.Alias == "" {
				link.Alias = card.Alias
			}
			if link.Lang == "" {
				link.Lang = card.Lang
			}
			if link.Region == "" {
				link.Region = card.Region
			}
		}
		FillSlug(db, &link)
		if err := db.Create(&link).Error; err != nil {
			return err
		}
	}
	// 这两条原先收在实用查询。分类还停在那里时，归到翻译和地图，方便顶栏下拉。
	moves := map[string]string{"百度翻译": "translate", "高德地图": "map"}
	for name, slug := range moves {
		if err := db.Model(&models.Link{}).Where("name = ? AND category_slug = ?", name, "query").Update("category_slug", slug).Error; err != nil {
			return err
		}
	}
	return nil
}
