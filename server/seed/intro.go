package seed

import "strings"

// IsPlaceholderIntro identifies generic introductions stored by older releases.
func IsPlaceholderIntro(body string) bool {
	const suffix = "。先看它是不是你要找的那一个，再打开。\n\n页面改版、要登录或分地区开放时，以网站自己的说明为准。"
	lead, ok := strings.CutSuffix(strings.TrimSpace(body), suffix)
	return ok && strings.Contains(lead, "，") && !strings.ContainsAny(lead, "\r\n")
}
