package models

import "time"

type Category struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Slug        string `gorm:"uniqueIndex;size:64" json:"slug"`
	Name        string `gorm:"size:64" json:"name"`
	ParentSlug  string `gorm:"size:64;index" json:"parentSlug"`
	Kind        string `gorm:"size:16" json:"kind"`
	ContentKind string `gorm:"size:16" json:"contentKind"`
	Sort        int    `json:"sort"`
	ShowOnHome  bool   `json:"showOnHome"`
}

type CategoryGuide struct {
	CategorySlug string `gorm:"primaryKey;size:64"`
	Title        string `gorm:"size:128"`
	Intro        string `gorm:"type:text"`
	HomeLabel    string `gorm:"size:64"`
}

type HistoryEvent struct {
	ID    uint   `gorm:"primaryKey"`
	Month int    `gorm:"not null;uniqueIndex:ux_history_event"`
	Day   int    `gorm:"not null;uniqueIndex:ux_history_event"`
	Year  int    `gorm:"not null;uniqueIndex:ux_history_event"`
	Text  string `gorm:"size:255;not null;uniqueIndex:ux_history_event"`
}

type Link struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ImportedFrom *string    `gorm:"uniqueIndex;size:96" json:"-"`
	Slug         string     `gorm:"size:80;index" json:"slug"`
	CategorySlug string     `gorm:"size:64;index" json:"categorySlug"`
	Name         string     `gorm:"size:128" json:"name"`
	URL          string     `gorm:"size:512" json:"url"`
	Desc         string     `gorm:"size:256" json:"desc"`
	Tags         string     `gorm:"size:160" json:"tags"`
	Alias        string     `gorm:"size:64" json:"alias"`
	Lang         string     `gorm:"size:32" json:"lang"`
	Region       string     `gorm:"size:32" json:"region"`
	Spare        string     `gorm:"size:512" json:"spare"`
	ICP          string     `gorm:"column:icp;size:64" json:"icp"`
	Body         string     `gorm:"type:text" json:"body"`
	Pinned       bool       `json:"pinned"`
	PinSort      int        `json:"pinSort"`
	Sort         int        `json:"sort"`
	Clicks       int        `json:"clicks"`
	Views        int        `json:"views"`
	CheckedAt    *time.Time `json:"checkedAt"`
	CheckFails   int        `json:"checkFails"`
	Status       string     `gorm:"size:16;index" json:"status"`
	UserID       uint       `gorm:"index" json:"userId"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type InternetRankMonth struct {
	Month     string    `gorm:"size:7;primaryKey" json:"month"`
	Source    string    `gorm:"size:128;not null" json:"source"`
	SourceURL string    `gorm:"size:512;not null" json:"sourceUrl"`
	Entries   int       `json:"entries"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type InternetRankEntry struct {
	Month    string `gorm:"size:7;primaryKey"`
	Domain   string `gorm:"size:253;primaryKey"`
	Category string `gorm:"size:128;not null"`
	Name     string `gorm:"size:128;not null"`
	Visits   uint64 `gorm:"not null"`
	Risky    bool   `gorm:"not null;default:false"`
}

type InternetRankBlock struct {
	Domain string `gorm:"size:253;primaryKey" json:"domain"`
}

type SubmissionReview struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	LinkID       uint      `gorm:"index" json:"linkId"`
	Name         string    `gorm:"size:128" json:"name"`
	URL          string    `gorm:"size:512" json:"url"`
	Decision     string    `gorm:"size:16" json:"decision"`
	CategorySlug string    `gorm:"size:64" json:"categorySlug"`
	Reviewer     string    `gorm:"size:32" json:"reviewer"`
	Note         string    `gorm:"size:300" json:"note"`
	CreatedAt    time.Time `json:"createdAt"`
}

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	Email        string    `gorm:"size:254" json:"-"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	IsAdmin      bool      `gorm:"not null;default:false" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type PasswordReset struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	TokenHash string    `gorm:"size:64;uniqueIndex"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

type UserLink struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:ux_user_link_kind" json:"userId"`
	LinkID    uint      `gorm:"uniqueIndex:ux_user_link_kind" json:"linkId"`
	Kind      string    `gorm:"size:16;uniqueIndex:ux_user_link_kind" json:"kind"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Article struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	LegacyID   *string   `gorm:"uniqueIndex;size:64" json:"-"`
	LegacySlug string    `gorm:"index;size:255" json:"-"`
	Title      string    `gorm:"size:255" json:"title"`
	Summary    string    `gorm:"type:text" json:"summary"`
	Body       string    `gorm:"type:longtext" json:"body"`
	Format     string    `gorm:"size:16" json:"format"`
	CoverURL   string    `gorm:"size:512" json:"coverUrl"`
	VideoURL   string    `gorm:"size:512" json:"videoUrl"`
	Category   string    `gorm:"size:64;index" json:"category"`
	Featured   bool      `json:"featured"`
	Sort       int       `json:"sort"`
	Draft      bool      `gorm:"not null;default:false;index" json:"draft"`
	Views      int       `json:"views"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ArticleComment struct {
	ID        uint   `gorm:"primaryKey"`
	LegacyID  *uint  `gorm:"uniqueIndex"`
	ArticleID uint   `gorm:"index;not null"`
	ParentID  *uint  `gorm:"index"`
	UserID    uint   `gorm:"index"`
	Author    string `gorm:"size:128"`
	Body      string `gorm:"type:text"`
	Hidden    bool   `gorm:"not null;default:false"`
	CreatedAt time.Time
}

type Feedback struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	LegacyID  *uint     `gorm:"uniqueIndex" json:"-"`
	UserID    uint      `gorm:"index" json:"userId"`
	Kind      string    `gorm:"size:16" json:"kind"`
	Body      string    `gorm:"type:text" json:"body"`
	Status    string    `gorm:"size:16;index" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type Tag struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Group string `gorm:"size:32;index" json:"group"`
	Name  string `gorm:"size:64" json:"name"`
	Sort  int    `json:"sort"`
}
