package linkcheck

import (
	"net"
	"syscall"
	"testing"
)

func TestDecideKeepsLivePrimary(t *testing.T) {
	got := Decide("https://a.example/x", "https://b.example", "online", 2, Result{Alive: true, Final: "https://a.example/x"}, Result{})
	if got.URL != "https://a.example/x" || got.Spare != "https://b.example" || got.Fails != 0 || got.Status != "online" {
		t.Fatalf("%+v", got)
	}
}

func TestDecideFollowsPermanentRedirect(t *testing.T) {
	got := Decide("https://old.example", "", "online", 0, Result{Alive: true, Permanent: true, Final: "https://new.example/home"}, Result{})
	if got.URL != "https://new.example/home" || got.Spare != "https://old.example" || got.Note != "改成跳转后的地址" {
		t.Fatalf("%+v", got)
	}
}

func TestDecideUpgradesHTTPS(t *testing.T) {
	got := Decide("http://a.example/path", "", "online", 1, Result{Alive: true, Final: "https://a.example/path"}, Result{})
	if got.URL != "https://a.example/path" || got.Fails != 0 {
		t.Fatalf("%+v", got)
	}
}

func TestDecideIgnoresTemporaryRedirect(t *testing.T) {
	got := Decide("https://a.example", "", "online", 0, Result{Alive: true, Final: "https://other.example/login"}, Result{})
	if got.URL != "https://a.example" || got.Spare != "" {
		t.Fatalf("%+v", got)
	}
}

func TestDecideSwapsSpare(t *testing.T) {
	got := Decide("https://dead.example", "https://live.example", "online", 2, Result{HardDead: true}, Result{Alive: true, Final: "https://live.example"})
	if got.URL != "https://live.example" || got.Spare != "https://dead.example" || got.Fails != 0 || got.Status != "online" {
		t.Fatalf("%+v", got)
	}
}

func TestDecideRetiresAfterThreeMisses(t *testing.T) {
	first := Decide("https://dead.example", "", "online", 0, Result{HardDead: true}, Result{})
	if first.Status != "online" || first.Fails != 1 {
		t.Fatalf("%+v", first)
	}
	third := Decide("https://dead.example", "", "online", 2, Result{HardDead: true}, Result{})
	if third.Status != "off" || third.Fails != 3 {
		t.Fatalf("%+v", third)
	}
}

func TestDecideLeavesTimeoutAlone(t *testing.T) {
	got := Decide("https://a.example", "https://b.example", "online", 2, Result{Reason: "暂时连不上"}, Result{Alive: true})
	if got.URL != "https://a.example" || got.Fails != 2 || got.Status != "online" {
		t.Fatalf("%+v", got)
	}
}

func TestClassify(t *testing.T) {
	alive, hard, _ := classify(200, nil)
	if !alive || hard {
		t.Fatal("200")
	}
	alive, hard, _ = classify(403, nil)
	if !alive || hard {
		t.Fatal("403")
	}
	alive, hard, _ = classify(404, nil)
	if alive || !hard {
		t.Fatal("404")
	}
	alive, hard, _ = classify(503, nil)
	if alive || hard {
		t.Fatal("503")
	}
	alive, hard, _ = classify(0, &net.DNSError{IsNotFound: true})
	if alive || !hard {
		t.Fatal("dns")
	}
	alive, hard, _ = classify(0, syscall.ECONNREFUSED)
	if alive || !hard {
		t.Fatal("refused")
	}
	alive, hard, reason := classify(0, errPrivate)
	if alive || hard || reason == "" {
		t.Fatal("private")
	}
}

func TestCheckSkipsLoopback(t *testing.T) {
	got := Check("http://127.0.0.1/secret")
	if got.Alive || got.HardDead {
		t.Fatalf("%+v", got)
	}
}

func TestPublicIP(t *testing.T) {
	if publicIP(net.ParseIP("8.8.8.8")) != true {
		t.Fatal("public")
	}
	for _, raw := range []string{"127.0.0.1", "10.1.2.3", "192.168.1.1", "172.16.0.1", "169.254.1.1", "100.64.0.1", "::1"} {
		if publicIP(net.ParseIP(raw)) {
			t.Fatal(raw)
		}
	}
}
