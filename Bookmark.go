package main

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
