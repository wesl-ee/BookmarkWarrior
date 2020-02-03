package main

import (
	"net/http"
	"strings"
	"crypto/md5"
	"encoding/base64"
	"time"
)

type Session struct {
	SessID string
	Username string
	Expires string }

type WebSession struct {
	SessID string
}

/* func LoadWebSession(db *sql.DB, r *http.Request) (ws WebSession, err error) {
	id, err := r.Cookie(Settings.Web.SessionCookie)
	if err != nil { return }
	sess, err := SessById(db, id)

	ws.SessID = s.SessID
} */

func InitWebSession(w http.ResponseWriter, r *http.Request) {
	sesscookie := http.Cookie{
		Name: Settings.Web.SessionCookie,
		Value: DeriveSessID(r) }

	// Send the HTTP header...
	http.SetCookie(w, &sesscookie)

	// ...and set cookie in the current request to avoid refreshes
	r.AddCookie(&sesscookie)
}

func HasWebSession(r *http.Request) (bool) {
	_, err := r.Cookie(Settings.Web.SessionCookie)
	return err == nil
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
