package main

import (
	"log"
	"net/http"
	"os"
	"html/template"
	"strconv"
	"net/url"
	"fmt"
	"strings"
	"database/sql"
	_ "golang.org/x/crypto/bcrypt"
)

var Settings Config
var Templates = map[string]*template.Template{}

func PageDependencies(page string) ([]string) {
	for _, p := range Settings.Templates {
		if p.Name == page { return p.Dependencies }
	}
	return nil
}

type IndexPage struct {
	UserGraph *Bargraph
	BookmarkGraph *Bargraph
	UX *UserExperience
	Settings *Config }

type UserEditPage struct {
	Canon string
	User WebUserProfile
	Title string
	Mark Bookmark
	UX *UserExperience
	Settings *Config }

type UserAddPage struct {
	Canon string
	Title string
	User WebUserProfile
	UX *UserExperience
	Settings *Config }

type LoginPage struct {
	Error *LoginError
	Settings *Config
	UX *UserExperience }

type UserPage struct {
	Canon string
	Settings *Config
	User WebUserProfile
	UX *UserExperience
	Title string }

type ArchivePage struct {
	Canon string
	Settings *Config
	User WebUserProfile
	UX *UserExperience
	Title string }

type InfoPage struct {
	UX *UserExperience
	Settings *Config }

type UserSettingsPage struct {
	Canon string
	Error *SignupError
	Title string
	User WebUserProfile
	UX *UserExperience
	Settings *Config }

type SignupError struct {
	Mismatch bool
	Taken bool
	BadUName bool
	BadDispName bool
	ShortPassword bool
	BadPassword bool
	AlreadyLoggedIn bool }

type LoginError struct {
	DBError bool
	CredsError bool }

type SignupNewPage struct {
	Settings *Config
	UX *UserExperience
	Error *SignupError}

type SignupCreatePage struct {
	Username string
	DisplayName string
	Password string
	Promo string
	Cost float64
	UX *UserExperience
	Settings *Config }

type SignupFreePage struct {
	Username string
	DisplayName string
	Password string
	Promo string
	UX *UserExperience
	Settings *Config }

type SignupReceiptPage struct {
	UX *UserExperience
	Settings *Config
}

type ServerRes struct {
	DB *sql.DB
	Writer http.ResponseWriter
	Request *http.Request
}

func InitTemplates() {
	for _, this := range Settings.Templates {
		file := this.Name
		incl := this.Dependencies
		tmpl := template.Must(template.ParseFiles(file))
		for _, t := range incl {
			template.Must(tmpl.ParseFiles(t)) }
		Templates[file] = tmpl
	}
}

func (ux *UserExperience) HandleWebIndex(res *ServerRes) {
	w := res.Writer
	r := res.Request
	page := "tmpl/index.html"

	usage, _ := SiteUsage(res.DB)
	b := usage.Users.AsBarGraph()
	c := usage.Bookmarks.AsBarGraph()

	tmpl := Templates[page]
	err := tmpl.Execute(w, IndexPage{
		UserGraph: &b,
		BookmarkGraph: &c,
		Settings: &Settings,
		UX: ux})
	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
}

func QueryAsOrder(q string) (*BOrder) {
	switch(q) {
		case "ascending-name": return &BOrder{
			Parameter: SortByTitle,
			Order: OrderAscending }
		case "descending-name": return &BOrder{
			Parameter: SortByTitle,
			Order: OrderDescending }
		case "ascending-date": return &BOrder{
			Parameter: SortByAdded,
			Order: OrderAscending }
		case "descending-date": return &BOrder{
			Parameter: SortByAdded,
			Order: OrderDescending }
		default: return nil }
}

func (ux *UserExperience) HandleUserReq(res *ServerRes, uname string) {
	w := res.Writer
	r := res.Request
	db := res.DB
	page := "tmpl/user.html"

	user, err := UserByName(db, uname)
	if err != nil {
		// User not found...
		HandleWebError(w, r, http.StatusNotFound)
		log.Println(err)
		return }

	var order *BOrder
	param, ok := res.Request.URL.Query()["order"]
	if ok && len(param[0]) > 0 { order = QueryAsOrder(param[0]) }
	if order == nil { order = &BOrder{
		Parameter: SortByAdded,
		Order: OrderDescending } }

	marks, err := user.UnarchivedBookmarks(db, order)
	if err != nil {
		// Databse error...
		HandleWebError(w, r, http.StatusServiceUnavailable)
		log.Println(err)
		return
	}

	webuser:= user.AsWebEntity()
	webuser.Bookmarks = marks.AsWebEntities()
	webuser.ThisIsMe = ux.Username == uname

	tmpl := Templates[page]
	err = tmpl.Execute(w, UserPage{
		Settings: &Settings,
		Canon: Settings.Web.Canon + "u/" + uname,
		User: webuser,
		UX: ux,
		Title: user.DisplayName + " (" + uname + ") - Bookmarks" })

	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
		log.Println(err)
	}
}

func (ux *UserExperience) HandleLogout(res *ServerRes) {
	ws := ThisSession(res.Request)
	ws.Disassociate(res.DB)

	http.Redirect(res.Writer, res.Request, "/", http.StatusSeeOther)
}

func (ux *UserExperience) HandleUserSettings(res *ServerRes, uname, option string) {
	user, err := UserByName(res.DB, uname)
	if err != nil {
		HandleWebError(res.Writer, res.Request, http.StatusNotFound)
		return }

	if !ux.LoggedIn {
		http.Redirect(res.Writer, res.Request, "/login", http.StatusSeeOther)
		log.Println(err)
		return
	}

	if ux.Username != uname {
		HandleWebError(res.Writer, res.Request, http.StatusForbidden)
		log.Println(err)
		return }

	var procErr *SignupError
	if (res.Request.Method == "POST") {
		if err := res.Request.ParseForm(); err != nil {
			HandleWebError(res.Writer, res.Request,
				http.StatusInternalServerError)
			log.Println(err)
			return
		}
		switch(option) {
		case "derez":
			derez := res.Request.FormValue("derez")
			password := res.Request.FormValue("password")

			if derez == "" {
				return // TODO: P-R-G
			}
			u, err := LetMeIn(res.DB, uname, password)
			if err != nil {
				return // TODO: P-R-G
			}

			// Actually delete the account
			// ...
			log.Printf("User %s (@%s) has opted to delete their account!",
				u.DisplayName, u.Username)
			err = u.Derez(res.DB)
			if err != nil {
				log.Printf("Failed to delete user %s (@%s): %s",
					u.DisplayName, u.Username, err)
			} else {
				log.Printf("Successfully derezzed %s (@%s)",
					u.DisplayName, u.Username)
			}
			http.Redirect(res.Writer, res.Request,
				Settings.Web.Canon, http.StatusSeeOther)
			return
		case "change-name":
			newname := res.Request.FormValue("newname")

			u, err := UserByName(res.DB, uname)
			if err != nil {
				log.Println(err)
				HandleWebError(res.Writer, res.Request,
					http.StatusInternalServerError)
				return
			}

			if newname == u.DisplayName {
				http.Redirect(res.Writer, res.Request,
					Settings.Web.Canon + "/u/" + uname, http.StatusSeeOther)
				return
			}

			if !ValidDisplayName(newname) {
				procErr = &SignupError{ BadDispName: true }
			} else {
				// ...
				err = u.ChangeDisplayName(res.DB, newname)
				if err != nil {
					log.Println(err)
					HandleWebError(res.Writer, res.Request,
						http.StatusInternalServerError)
					return
				}

				log.Printf("User %s (@%s) changed name -> %s\n",
					u.DisplayName, u.Username, newname)
				http.Redirect(res.Writer, res.Request,
					Settings.Web.Canon + "/u/" + uname, http.StatusSeeOther)
				}
		case "change-password":
			currpassword := res.Request.FormValue("currpassword")
			newpassword := res.Request.FormValue("newpassword")
			confirmpassword := res.Request.FormValue("confirmpassword")

			if newpassword != confirmpassword {
				procErr = &SignupError{ Mismatch: true }
				break
			}

			u, err := LetMeIn(res.DB, uname, currpassword)
			if err != nil {
				procErr = &SignupError{ BadPassword: true }
				break
			}

			err = u.NewPassword(res.DB, newpassword)
			if err != nil {
				log.Println(err)
				HandleWebError(res.Writer, res.Request,
					http.StatusInternalServerError)
				return
			}

			log.Printf("User %s (@%s) changed their password!",
				u.DisplayName, uname)
			http.Redirect(res.Writer, res.Request,
				Settings.Web.Canon + "/u/" + uname, http.StatusSeeOther)
		}
	}

	var page string
	switch(option) {
	case "change-name":
		page = "tmpl/user-change-name.html"
	case "change-password":
		page = "tmpl/user-change-password.html"
	case "derez":
		page = "tmpl/user-derez.html"
	case "":
		page = "tmpl/user-settings.html"
	}

	tmpl := Templates[page]
	webuser:= user.AsWebEntity()
	webuser.ThisIsMe = ux.Username == uname

	err = tmpl.Execute(res.Writer, UserSettingsPage{
		Canon: Settings.Web.Canon + "u/" + uname,
		Error: procErr,
		User: webuser,
		Title: user.DisplayName + " (" + uname + ") - Settings",
		UX: ux,
		Settings: &Settings })
	if err != nil {
		HandleWebError(res.Writer, res.Request,
			http.StatusInternalServerError)
		log.Println(err)
	}
}

func (ux *UserExperience) HandleUserAdd(res *ServerRes, uname string) {
	user, err := UserByName(res.DB, uname)
	if err != nil {
		HandleWebError(res.Writer, res.Request, http.StatusNotFound)
		return }

	if !ux.LoggedIn {
		http.Redirect(res.Writer, res.Request, "/login", http.StatusSeeOther)
		log.Println(err)
		return
	}

	if ux.Username != uname {
		HandleWebError(res.Writer, res.Request, http.StatusForbidden)
		log.Println(err)
		return }

	if (res.Request.Method == "POST") {
		if err := res.Request.ParseForm(); err != nil { panic(err) }
		name := res.Request.FormValue("name")
		url := res.Request.FormValue("url")

		if !IsURL(url) {
			HandleWebError(res.Writer, res.Request,
				http.StatusBadRequest)
			return
		}

		b := Bookmark{
			Username: uname,
			Title: name,
			URL: url }
		err = b.Add(res.DB)
		if err != nil {
			HandleWebError(res.Writer, res.Request,
				http.StatusInternalServerError)
			log.Println(err)
			return
		}
		http.Redirect(res.Writer, res.Request, "/u/" + uname, http.StatusSeeOther)
		return
	}
	page := "tmpl/user-add.html"
	tmpl := Templates[page]
	webuser:= user.AsWebEntity()
	webuser.ThisIsMe = ux.Username == uname

	err = tmpl.Execute(res.Writer, UserAddPage{
		Canon: Settings.Web.Canon + "u/" + uname,
		User: webuser,
		Title: user.DisplayName + " (" + uname + ") - Add Bookmark",
		UX: ux,
		Settings: &Settings })
	if err != nil {
		HandleWebError(res.Writer, res.Request,
			http.StatusInternalServerError)
		log.Println(err)
	}
}

func (ux *UserExperience) HandleUserViewArchive(res *ServerRes, uname string) {
	page := "tmpl/user-archive.html"
	user, err := UserByName(res.DB, uname)
	if err != nil {
		// User not found...
		HandleWebError(res.Writer, res.Request,
			http.StatusNotFound)
		return }

	var order *BOrder
	param, ok := res.Request.URL.Query()["order"]
	if ok && len(param[0]) > 0 { order = QueryAsOrder(param[0]) }
	if order == nil { order = &BOrder{
		Parameter: SortByAdded,
		Order: OrderDescending } }

	marks, err := user.ArchivedBookmarks(res.DB, order)
	if err != nil {
		// Databse error...
		HandleWebError(res.Writer, res.Request,
			http.StatusServiceUnavailable)
		log.Println(err)
		return
	}

	webuser:= user.AsWebEntity()
	webuser.Bookmarks = marks.AsWebEntities()
	webuser.ThisIsMe = ux.Username == uname

	tmpl := Templates[page]
	err = tmpl.Execute(res.Writer, ArchivePage{
		Settings: &Settings,
		Canon: Settings.Web.Canon + "u/" + uname,
		User: webuser,
		UX: ux,
		Title: user.DisplayName + " (" + uname + ") - Archived Bookmarks" })

	if err != nil {
		HandleWebError(res.Writer, res.Request,
			http.StatusInternalServerError)
		log.Println(err)
	}
}

func (ux *UserExperience) HandleBMarkAction(res *ServerRes, uname string, bID int, action string) {
	if ux.Username != uname {
		HandleWebError(res.Writer, res.Request, http.StatusForbidden)
		return
	}

	u, err := UserByName(res.DB, uname)
	if err != nil { panic(err) }
	marks, err := u.Bookmarks(res.DB)
	if err != nil { panic(err) }
	mark, isValid := marks[bID]
	if !isValid {
		HandleWebError(res.Writer, res.Request,
			http.StatusNotFound)
		return
	}

	switch(action) {
		case "read":
			mark.MarkRead(res.DB)
		case "unread":
			mark.MarkUnread(res.DB)
		case "edit":
			ux.HandleUserEdit(res, mark)
			return
		case "unarchive":
			mark.Unarchive(res.DB)
		case "archive":
			mark.Archive(res.DB)
		case "remove":
			mark.Del(res.DB)
		default:
			HandleWebError(res.Writer, res.Request,
				http.StatusMethodNotAllowed)
			return
	}
	http.Redirect(res.Writer, res.Request, "/u/" + uname, http.StatusSeeOther)
}

func (ux *UserExperience) HandleUserEdit(res *ServerRes, mark Bookmark) {
	uname := mark.Username
	user, err := UserByName(res.DB, uname)
	if err != nil {
		HandleWebError(res.Writer, res.Request, http.StatusNotFound)
		return }

	if !ux.LoggedIn {
		http.Redirect(res.Writer, res.Request, "/login", http.StatusSeeOther)
		return
	}

	if ux.Username != uname {
		HandleWebError(res.Writer, res.Request, http.StatusForbidden)
		return }

	if (res.Request.Method == "POST") {
		if err := res.Request.ParseForm(); err != nil { panic(err) }
		name := res.Request.FormValue("name")
		url := res.Request.FormValue("url")

		mark.Title = name
		mark.URL = url
		err = mark.Edit(res.DB)

		if err != nil {
			HandleWebError(res.Writer, res.Request,
				http.StatusInternalServerError)
			return
		}
		http.Redirect(res.Writer, res.Request, "/u/" + uname, http.StatusSeeOther)
		return
	}
	page := "tmpl/user-edit.html"
	tmpl := Templates[page]
	webuser:= user.AsWebEntity()

	err = tmpl.Execute(res.Writer, UserEditPage{
		Canon: Settings.Web.Canon + "u/" + uname,
		User: webuser,
		Title: user.DisplayName + " (" + uname + ") - Edit Bookmark",
		UX: ux,
		Mark: mark,
		Settings: &Settings })
	if err != nil {
		HandleWebError(res.Writer, res.Request,
			http.StatusInternalServerError)
	}
}

func (ux *UserExperience) HandleLogin(res *ServerRes, e *LoginError) {
	w := res.Writer
	r := res.Request
	db := res.DB
	page := "tmpl/login.html"
	tmpl := Templates[page]

	// Handle login attempts
	if e == nil && r.Method == "POST" {
		if err := r.ParseForm(); err != nil { panic(err) }
		username := strings.ToLower(r.FormValue("username"))
		password := r.FormValue("password")

		u, err := LetMeIn(db, username, password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			ux.HandleLogin(res, &LoginError{CredsError: true})
			return
		}

		ws := ThisSession(r)
		ws.Associate(db, u.Username)

		http.Redirect(res.Writer, res.Request, "/u/" + username, http.StatusSeeOther)
		return
	}

	err := tmpl.Execute(w, LoginPage{
		Error: e,
		UX: ux,
		Settings: &Settings })
	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
}

func HandleStatic(res *ServerRes) {
	w := res.Writer
	r := res.Request
	filename := "static/" + strings.TrimPrefix(
		r.URL.Path, "/static/")
	stat, err := os.Stat(filename)
	if err != nil {
		HandleWebError(w, r, http.StatusNotFound)
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		HandleWebError(w, r, http.StatusNotFound)
		return
	}

	modifiedTime := stat.ModTime()
	http.ServeContent(w, r, filename, modifiedTime, file)
}

func (ux *UserExperience) HandleInfo(res *ServerRes, which string) {
	var page string
	switch(which) {
	case "about":
		page = "tmpl/about.html"
	case "privacy":
		page = "tmpl/privacy.html"
	case "mission":
		page = "tmpl/mission.html"
	case "technology":
		page = "tmpl/technology.html"
	}
	Templates[page].Execute(res.Writer, InfoPage{
		UX: ux,
		Settings: &Settings })
}

// 1st step in acc creation...
func (ux *UserExperience) HandleSignupNew(res *ServerRes, e *SignupError) {
	w := res.Writer
	r := res.Request
	page := "tmpl/signup-new.html"

	tmpl := Templates[page]
	err := tmpl.Execute(w, SignupNewPage{
		UX: ux,
		Error: e,
		Settings: &Settings })
	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
}

// 2nd step in acc creation...
func (ux *UserExperience) HandleSignupCreate(res *ServerRes) {
	w := res.Writer
	r := res.Request
	db := res.DB

	// Parse form submission
	if (r.Method != "POST") {
		http.Redirect(w, r, "/signup/new", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil { panic(err) }

	username := strings.ToLower(r.FormValue("username"))
	displayname := r.FormValue("displayname")
	password := r.FormValue("password")
	confirmpassword := r.FormValue("confirmpassword")
	promo := r.FormValue("promo")

	if displayname == "" { displayname = username }

	if !ValidPassword(password) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		ux.HandleSignupNew(res, &SignupError{ShortPassword: true})
		return
	}

	if !ValidUsername(username) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		ux.HandleSignupNew(res, &SignupError{BadUName: true})
		return
	}

	existing, err := UserByName(db, username)
	if err == nil && existing.Username == username {
		w.WriteHeader(http.StatusConflict)
		ux.HandleSignupNew(res, &SignupError{Taken: true})
		return
	}

	if confirmpassword != password {
		w.WriteHeader(http.StatusUnprocessableEntity)
		ux.HandleSignupNew(res, &SignupError{Mismatch: true})
		return
	}

	// Promotional discount
	cost := Settings.PayPal.OneTimeCost
	if promo != "" {
		discount, _ := PromoDiscount(db, promo)
		cost -= discount
	}

	if cost <= 0 {
		page := "tmpl/signup-free.html"
		tmpl := Templates[page]
		err = tmpl.Execute(w, SignupFreePage{
			Username: username,
			DisplayName: displayname,
			Password: password,
			Promo: promo,
			UX: ux,
			Settings: &Settings })
		if err != nil {
			HandleWebError(w, r, http.StatusInternalServerError)
		}
	} else {
		page := "tmpl/signup-create.html"
		tmpl := Templates[page]
		err = tmpl.Execute(w, SignupCreatePage{
			Username: username,
			DisplayName: displayname,
			Password: password,
			Cost: cost,
			Promo: promo,
			UX: ux,
			Settings: &Settings })
		if err != nil {
			HandleWebError(w, r, http.StatusInternalServerError)
		}
	}
}

// 3rd step in acc creation...
func (ux *UserExperience) HandleSignupPay(res *ServerRes) {
	var err error
	w := res.Writer
	r := res.Request
	db := res.DB

	// Parse form submission
	if (r.Method != "POST") {
		HandleWebError(w, r, http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil { panic(err) }

	username := r.FormValue("username")
	displayname := r.FormValue("displayname")
	password := r.FormValue("password")
	promo := r.FormValue("promo")
	orderID := r.FormValue("orderid")

	// Promotional discount
	cost := Settings.PayPal.OneTimeCost
	if promo != "" {
		discount, _ := PromoDiscount(db, promo)
		cost -= discount
	}

	if displayname == "" { displayname = username }

	paid := false
	if cost > 0 {
		// Poll PayPal for payment verification
		paid, err = VerifyPayment(orderID)
		if err != nil { panic(err) }
	} else { paid = true }

	if paid {
		// Actually create account in DB
		// ...
		newUser := UserProfile{
			Username: username,
			DisplayName: displayname, }

		u, err := newUser.Create(db, password)
		if err != nil {
			HandleWebError(w, r, http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// Log to console
		log.Printf("Created user %s (%s)\n", u.DisplayName, u.Username)

		// Log us in immediately after acc. creation
		ThisSession(r).Associate(db, u.Username)

		// ...P-R-G and to show receipt (minimize refresh errors)
		http.Redirect(w, r, "/signup/receipt", http.StatusFound);
		return
	}
	// Confirmation was received but was not paid (for whatever reason...)
	// (e.g. insufficient funds, transaction denied, etc. etc.)
	// ...
	// ...
}

// 4th step in acc creation...
func (ux *UserExperience) HandleSignupReceipt(res *ServerRes) {
	w := res.Writer
	r := res.Request
	page := "tmpl/signup-receipt.html"

	tmpl := Templates[page]
	err := tmpl.Execute(w, SignupReceiptPage{
		UX: ux,
		Settings: &Settings })
	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
}

func HandleReq(w http.ResponseWriter, r *http.Request) {
	const serveDir string = "/"
	dispatchers := map[string]bool {
		"login": true,
		"logout": true,
		"signup": true,
		"static": true,
		"privacy": true,
		"about": true,
		"mission": true,
		"technology": true,
		"u": true }

	parts := strings.Split(
		strings.TrimLeft(
			r.URL.Path, serveDir),
		"/")
	dispatcher, args := parts[0], parts[1:]

	// query := r.URL.Query()
	hasTrailingSlash := (r.URL.Path != strings.TrimRight(r.URL.Path, "/"))

	// Top-level index page should redirect to a search bar
	/* if dispatcher == "" {
		http.Redirect(w, r,
			CONFIG_CANON + "search/", http.StatusMovedPermanently)
		return
	} */

	// ws, err := LoadWebSession(r)
	if !HasWebSession(r) {
		InitWebSession(w, r)
	}

	// Load the user experience...
	db, _ := DBConnect(&Settings)
	ux := LoadUX(db, r)
	res := &ServerRes{
		DB: db,
		Writer: w,
		Request: r}

	// Top-level index page should redirect to a search bar
	if dispatcher == "" {
		/* http.Redirect(w, r,
			serveDir + "about", http.StatusMovedPermanently)
		return */
		ux.HandleWebIndex(res)
		return
	}

	validCall := dispatchers[dispatcher]
	// Undefined func call
	if !validCall {
		HandleWebError(w, r, http.StatusNotFound)
		return
	}

	// Trailing slashes are non-canonical resources
	if hasTrailingSlash {
		http.Redirect(w, r, strings.TrimRight(r.URL.Path, "/"),
			http.StatusMovedPermanently)
		return
	}

	switch(dispatcher) {
	case "signup":
		switch(len(args)) {
		case 0:
			http.Redirect(w, r,
				serveDir + "signup/new",
				http.StatusMovedPermanently)
		case 1:
			step := args[0]

			switch (step) {
			case "new":
				ux.HandleSignupNew(res, nil)
			case "create":
				ux.HandleSignupCreate(res)
			case "pay":
				ux.HandleSignupPay(res)
			case "receipt":
				ux.HandleSignupReceipt(res)
			default:
				HandleWebError(w, r, http.StatusBadRequest)
			}
		}
	case "login":
		ux.HandleLogin(res, nil)
	case "logout":
		ux.HandleLogout(res)
	case "u":
		switch(len(args)) {
		case 3:
			uname := args[0]
			if args[1] == "settings" {
				// User settings at /u/{USER}/settings/{OPTION}
				option := args[2]
				ux.HandleUserSettings(res, uname, option)
			} else {
				// Edit bookmark /u/{USER}/{ID}/{ACTION}
				bID, err := strconv.Atoi(args[1])
				action := args[2]

				if err != nil {
					HandleWebError(w, r, http.StatusBadRequest)
					return
				}
				ux.HandleBMarkAction(res, uname, bID, action)
			}
		case 2:
			uname := args[0]
			action := args[1]

			switch(action) {
			case "add": ux.HandleUserAdd(res, uname)
			case "archive": ux.HandleUserViewArchive(res, uname)
			case "settings": ux.HandleUserSettings(res, uname, "")
			default: HandleWebError(w, r, http.StatusNotFound)
			}
		case 1:
			uname := args[0]

			ux.HandleUserReq(res, uname)
		default:
			HandleWebError(w, r, http.StatusBadRequest)
		}
	case "static":
		HandleStatic(res)
	case "about": fallthrough
	case "technology": fallthrough
	case "privacy": fallthrough
	case "mission":
		// Static pages
		if len(args) > 0 {
			HandleWebError(w, r, http.StatusNotFound)
		} else { ux.HandleInfo(res, dispatcher) }
	default:
		HandleWebError(w, r, http.StatusNotFound)
	}
}

func main() {
	err := ReadDefaultConfig(&Settings)
	if err != nil { panic(err) }

	InitTemplates()

	log.Println("Starting server...")

	http.HandleFunc("/", HandleReq)
	http.ListenAndServe(Settings.Web.Host, nil)
}

func HandleWebError(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	switch(status) {
	case http.StatusNotFound:
		fmt.Fprint(w, "Custom 404")
	case http.StatusInternalServerError:
		fmt.Fprint(w, "Custom 500")
	case http.StatusBadRequest:
		fmt.Fprint(w, "Custom 400")
	case http.StatusNoContent:
		fmt.Fprint(w, "Custom 204")
	case http.StatusServiceUnavailable:
		fmt.Fprint(w, "Custom 503")
	case http.StatusForbidden:
		fmt.Fprint(w, "Custom 403")
	case http.StatusMethodNotAllowed:
		fmt.Fprint(w, "Custom 405")
	}
}

func AppendQuery(u *url.URL, key, value string, clobber bool) (*url.URL) {
	ret := new(url.URL)

	query := u.Query()
	if clobber { query.Set(key, value)
	} else { query.Add(key, value) }

	*ret = *u
	ret.RawQuery = query.Encode()
	return ret
}

func StarStarStar(times int) (string) { return strings.Repeat("*", times) }
