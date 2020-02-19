package main

import (
	"net/url"
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
	AddedOnRFC3339 string
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

func (marks Bookmarks) AsWebEntities() (wb []WebBookmark) {
	for _, b := range marks {
		wb = append(wb, b.AsWebEntity())
	}
	return
}

func (b *Bookmark) AsWebEntity() (wb WebBookmark) {
	t, _ := ParseDBDate(b.AddedOn)

	wb.BId = b.BId
	wb.Username = b.Username
	wb.URL = b.URL
	wb.Title = b.Title
	wb.Unread = b.Unread
	wb.Archived = b.Archived
	wb.AddedOn = WebDate(t)
	wb.AddedOnRFC3339 = RFC3339Date(t)
	return
}

