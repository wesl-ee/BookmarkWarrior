package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// Uplink to the Scrin mothership
func DBConnect(c *Config) (db *sql.DB, err error) {
	db, _ = sql.Open("mysql", c.Database.ConnectionString)
	err = db.Ping()
	return
}

func UserByName(db *sql.DB, uname string) (u UserProfile, err error) {
	selForm, err := db.Prepare(`SELECT
		Username, DisplayName, JoinedOn, Shadow, APISecret
		FROM Users WHERE Username=?`)
	if err != nil { return }
	err = selForm.QueryRow(uname).Scan(
		&u.Username,
		&u.DisplayName,
		&u.JoinedOn,
		&u.Shadow,
		&u.APISecret)

	/* if err == sql.ErrNoRows {
		return
	} else {
		return
	} */

	return
}

func (b Bookmark) Add(db *sql.DB) (error) {
	q := `INSERT INTO Bookmarks
		(Username, Title, URL) VALUES (?, ?, ?)`
	insForm, err := db.Prepare(q)
	if err != nil { return err }

	_, err = insForm.Exec(b.Username, b.Title, b.URL)
	return err
}

func (b Bookmark) Edit(db *sql.DB) (error) {
	q := `UPDATE Bookmarks
		SET Title=?, URL=? WHERE BId=? AND Username=?`
	insForm, err := db.Prepare(q)
	if err != nil { return err }

	_, err = insForm.Exec(b.Title, b.URL, b.BId, b.Username)
	return err
}

func (b Bookmark) Archive(db *sql.DB) (error) {
	q := `UPDATE Bookmarks SET Archived=true
		WHERE BId=? AND Username=?`
	upForm, err := db.Prepare(q)
	if err != nil { return err }

	_, err = upForm.Exec(b.BId, b.Username)
	return err
}

func (b Bookmark) Unarchive(db *sql.DB) (error) {
	q := `UPDATE Bookmarks SET Archived=false
		WHERE BId=? AND Username=?`
	upForm, err := db.Prepare(q)
	if err != nil { return err }

	_, err = upForm.Exec(b.BId, b.Username)
	return err
}

func (b Bookmark) Del(db *sql.DB) (error) {
	q := `DELETE FROM Bookmarks
		WHERE BId=?`
	delForm, err := db.Prepare(q)
	if err != nil { return err }

	_, err = delForm.Exec(b.BId)
	return err
}

func (ws WebSession) Associated(db *sql.DB) (s Session, err error) {
	selForm, err := db.Prepare(`SELECT
		SessID, Username, Expires
		FROM Sessions WHERE SessID=?`)
	if err != nil { return }
	err = selForm.QueryRow(ws.SessID).Scan(
		&s.SessID,
		&s.Username,
		&s.Expires)

	return
}

func (ws WebSession) Associate(db *sql.DB, uname string) (error) {
	q := `INSERT INTO Sessions
		(SessID, Username) VALUES (?, ?)`
	insForm, err := db.Prepare(q)
	if err != nil { return err }
	_, err = insForm.Exec(ws.SessID, uname)
	return err
}

func (ws WebSession) Disassociate(db *sql.DB) (error) {
	q := `DELETE FROM Sessions
		WHERE SessID=?`
	delForm, err := db.Prepare(q)
	if err != nil { return err }
	_, err = delForm.Exec(ws.SessID)
	return err
}

func (b Bookmark) MarkRead(db *sql.DB) (error) {
	q := `UPDATE Bookmarks SET Unread=False
		WHERE BId=?`
	upForm, err := db.Prepare(q)
	if err != nil { return err }
	_, err = upForm.Exec(b.BId)
	return err
}

func (b Bookmark) MarkUnread(db *sql.DB) (error) {
	q := `UPDATE Bookmarks SET Unread=true
		WHERE BId=?`
	upForm, err := db.Prepare(q)
	if err != nil { return err }
	_, err = upForm.Exec(b.BId)
	return err
}

func (u UserProfile) ArchivedBookmarks(db *sql.DB) (map[int]Bookmark, error) {
	marks := make(map[int]Bookmark)
	q := `SELECT
		BId, Username, URL, Title, Unread, Archived, AddedOn
		FROM Bookmarks WHERE Username=? AND Archived`
	selForm, err := db.Prepare(q)
	if err != nil { return marks, err }
	var m Bookmark
	rows, err := selForm.Query(u.Username)
	for rows.Next() { rows.Scan(
		&m.BId,
		&m.Username,
		&m.URL,
		&m.Title,
		&m.Unread,
		&m.Archived,
		&m.AddedOn)
		marks[m.BId] = m
	}
	return marks,err
}

func (u UserProfile) UnarchivedBookmarks(db *sql.DB) (map[int]Bookmark, error) {
	marks := make(map[int]Bookmark)
	q := `SELECT
		BId, Username, URL, Title, Unread, Archived, AddedOn
		FROM Bookmarks WHERE Username=? AND !Archived`
	selForm, err := db.Prepare(q)
	if err != nil { return marks, err }
	var m Bookmark
	rows, err := selForm.Query(u.Username)
	for rows.Next() { rows.Scan(
		&m.BId,
		&m.Username,
		&m.URL,
		&m.Title,
		&m.Unread,
		&m.Archived,
		&m.AddedOn)
		marks[m.BId] = m
	}
	return marks,err
}

func (u UserProfile) Bookmarks(db *sql.DB) (map[int]Bookmark, error) {
	marks := make(map[int]Bookmark)
	q := `SELECT
		BId, Username, URL, Title, Unread, Archived, AddedOn
		FROM Bookmarks WHERE Username=?`
	selForm, err := db.Prepare(q)
	if err != nil { return marks, err }
	var m Bookmark
	rows, err := selForm.Query(u.Username)
	for rows.Next() { rows.Scan(
		&m.BId,
		&m.Username,
		&m.URL,
		&m.Title,
		&m.Unread,
		&m.Archived,
		&m.AddedOn)
		marks[m.BId] = m
	}
	return marks,err
}

func (u UserProfile) Create(db *sql.DB, pass string) (UserProfile, error) {
	shadow := DoShadow(pass)
	apisecret := APISecret()

	q := `INSERT INTO Users
		(Username, DisplayName, Shadow, APISecret) VALUES
		(?, ?, ?, ?)`

	insForm, err := db.Prepare(q)
	if err != nil { return UserProfile{}, err }
	_, err = insForm.Exec(
		u.Username,
		u.DisplayName,
		shadow,
		apisecret)
	if err != nil { return UserProfile{}, err }

	return UserByName(db, u.Username)
}

func LetMeIn(db *sql.DB, uname, pass string) (UserProfile, error) {
	u, err := UserByName(db, uname)
	if err != nil { return u, err }

	fail := CompareShadow(u.Shadow, pass)
	if fail != nil { return u, err }

	return u, nil
}
