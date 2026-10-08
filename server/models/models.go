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

type Link struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Slug         string    `gorm:"size:80;index" json:"slug"`
	CategorySlug string    `gorm:"size:64;index" json:"categorySlug"`
	Name         string    `gorm:"size:128" json:"name"`
	URL          string    `gorm:"size:512" json:"url"`
	Desc         string    `gorm:"size:256" json:"desc"`
	Tags         string    `gorm:"size:160" json:"tags"`
	Alias        string    `gorm:"size:64" json:"alias"`
	Lang         string    `gorm:"size:32" json:"lang"`
	Region       string    `gorm:"size:32" json:"region"`
	Spare        string    `gorm:"size:512" json:"spare"`
	ICP          string    `gorm:"column:icp;size:64" json:"icp"`
	Body         string    `gorm:"type:text" json:"body"`
	Pinned       bool      `json:"pinned"`
	PinSort      int       `json:"pinSort"`
	Sort         int       `json:"sort"`
	Clicks       int       `json:"clicks"`
	Views        int       `json:"views"`
	Status       string    `gorm:"size:16;index" json:"status"`
	UserID       uint      `gorm:"index" json:"userId"`
	CreatedAt    time.Time `json:"createdAt"`
}

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type UserLink struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:ux_user_link_kind" json:"userId"`
	LinkID    uint      `gorm:"uniqueIndex:ux_user_link_kind" json:"linkId"`
	Kind      string    `gorm:"size:16;uniqueIndex:ux_user_link_kind" json:"kind"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Article struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:200" json:"title"`
	Summary   string    `gorm:"size:400" json:"summary"`
	Body      string    `gorm:"type:text" json:"body"`
	Views     int       `json:"views"`
	CreatedAt time.Time `json:"createdAt"`
}

type Tag struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Group string `gorm:"size:32;index" json:"group"`
	Name  string `gorm:"size:64" json:"name"`
	Sort  int    `json:"sort"`
}

type Admin struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}
