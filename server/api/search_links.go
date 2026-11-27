package api

import (
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"opentools/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var searchLikeEscape = strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")

func containsSearchTerm(term string) string {
	return "%" + searchLikeEscape.Replace(term) + "%"
}

func searchTerms(text string) []string {
	var terms []string
	for _, word := range strings.Fields(text) {
		start := 0
		han := false
		for offset, r := range word {
			next := unicode.Is(unicode.Han, r)
			if offset > 0 && next != han {
				terms = append(terms, word[start:offset])
				start = offset
			}
			han = next
		}
		terms = append(terms, word[start:])
	}
	return terms
}

func matchingLinks(db *gorm.DB, text string) *gorm.DB {
	query := db.Model(&models.Link{}).Where("status = ?", "online")
	for _, term := range searchTerms(text) {
		like := containsSearchTerm(term)
		query = query.Where("(name ILIKE ? ESCAPE '!' OR \"desc\" ILIKE ? ESCAPE '!' OR tags ILIKE ? ESCAPE '!' OR alias ILIKE ? ESCAPE '!' OR url ILIKE ? ESCAPE '!')",
			like, like, like, like, like)
	}
	return query
}

type searchScope struct {
	Name   string
	URL    string
	Count  int
	Active bool
}

func searchURL(text, scope string, page int) string {
	values := url.Values{"q": {text}}
	if scope != "" {
		values.Set("scope", scope)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	return "/search?" + values.Encode()
}

func searchCatalog(db *gorm.DB, text, requested string, view *searchView) ([]string, map[string]string) {
	var categories []models.Category
	db.Select("slug", "name", "parent_slug", "kind").Order("sort asc, id asc").Find(&categories)
	members := make(map[string][]string, len(categories))
	labels := make(map[string]string, len(categories))
	groups := make(map[string]string, len(categories))
	for _, category := range categories {
		labels[category.Slug] = category.Name
		root := category.Slug
		if category.ParentSlug != "" {
			root = category.ParentSlug
		}
		groups[category.Slug] = root
		members[root] = append(members[root], category.Slug)
		if category.Kind == "section" && category.Slug == requested {
			view.Scope, view.ScopeName = category.Slug, category.Name
		}
	}
	var counts []struct {
		CategorySlug string
		Count        int
	}
	matchingLinks(db, text).Select("category_slug, COUNT(*) AS count").Group("category_slug").Scan(&counts)
	var total int
	totals := make(map[string]int, len(counts))
	for _, row := range counts {
		total += row.Count
		totals[groups[row.CategorySlug]] += row.Count
	}
	view.Total = total
	view.Scopes = append(view.Scopes, searchScope{Name: "全部", URL: searchURL(text, "", 1), Count: total, Active: view.Scope == ""})
	for _, category := range categories {
		if category.Kind != "section" {
			continue
		}
		n, active := totals[category.Slug], view.Scope == category.Slug
		if n > 0 || active {
			view.Scopes = append(view.Scopes, searchScope{Name: category.Name, URL: searchURL(text, category.Slug, 1), Count: n, Active: active})
		}
	}
	if view.Scope == "" {
		return nil, labels
	}
	view.Total = totals[view.Scope]
	return members[view.Scope], labels
}

func searchLinkViews(links []models.Link, labels map[string]string) []linkView {
	views := toViews(links, "result")
	for i, link := range links {
		views[i].Category = labels[link.CategorySlug]
		parsed, err := url.Parse(link.URL)
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != "" {
			views[i].Domain = parsed.Hostname()
		}
	}
	return views
}

func orderLinkMatches(query *gorm.DB, text string) *gorm.DB {
	compact := strings.Join(strings.Fields(text), "")
	phrase, joined := containsSearchTerm(text), containsSearchTerm(compact)
	return query.Order(clause.OrderBy{Expression: clause.Expr{
		SQL: "CASE WHEN lower(name) IN (lower(?), lower(?)) OR lower(alias) IN (lower(?), lower(?)) THEN 0 " +
			"WHEN name ILIKE ? ESCAPE '!' OR name ILIKE ? ESCAPE '!' THEN 1 " +
			"WHEN \"desc\" ILIKE ? ESCAPE '!' OR \"desc\" ILIKE ? ESCAPE '!' THEN 2 " +
			"ELSE 3 END, clicks DESC, id ASC",
		Vars: []any{text, compact, text, compact, phrase, joined, phrase, joined},
	}})
}
