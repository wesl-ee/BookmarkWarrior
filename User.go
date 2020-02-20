package main

import (
	"golang.org/x/crypto/bcrypt"
	"crypto/rand"
	"encoding/hex"
)

type WebUserProfile struct {
	Username string
	DisplayName string
	JoinedOn string
	JoinedOnRFC3339 string
	Bookmarks []WebBookmark
	Homepage string
	ThisIsMe bool
}

type UserProfile struct {
	Username string
	DisplayName string
	JoinedOn string
	Shadow string
	APISecret string
}

func (u *UserProfile) AsWebEntity() (wu WebUserProfile) {
	t, _ := ParseDBDate(u.JoinedOn)

	wu.Username = u.Username
	wu.DisplayName = u.DisplayName
	wu.JoinedOn = WebDate(t)
	wu.JoinedOnRFC3339 = RFC3339Date(t)
	wu.Homepage = Settings.Web.Canon + "u/" + u.Username
	return
}

func DoShadow(password string) (string) {
	hash, _ := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost)
	return string(hash)
}

func CompareShadow(shadow, password string) (error) {
	fail := bcrypt.CompareHashAndPassword([]byte(shadow), []byte(password))
	return fail
}

func APISecret() (string) {
	length := 15
	bytes := make([]byte, length)
	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

func ValidUsername(uname string) (bool) {
	for _, r := range uname {
		if ((r < 'a' || r > 'z') &&
			(r < '0' || r > '9') &&
			(r != '-')) {
			return false
		}
	}
	return len(uname) <= Settings.MaxUsernameLength
}

func ValidDisplayName(dname string) bool {
	return len(dname) <= Settings.MaxDisplaynameLength
}

func ValidPassword(pass string) (bool) {
	return len(pass) >= Settings.MinimumPasswordLength
}
