package main

import (
	"net/url"
	"log"
	"time"
)

type Bookmarks []Bookmark

type WebBookmark struct {
	BId int
	Username string
	URL string
	Title string
	Unread bool
	Archived bool
	AddedOn string
}

type Bookmark struct {
	BId int
	Username string
	URL string
	Title string
	Unread bool
	Archived bool
	AddedOn string
}

/* func (u *UserProfile) AsWebEntity() (wu WebUserProfile) {
	wu.Username = u.Username
	wu.DisplayName = u.DisplayName
	wu.JoinedOn = u.JoinedOn
	return
} */

func IsURL(str string) bool {
	u, err := url.Parse(str)

	if u.Host == "" {
		return false }

	if err != nil ||
		u.Host == "" ||
		!(u.Scheme == "http" || u.Scheme == "https") {
		return false }
	return true
}

func FormatDBDate(d string) (string) {
	t, _ := time.Parse(Settings.Database.DatetimeFormat, d)
	log.Println(d)
	log.Println(Settings.Database.DatetimeFormat)
	return t.Format(Settings.Web.DateFormat)
}

func (marks Bookmarks) AsWebEntities() (wb []WebBookmark) {
	for _, b := range marks {
		wb = append(wb, b.AsWebEntity())
	}
	return
}

func (b *Bookmark) AsWebEntity() (wb WebBookmark) {
	wb.BId = b.BId
	wb.Username = b.Username
	wb.URL = b.URL
	wb.Title = b.Title
	wb.Unread = b.Unread
	wb.Archived = b.Archived
	wb.AddedOn = FormatDBDate(b.AddedOn)
	return
}

