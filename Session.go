package main

import (
	"database/sql"
	"net/http"
	"strings"
	"crypto/md5"
	"encoding/base64"
	"time"
)

type UserExperience struct {
	SessID string
	Username string
	LoggedIn bool
	Theme string }

type Session struct {
	SessID string
	Username string
	Expires string }

type WebSession struct {
	SessID string }

/* func LoadWebSession(db *sql.DB, r *http.Request) (ws WebSession, err error) {
	id, err := r.Cookie(Settings.Web.SessionCookie)
	if err != nil { return }
	sess, err := SessById(db, id)

	ws.SessID = s.SessID
} */

func (UX *UserExperience) LoadSession(s Session) {
	UX.SessID = s.SessID
	UX.Username = s.Username
	UX.LoggedIn = true
}

func (UX *UserExperience) LoadGeneric(ws WebSession) {
	UX.SessID = ws.SessID
	UX.LoggedIn = false
}

func LoadUX(db *sql.DB, r *http.Request) (*UserExperience) {
	UX := &UserExperience{}
	ws := ThisSession(r)
	s, err := ws.Associated(db)
	if err != nil { UX.LoadGeneric(ws)
	} else { UX.LoadSession(s) }
	return UX
}

func InitWebSession(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	afterHours := time.Duration(24 * Settings.Web.SessionExpiryDays)
	sesscookie := http.Cookie{
		Name: Settings.Web.SessionCookie,
		Expires: now.Add(time.Hour * afterHours),
		Value: DeriveSessID(r) }

	// Send the HTTP header...
	http.SetCookie(w, &sesscookie)

	// ...and set cookie in the current request to avoid refreshes
	r.AddCookie(&sesscookie)
}

func (ws WebSession) ForgetMe(w http.ResponseWriter, r *http.Request) {
	sesscookie := http.Cookie{
		Name: Settings.Web.SessionCookie,
		Value: "",
		Expires: time.Now() }
	http.SetCookie(w, &sesscookie)
	r.AddCookie(&sesscookie)
}

func HasWebSession(r *http.Request) (bool) {
	_, err := r.Cookie(Settings.Web.SessionCookie)
	return err == nil
}

func ThisSession(r *http.Request) (WebSession) {
	cookie, err := r.Cookie(Settings.Web.SessionCookie)
	if err != nil { panic(err) }
	return WebSession{ SessID: cookie.Value }
}

func DeriveSessID(r *http.Request) (string) {
	var sb strings.Builder
	components := []string{
		RealIP(r),
		r.UserAgent(),
		string(time.Now().Unix()) }

	for _, c := range components {
		sb.WriteString(c)
	}

	hash := md5.Sum([]byte(sb.String()))
	return base64.StdEncoding.EncodeToString(hash[:])[:22]
}

func RealIP(r *http.Request) string {
	if realip := r.Header.Get("x-real-ip"); realip != "" {
		return realip
	}
	return r.RemoteAddr
}
