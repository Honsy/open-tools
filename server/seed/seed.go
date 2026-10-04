package seed

import (
	"time"

	"opentools/models"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	var n int64
	if err := db.Model(&models.Category{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(categories()).Error; err != nil {
			return err
		}
		links := linkRows()
		if err := tx.Create(&links).Error; err != nil {
			return err
		}
		if err := tx.Create(articles()).Error; err != nil {
			return err
		}
		return tx.Create(tags()).Error
	})
}

func cat(slug, name, parent, kind, content string, sort int, home bool) models.Category {
	return models.Category{
		Slug: slug, Name: name, ParentSlug: parent, Kind: kind,
		ContentKind: content, Sort: sort, ShowOnHome: home,
	}
}

func categories() []models.Category {
	return []models.Category{
		cat("fun", "休闲娱乐", "", "section", "links", 1, true),
		cat("video", "在线影视", "fun", "tab", "links", 1, false),
		cat("anime", "动漫", "fun", "tab", "links", 2, false),
		cat("comic", "漫画", "fun", "tab", "links", 3, false),
		cat("music", "音乐", "fun", "tab", "links", 4, false),
		cat("novel", "小说", "fun", "tab", "links", 5, false),
		cat("live", "直播", "fun", "tab", "links", 6, false),
		cat("wallpaper", "壁纸", "fun", "tab", "links", 7, false),
		cat("game", "小游戏", "fun", "tab", "links", 8, false),

		cat("discover", "探索发现", "", "section", "mixed", 2, true),
		cat("daily", "每日一篇", "discover", "tab", "articles", 1, false),
		cat("learn", "学习网站", "discover", "tab", "links", 2, false),
		cat("cool", "趣味酷站", "discover", "tab", "links", 3, false),
		cat("design", "素材创意", "discover", "tab", "links", 4, false),

		cat("resource", "资源工具", "", "section", "links", 3, true),
		cat("software", "软件", "resource", "tab", "links", 1, false),
		cat("disk", "网盘", "resource", "tab", "links", 2, false),
		cat("webtool", "网页工具", "resource", "tab", "links", 3, false),
		cat("query", "实用查询", "resource", "tab", "links", 4, false),
		cat("res", "资源", "resource", "tab", "links", 5, false),

		cat("collection", "网址集", "", "section", "tags", 4, true),

		cat("ai", "AI 工具", "", "section", "links", 5, true),
		cat("chat", "对话", "ai", "tab", "links", 1, false),
		cat("draw", "创作", "ai", "tab", "links", 2, false),

		cat("life", "生活常用", "", "section", "links", 6, true),
		cat("shop", "购物", "life", "tab", "links", 1, false),
		cat("job", "招聘", "life", "tab", "links", 2, false),
		cat("finance", "财经", "life", "tab", "links", 3, false),
		cat("news", "资讯", "life", "tab", "links", 4, false),
		cat("community", "社区", "life", "tab", "links", 5, false),
		cat("mail", "邮箱", "life", "tab", "links", 6, false),
	}
}

func L(catSlug, name, url, desc string, sort, clicks, pin int) models.Link {
	return models.Link{
		CategorySlug: catSlug,
		Name:         name,
		URL:          url,
		Desc:         desc,
		Sort:         sort,
		Clicks:       clicks,
		Pinned:       pin > 0,
		PinSort:      pin,
		Status:       "online",
	}
}

func linkRows() []models.Link {
	links := []models.Link{
		L("video", "哔哩哔哩", "https://www.bilibili.com", "弹幕视频、番剧和直播", 1, 48, 1),
		L("video", "爱奇艺", "https://www.iqiyi.com", "正版电影和剧集", 2, 20, 0),
		L("video", "腾讯视频", "https://v.qq.com", "电视剧和综艺", 3, 18, 0),
		L("video", "优酷", "https://www.youku.com", "电影和剧集", 4, 12, 0),
		L("video", "芒果TV", "https://www.mgtv.com", "综艺和电视剧", 5, 11, 0),
		L("video", "豆瓣电影", "https://movie.douban.com", "影片资料和评分", 6, 15, 0),

		L("anime", "哔哩哔哩番剧", "https://www.bilibili.com/anime", "连载番剧和新番表", 1, 22, 0),
		L("anime", "AcFun", "https://www.acfun.cn", "弹幕视频社区", 2, 8, 0),
		L("anime", "腾讯视频动漫", "https://v.qq.com/channel/cartoon", "正版动画", 3, 7, 0),

		L("comic", "哔哩哔哩漫画", "https://manga.bilibili.com", "正版漫画", 1, 9, 0),
		L("comic", "快看漫画", "https://www.kuaikanmanhua.com", "条漫和连载", 2, 8, 0),
		L("comic", "腾讯动漫", "https://ac.qq.com", "正版漫画", 3, 6, 0),

		L("music", "网易云音乐", "https://music.163.com", "歌单和播客", 1, 36, 5),
		L("music", "QQ音乐", "https://y.qq.com", "正版曲库", 2, 14, 0),
		L("music", "酷狗音乐", "https://www.kugou.com", "音乐播放", 3, 6, 0),

		L("novel", "起点中文网", "https://www.qidian.com", "网络小说", 1, 16, 0),
		L("novel", "晋江文学城", "https://www.jjwxc.net", "原创文学", 2, 10, 0),
		L("novel", "豆瓣阅读", "https://read.douban.com", "电子书", 3, 5, 0),

		L("live", "哔哩哔哩直播", "https://live.bilibili.com", "游戏和生活直播", 1, 13, 0),
		L("live", "斗鱼", "https://www.douyu.com", "游戏直播", 2, 9, 0),
		L("live", "虎牙直播", "https://www.huya.com", "游戏直播", 3, 7, 0),

		L("video", "央视频", "https://www.yangshipin.cn", "央视正版直播和回看", 7, 0, 0),
		L("video", "咪咕视频", "https://www.miguvideo.com", "体育和影视", 8, 0, 0),
		L("video", "抖音", "https://www.douyin.com", "短视频", 9, 0, 0),

		L("music", "酷我音乐", "https://www.kuwo.cn", "正版曲库", 4, 0, 0),
		L("music", "喜马拉雅", "https://www.ximalaya.com", "有声书和播客", 5, 0, 0),

		L("novel", "微信读书", "https://weread.qq.com", "电子书", 4, 0, 0),
		L("novel", "纵横中文网", "https://www.zongheng.com", "网络小说", 5, 0, 0),
		L("novel", "番茄小说", "https://fanqienovel.com", "免费网文", 6, 0, 0),

		L("live", "快手", "https://www.kuaishou.com", "短视频和直播", 4, 0, 0),

		L("wallpaper", "花瓣", "https://huaban.com", "图片采集和灵感", 1, 0, 0),
		L("wallpaper", "图虫", "https://tuchong.com", "摄影师社区", 2, 0, 0),
		L("wallpaper", "Wallhaven", "https://wallhaven.cc", "壁纸收藏", 3, 5, 0),

		L("game", "Steam", "https://store.steampowered.com", "电脑游戏商店", 1, 21, 0),
		L("game", "TapTap", "https://www.taptap.cn", "手机游戏社区", 2, 12, 0),
		L("game", "4399", "https://www.4399.com", "小游戏", 3, 9, 0),
		L("game", "7k7k", "https://www.7k7k.com", "小游戏", 4, 0, 0),
		L("game", "腾讯游戏", "https://game.qq.com", "腾讯游戏官网", 5, 0, 0),

		L("learn", "MDN", "https://developer.mozilla.org", "Web 开发文档", 1, 28, 9),
		L("learn", "菜鸟教程", "https://www.runoob.com", "入门示例", 2, 19, 0),
		L("learn", "中国大学MOOC", "https://www.icourse163.org", "大学公开课", 3, 14, 0),
		L("learn", "学堂在线", "https://www.xuetangx.com", "大学公开课", 4, 0, 0),
		L("learn", "网易云课堂", "https://study.163.com", "职业课程", 5, 0, 0),
		L("learn", "可汗学院", "https://www.khanacademy.org", "基础学科课程", 6, 8, 0),
		L("learn", "网易公开课", "https://open.163.com", "公开课程", 7, 7, 0),
		L("learn", "掘金", "https://juejin.cn", "技术文章", 8, 0, 0),
		L("learn", "廖雪峰", "https://www.liaoxuefeng.com", "编程教程", 9, 0, 0),

		L("cool", "知乎", "https://www.zhihu.com", "问答社区", 1, 33, 2),
		L("cool", "少数派", "https://sspai.com", "效率和工作方法", 2, 17, 0),
		L("cool", "小众软件", "https://www.appinn.com", "软件和工具介绍", 3, 24, 10),
		L("cool", "什么值得买", "https://www.smzdm.com", "商品价格和评测", 4, 0, 0),
		L("cool", "V2EX", "https://www.v2ex.com", "技术社区", 5, 0, 0),
		L("cool", "Product Hunt", "https://www.producthunt.com", "新产品发布", 6, 6, 0),
		L("cool", "Hacker News", "https://news.ycombinator.com", "技术讨论", 7, 8, 0),

		L("design", "iconfont", "https://www.iconfont.cn", "图标", 1, 18, 0),
		L("design", "站酷", "https://www.zcool.com.cn", "设计作品", 2, 11, 0),
		L("design", "即时设计", "https://js.design", "在线界面设计", 3, 0, 0),
		L("design", "稿定设计", "https://www.gaoding.com", "模板和图片编辑", 4, 0, 0),
		L("design", "Dribbble", "https://dribbble.com", "界面设计", 5, 7, 0),

		L("software", "GitHub", "https://github.com", "代码托管", 1, 40, 3),
		L("software", "Gitee", "https://gitee.com", "代码托管", 2, 12, 0),
		L("software", "腾讯文档", "https://docs.qq.com", "在线文档和表格", 3, 15, 6),
		L("software", "飞书", "https://www.feishu.cn", "文档、表格和会议", 4, 0, 0),
		L("software", "语雀", "https://www.yuque.com", "知识库和文档", 5, 0, 0),
		L("software", "WPS", "https://www.wps.cn", "文字、表格和演示", 6, 0, 0),
		L("software", "腾讯会议", "https://meeting.tencent.com", "视频会议", 7, 0, 0),
		L("software", "Microsoft", "https://www.microsoft.com/zh-cn", "系统和官方软件", 8, 5, 0),

		L("disk", "阿里云盘", "https://www.alipan.com", "网盘", 1, 26, 7),
		L("disk", "百度网盘", "https://pan.baidu.com", "网盘", 2, 16, 0),
		L("disk", "夸克网盘", "https://pan.quark.cn", "网盘", 3, 10, 0),
		L("disk", "坚果云", "https://www.jianguoyun.com", "同步盘", 4, 6, 0),
		L("disk", "天翼云盘", "https://cloud.189.cn", "网盘", 5, 0, 0),

		L("webtool", "格式化", "https://www.json.cn", "在线 JSON 格式化", 1, 4, 0),
		L("webtool", "加密解密", "https://tool.oschina.net/encrypt", "MD5 和 Base64", 2, 3, 0),
		L("webtool", "进制转换", "https://tool.oschina.net/hexconvert", "二、八、十、十六进制", 3, 5, 0),
		L("webtool", "时间戳", "https://www.matools.com/timestamp", "时间和时间戳互转", 4, 4, 0),
		L("webtool", "颜色", "https://zhongguose.com", "中国传统色对照", 5, 3, 0),
		L("webtool", "计算器", "https://www.zxgj.cn/g/jisuanqi", "在线计算器", 6, 2, 0),
		L("webtool", "正则", "https://c.runoob.com/front-end/854", "在线测试正则表达式", 7, 2, 0),
		L("webtool", "人民币大写", "https://www.zxgj.cn/g/rmbdaxie", "金额转中文大写", 8, 2, 0),

		L("query", "百度", "https://www.baidu.com", "网页搜索", 1, 30, 4),
		L("query", "中国天气", "https://www.weather.com.cn", "天气预报查询", 2, 9, 0),
		L("query", "12306", "https://www.12306.cn", "火车票查询", 3, 14, 0),
		L("query", "快递100", "https://www.kuaidi100.com", "快递物流查询", 4, 8, 0),
		L("query", "高德地图", "https://www.amap.com", "地图和路线", 5, 0, 0),
		L("query", "百度翻译", "https://fanyi.baidu.com", "在线翻译", 6, 0, 0),
		L("query", "携程", "https://www.ctrip.com", "机票和酒店", 7, 0, 0),
		L("query", "航旅纵横", "https://www.umetrip.com", "航班动态", 8, 0, 0),
		L("query", "新浪外汇", "https://finance.sina.com.cn/forex/", "汇率行情", 9, 0, 0),
		L("query", "天眼查", "https://www.tianyancha.com", "企业信息查询", 10, 0, 0),
		L("query", "XE 汇率", "https://www.xe.com", "汇率换算", 11, 4, 0),

		L("chat", "Kimi", "https://kimi.moonshot.cn", "长文对话", 1, 27, 8),
		L("chat", "DeepSeek", "https://chat.deepseek.com", "对话和代码", 2, 0, 0),
		L("chat", "文心一言", "https://yiyan.baidu.com", "对话", 3, 0, 0),
		L("chat", "通义千问", "https://tongyi.aliyun.com", "对话", 4, 9, 0),
		L("chat", "豆包", "https://www.doubao.com", "对话", 5, 8, 0),
		L("chat", "腾讯元宝", "https://yuanbao.tencent.com", "对话", 6, 0, 0),
		L("chat", "智谱清言", "https://chatglm.cn", "对话", 7, 0, 0),
		L("chat", "秘塔", "https://metaso.cn", "带来源的搜索", 8, 0, 0),
		L("draw", "即梦", "https://jimeng.jianying.com", "图像和视频", 1, 7, 0),
		L("draw", "可灵", "https://klingai.com", "图像和视频", 2, 0, 0),
		L("draw", "文心一格", "https://yige.baidu.com", "图像", 3, 0, 0),
	}
	links = append(links, moreLinks()...)

	now := time.Now()
	fresh := map[string]int{
		"即梦": 1, "豆包": 2, "夸克网盘": 3, "坚果云": 4,
		"可汗学院": 5, "TapTap": 6, "快递100": 7, "iconfont": 8,
		"虎牙直播": 9, "网易公开课": 10, "可灵": 11, "腾讯文档": 12,
	}
	for i := range links {
		if body, ok := intros[links[i].Name]; ok {
			links[i].Body = body
		} else if links[i].Desc != "" {
			links[i].Body = links[i].Name + "，" + links[i].Desc + "。先看它是不是你要找的那一个，再打开。\n\n页面改版、要登录或分地区开放时，以网站自己的说明为准。"
		}
		links[i].CreatedAt = now.Add(-time.Hour * time.Duration(100+i))
		if h, ok := fresh[links[i].Name]; ok {
			links[i].CreatedAt = now.Add(-time.Minute * time.Duration(h))
		}
	}
	return links
}

func articles() []models.Article {
	now := time.Now()
	return []models.Article{
		{
			Title:     "网盘里的文件，先分成要同步和只是囤着",
			Summary:   "同步盘适合正在改的文档。装机包和影视放进容量大的网盘，不要跟每天要打开的文件挤在一起。",
			Body:      "同步盘的价值是几台设备上的同一份文件始终一致，适合稿子、表格和正在写的代码。装机镜像、录屏和压缩包几乎不会改，放进去只会把同步队列拖慢。\n\n一个能长期用的分法：正在改的进同步盘，一个月没打开的挪到普通网盘，确定不会再用的就删。导航上把这两类网盘分开收藏，比在一个客户端里翻目录快。",
			Views:     128,
			CreatedAt: now.Add(-48 * time.Hour),
		},
		{
			Title:     "工具变多以后，导航比一排菜单好找",
			Summary:   "计算器、进制、时间戳这些小工具还在，只是不再占据整个首页。首页改成按场景找网站。",
			Body:      "旧站把每个工具做成一个顶部入口。工具少的时候这样清楚，工具一多，菜单就变成一串记不住的名字。\n\n现在这些工具收在「资源工具 / 网页工具」里，源码留在仓库的 legacy 目录。新首页只回答一件事：这件事该打开哪个网站。工具本身下一期再迁进新界面。",
			Views:     86,
			CreatedAt: now.Add(-24 * time.Hour),
		},
		{
			Title:     "学习网站按正在看收，不要按以后再说收",
			Summary:   "公开课和文档站很容易收成一长列。只把这周会打开的留在常用，其余放进分类。",
			Body:      "MDN、公开课和教程站都值得留，但常用栏只有十个位置。把「以后可能会看」的课放进常用，常用就不再常用。\n\n这周正在查的文档留在热门推荐或自己的星标里。一门课看完，就把星去掉。分类页还在，想起来的时候搜名字就能回到原处。",
			Views:     54,
			CreatedAt: now.Add(-6 * time.Hour),
		},
	}
}

func tags() []models.Tag {
	rows := []struct{ group, name string }{
		{"网址集", "设计"},
		{"网址集", "技术博客"},
		{"网址集", "公开课"},
		{"网址集", "导航"},
		{"小游戏", "独立游戏"},
		{"小游戏", "老游戏"},
		{"小游戏", "在线玩"},
		{"小游戏", "Steam"},
		{"查询", "天气"},
		{"查询", "快递"},
		{"查询", "火车票"},
		{"查询", "汇率"},
		{"工具", "JSON"},
		{"工具", "时间戳"},
		{"工具", "进制"},
		{"工具", "颜色"},
		{"资源", "软件"},
		{"资源", "字体"},
		{"资源", "模板"},
		{"资源", "字幕"},
		{"游戏", "平台"},
		{"游戏", "资讯"},
		{"AI", "对话"},
		{"AI", "图像"},
		{"AI", "视频"},
		{"生活", "购物"},
		{"生活", "招聘"},
		{"生活", "财经"},
		{"生活", "社区"},
		{"查询", "备案"},
		{"查询", "企业"},
	}
	out := make([]models.Tag, 0, len(rows))
	for i, row := range rows {
		out = append(out, models.Tag{Group: row.group, Name: row.name, Sort: i + 1})
	}
	return out
}
