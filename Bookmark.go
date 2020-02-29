package main

import (
	"net/url"
	"strconv"
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

type URLError struct {
	ParseError bool
	BadScheme bool
	NoHost bool
}

func (e *URLError) Error() string {
	if e.ParseError { return "URL must start with http:// or https://" }
	if e.NoHost { return "No host was specified" }
	return "Not a URL"
}

/* func (u *UserProfile) AsWebEntity() (wu WebUserProfile) {
	wu.Username = u.Username
	wu.DisplayName = u.DisplayName
	wu.JoinedOn = u.JoinedOn
	return
} */

func IsURL(str string) error {
	u, err := url.Parse(str)

	if !(u.Scheme == "http" || u.Scheme == "https") {
		return &URLError{ BadScheme: true } }
	if u.Host == "" { return &URLError{ NoHost: true } }

	return err
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
	wb.URL = Settings.Web.Canon + "out/" + strconv.Itoa(b.BId)
	wb.Title = b.Title
	wb.Unread = b.Unread
	wb.Archived = b.Archived
	wb.AddedOn = WebDate(t)
	wb.AddedOnRFC3339 = RFC3339Date(t)
	return
}

