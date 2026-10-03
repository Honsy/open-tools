package seed

import "fmt"
import "time"

type Event struct {
	Year int    `json:"year"`
	Text string `json:"text"`
}

func Today(now time.Time) (string, []Event) {
	weeks := [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	label := fmt.Sprintf("%d月%d日 %s", int(now.Month()), now.Day(), weeks[now.Weekday()])
	key := fmt.Sprintf("%02d-%02d", int(now.Month()), now.Day())
	list := events[key]
	if len(list) == 0 {
		return label, []Event{{Text: "这一天还没有收录事件。"}}
	}
	return label, append([]Event(nil), list...)
}

var events = map[string][]Event{
	"01-01": {{Year: 1863, Text: "美国《解放奴隶宣言》生效。"}},
	"01-24": {{Year: 1984, Text: "苹果麦金塔个人电脑开始发售。"}},
	"02-11": {{Year: 1847, Text: "爱迪生出生。"}},
	"03-07": {{Year: 1876, Text: "贝尔获得电话专利。"}},
	"03-14": {{Year: 1879, Text: "爱因斯坦出生。"}},
	"04-12": {{Year: 1961, Text: "加加林完成人类首次载人航天。"}},
	"04-23": {{Year: 1995, Text: "联合国教科文组织把 4 月 23 日定为世界读书日。"}},
	"05-04": {{Year: 1919, Text: "五四运动爆发。"}},
	"06-05": {{Year: 1972, Text: "联合国把 6 月 5 日定为世界环境日。"}},
	"07-01": {{Year: 1997, Text: "香港回归中国。"}},
	"07-20": {{Year: 1969, Text: "阿波罗 11 号登月。"}},
	"08-08": {{Year: 2008, Text: "北京奥运会开幕。"}},
	"09-10": {{Year: 1985, Text: "中国第一个教师节。"}},
	"10-01": {{Year: 1949, Text: "中华人民共和国成立，开国大典举行。"}},
	"10-03": {
		{Year: 1990, Text: "两德正式统一。"},
		{Year: 1929, Text: "塞尔维亚-克罗地亚-斯洛文尼亚王国改名为南斯拉夫王国。"},
	},
	"11-09": {{Year: 1989, Text: "柏林墙开放。"}},
	"12-17": {{Year: 1903, Text: "莱特兄弟完成首次有动力、可持续的飞行。"}},
}
