package linkcheck

// failLimit is how many confirmed hard failures, on separate rounds, take a
// link off the public catalog. A working spare address is used immediately.
const failLimit = 3

// Decision is what to store after one probe round.
type Decision struct {
	URL    string
	Spare  string
	Status string
	Fails  int
	Note   string
}

// Decide picks the address visitors should open.
// A live primary stays, and a permanent redirect or http→https upgrade replaces it.
// A confirmed-dead primary is replaced by a live spare. With no spare, three
// consecutive hard failures take the link off the shelf. Timeouts and 5xx do not.
func Decide(url, spare, status string, fails int, primary, spareProbe Result) Decision {
	out := Decision{URL: url, Spare: spare, Status: status, Fails: fails}
	if primary.Alive {
		out.Fails = 0
		next := cleanTarget(primary.Final)
		if next != "" && !sameURL(next, url) && (primary.Permanent || httpsUpgrade(url, next)) {
			out.URL = next
			if spare == "" && !sameURL(url, next) {
				out.Spare = url
			}
			out.Note = "改成跳转后的地址"
		}
		return out
	}
	if !primary.HardDead {
		out.Note = "暂时打不开，先不改"
		return out
	}
	alt := cleanTarget(spareProbe.Final)
	if alt == "" && spareProbe.Alive {
		alt = cleanTarget(spare)
	}
	if spareProbe.Alive && alt != "" && !sameURL(alt, url) {
		out.URL = alt
		out.Spare = url
		out.Fails = 0
		out.Note = "主地址失效，改用备用"
		return out
	}
	out.Fails = fails + 1
	out.Note = "主地址失效"
	if out.Fails >= failLimit {
		out.Status = "off"
		out.Note = "连续失效，下架"
	}
	return out
}

func cleanTarget(raw string) string {
	raw = trimURL(raw)
	if len(raw) > 512 || len(raw) == 0 {
		return ""
	}
	if len(raw) < 8 {
		return ""
	}
	if raw[:7] != "http://" && (len(raw) < 8 || raw[:8] != "https://") {
		return ""
	}
	return raw
}

func sameURL(a, b string) bool {
	return trimSlash(a) == trimSlash(b)
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
