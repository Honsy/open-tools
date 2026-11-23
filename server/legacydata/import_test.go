package legacydata

import "testing"

func TestSiteIdentityDeduplicatesAcrossProtocolAndTracking(t *testing.T) {
	expected := "example.com/tools?a=2&b=1"
	actual := siteIdentity("https://EXAMPLE.com:443/tools/?b=1&utm_source=old&inviteCode=campaign&a=2#featured")
	if actual != expected || actual != siteIdentity("http://example.com/tools?b=1&a=2") {
		t.Fatalf("site identity = %q, expected %q", actual, expected)
	}
	for _, invalid := range []string{"javascript:alert(1)", "/relative", "https://account:password@example.com"} {
		if got := siteIdentity(invalid); got != "" {
			t.Errorf("invalid URL %q was accepted", invalid)
		}
	}
}
