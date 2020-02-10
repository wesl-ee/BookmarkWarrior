package main

import (
	"net/url"
)

type BookmarkWeb struct {
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
