package api

import (
	"strings"
	"testing"

	"opentools/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestReadInternetRankCSV(t *testing.T) {
	file := "\ufeffdomain,category,visits,name,month,risk\nWWW.example.com,搜索,90400,示例,2026-10,\ncdn.example.net,工具,1200,,2026-10,blocked\n"
	rows, err := readInternetRankCSV(strings.NewReader(file), "2026-10")
	if err != nil || len(rows) != 2 {
		t.Fatalf("import = %v, %v", rows, err)
	}
	if rows[0].Domain != "example.com" || rows[0].Visits != 90400 || rows[0].Risky || !rows[1].Risky || rows[1].Name != "cdn.example.net" {
		t.Fatalf("unexpected import values: %+v", rows)
	}
	for name, input := range map[string]string{
		"domain duplicates": "domain,category,visits\nexample.com,搜索,100\nwww.example.com,工具,321\n",
		"not a domain":      "domain,category,visits\nhttps://example.com,搜索,100\n",
		"local hostname":    "domain,category,visits\nadmin.local,搜索,100\n",
		"wrong month":       "domain,category,visits,month\nexample.com,搜索,100,2026-09\n",
		"wrong risk flag":   "domain,category,visits,risk\nexample.com,搜索,100,review_me\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := readInternetRankCSV(strings.NewReader(input), "2026-10"); err == nil {
				t.Fatal("invalid CSV passed validation")
			}
		})
	}
}

func TestInternetRankExcludesProviderRiskAndBlockedDomains(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused@tcp(127.0.0.1:3306)/unused", SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var rows []models.InternetRankEntry
	query := publicInternetRank(db, "2026-10").Where("category = ?", "搜索").Order("visits desc, domain asc").Limit(100).Find(&rows)
	sql := query.Dialector.Explain(query.Statement.SQL.String(), query.Statement.Vars...)
	for _, expected := range []string{"month = '2026-10' AND risky = false", "NOT EXISTS", "b.domain = internet_rank_entries.domain", "LIKE CONCAT('%.', b.domain)", "category = '搜索'", "visits desc, domain asc LIMIT 100"} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("rank query is missing %q: %s", expected, sql)
		}
	}
}

func TestRankPreviewMatchesPublicRankingOrderAndExclusions(t *testing.T) {
	entries := []models.InternetRankEntry{
		{Domain: "risk.example.com", Visits: 400}, {Domain: "unwanted.org", Visits: 380, Risky: true},
		{Domain: "z.example.net", Visits: 200}, {Domain: "a.example.net", Visits: 200},
		{Domain: "safe.example.org", Visits: 100}, {Domain: "example.com", Visits: 500},
	}
	visible, excluded := rankPreviewEntries(entries, []string{"example.com"})
	if len(visible) != 3 || visible[0].Domain != "a.example.net" || visible[1].Domain != "z.example.net" || visible[2].Domain != "safe.example.org" {
		t.Fatalf("unexpected ranking preview: %+v", visible)
	}
	if len(excluded) != 3 || excluded[0].Reason != "命中风险名单" || excluded[1].Reason != "数据源风险标记" {
		t.Fatalf("unexpected exclusions: %+v", excluded)
	}
}
