package api

import (
	"testing"
	"time"
)

func TestHotCacheServesStaleDataDuringRefresh(t *testing.T) {
	old := []hotPack{{key: "old", items: []boardItem{{Title: "previous result"}}}}
	cache := hotCache{snapshot: hotSnap{at: time.Now().Add(-time.Hour), packs: old}}
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	fetch := func() []hotPack {
		started <- struct{}{}
		<-release
		return nil
	}

	result := make(chan []hotPack, 1)
	go func() { result <- cache.packs(fetch) }()
	select {
	case packs := <-result:
		if len(packs) != 1 || packs[0].key != "old" {
			t.Fatalf("first request returned %v, want stale data", packs)
		}
	case <-time.After(time.Second):
		t.Fatal("first request waited for the refresh")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("refresh did not start")
	}

	packs := cache.packs(fetch)
	if len(packs) != 1 || packs[0].key != "old" {
		t.Fatalf("request during refresh returned %v, want stale data", packs)
	}
}

func TestHotCacheRetainsDataWhenRefreshFails(t *testing.T) {
	old := []hotPack{{key: "old", items: []boardItem{{Title: "previous result"}}}}
	cache := hotCache{fetching: true, snapshot: hotSnap{at: time.Now().Add(-time.Hour), packs: old}}
	cache.refresh(func() []hotPack { return nil })
	if cache.fetching || !cache.snapshot.fresh() || cache.snapshot.packs[0].key != "old" {
		t.Fatalf("failed refresh left cache in unexpected state: %+v", cache.snapshot)
	}
}
