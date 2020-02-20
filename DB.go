package main

import (
	"database/sql"
	"time"
	"errors"
	_ "github.com/go-sql-driver/mysql"
)

var GlobalDB *sql.DB

type BOrder struct {
	Parameter string
	Order string
}

const (
	OrderAscending = "ASC"
	OrderDescending = "DESC"
	SortByAdded = "AddedOn"
	SortByTitle = "Title"
)

// Uplink to the Scrin mothership
func DBConnect(c *Config) (*sql.DB, error) {
	if GlobalDB != nil { return GlobalDB, nil }
	GlobalDB, _ = sql.Open("mysql", c.Database.ConnectionString)
	return GlobalDB, GlobalDB.Ping()
}

func PromoDiscount(db *sql.DB, promo string) (discount float64, err error) {
	selForm, err := db.Prepare(`SELECT Discount
		FROM Promos WHERE Code=? AND
		Expires >= CURRENT_TIMESTAMP`)
	if err != nil { return }
	err = selForm.QueryRow(promo).Scan(&discount)
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

func (u UserProfile) ChangeDisplayName(db *sql.DB, newname string) error {
	q := `UPDATE Users SET DisplayName=? WHERE Username=?`
	upForm, err := db.Prepare(q)
	if err != nil { return err }
	_, err = upForm.Exec(newname, u.Username)
	return err
}

func (u UserProfile) DeleteSessions(db *sql.DB) (error) {
	q := `DELETE FROM Sessions WHERE Username=?`
	delForm, err := db.Prepare(q)
	if err != nil { return err }
	_, err = delForm.Exec(u.Username)
	return err
}

func (u UserProfile) Derez(db *sql.DB) (error) {
	// Clear sessions first...
	err := u.DeleteSessions(db)
	if err != nil { return err }

	// ...then actually derez the user
	q := `DELETE FROM Users WHERE Username=?`
	delForm, err := db.Prepare(q)
	if err != nil { return err }
	_, err = delForm.Exec(u.Username)
	return err
}

func (u UserProfile) ArchivedBookmarks(db *sql.DB, order *BOrder) (Bookmarks, error) {
	var marks []Bookmark
	q := `SELECT
		BId, Username, URL, Title, Unread, Archived, AddedOn
		FROM Bookmarks WHERE Username=? AND Archived ORDER BY ` +
		order.Parameter + " " + order.Order
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
		marks = append(marks, m)
	}
	return marks,err
}

func (u UserProfile) UnarchivedBookmarks(db *sql.DB, order *BOrder) (Bookmarks, error) {
	var marks []Bookmark
	q := `SELECT
		BId, Username, URL, Title, Unread, Archived, AddedOn
		FROM Bookmarks WHERE Username=? AND !Archived ORDER BY ` +
		order.Parameter + " " + order.Order
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
		marks = append(marks, m)
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
	if err != nil { return u, errors.New("Login Error") }

	fail := CompareShadow(u.Shadow, pass)
	if fail != nil { return u, errors.New("Login Error") }

	return u, nil
}

func FormatDBDate(d string) (string) {
	t, _ := time.Parse(Settings.Database.DatetimeFormat, d)
	return t.Format(Settings.Web.DateFormat)
}

func ParseDBDate(d string) (time.Time, error) {
	return time.Parse(Settings.Database.DatetimeFormat, d)
}

func WebDate(t time.Time) (string) { return t.Format(Settings.Web.DateFormat) }
func RFC3339Date(t time.Time) (string) { return t.Format(time.RFC3339) }

