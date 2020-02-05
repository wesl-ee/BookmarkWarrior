package main

import (
	"log"
	"net/http"
	"html/template"
	_ "strconv"
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
	UX *UserExperience
	Settings *Config }

type LoginPage struct {
	Settings *Config
	UX *UserExperience }

type UserPage struct {
	Settings *Config
	User WebUserProfile
	UX *UserExperience
	Title string }

type SignupError struct {
	Mismatch bool
	Taken bool
	BadUName bool
	ShortPassword bool
	AlreadyLoggedIn bool }

type SignupNewPage struct {
	LoggedIn bool
	Settings *Config
	UX *UserExperience
	Error *SignupError}

type SignupCreatePage struct {
	Username string
	DisplayName string
	Password string
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

	tmpl := Templates[page]
	err := tmpl.Execute(w, IndexPage{
		Settings: &Settings,
		UX: ux})
	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
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
		return }

	marks, err := user.Bookmarks(db)
	if err != nil {
		// Databse error...
		HandleWebError(w, r, http.StatusServiceUnavailable)
		return
	}

	webuser:= user.AsWebEntity()
	webuser.Bookmarks = marks
	webuser.ThisIsMe = false

	tmpl := Templates[page]
	err = tmpl.Execute(w, UserPage{
		Settings: &Settings,
		User: webuser,
		UX: ux,
		Title: user.DisplayName + " (" + uname + ") - Bookmarks" })

	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
}

func (ux *UserExperience) HandleLogout(res *ServerRes) {
	ws := ThisSession(res.Request)
	ws.Disassociate(res.DB)

	http.Redirect(res.Writer, res.Request, "/", http.StatusSeeOther)
}

func (ux *UserExperience) HandleLogin(res *ServerRes) {
	w := res.Writer
	r := res.Request
	db := res.DB
	page := "tmpl/login.html"
	tmpl := Templates[page]

	// Handle login attempts
	if (r.Method == "POST") {
		if err := r.ParseForm(); err != nil { panic(err) }
		username := strings.ToLower(r.FormValue("username"))
		password := r.FormValue("password")

		u, err := LetMeIn(db, username, password)
		if err != nil { panic(err) }

		ws := ThisSession(r)
		ws.Associate(db, u.Username)

		http.Redirect(res.Writer, res.Request, "/u/" + username, http.StatusSeeOther)
		return
	}

	err := tmpl.Execute(w, LoginPage{
		UX: ux,
		Settings: &Settings })
	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
}

func HandleStatic(res *ServerRes) {
	w := res.Writer
	r := res.Request
	http.ServeFile(w, r, "static/" + strings.TrimPrefix(
		r.URL.Path, "/static/"))
}

// 1st step in acc creation...
func (ux *UserExperience) HandleSignupNew(res *ServerRes, e *SignupError) {
	w := res.Writer
	r := res.Request
	page := "tmpl/signup-new.html"

	tmpl := Templates[page]
	err := tmpl.Execute(w, SignupNewPage{
		Error: e,
		LoggedIn: false,
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
	page := "tmpl/signup-create.html"

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

	if displayname == "" {
		displayname = username
	}

	if len(password) < Settings.MinimumPasswordLength {
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

	tmpl := Templates[page]
	err = tmpl.Execute(w, SignupCreatePage{
		Username: username,
		DisplayName: displayname,
		Password: password,
		UX: ux,
		Settings: &Settings })
	if err != nil {
		HandleWebError(w, r, http.StatusInternalServerError)
	}
}

// 3rd step in acc creation...
func (ux *UserExperience) HandleSignupPay(res *ServerRes) {
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
	orderID := r.FormValue("orderid")

	// Poll PayPal for payment verification
	paid, err := VerifyPayment(orderID)
	if err != nil { panic(err) }

	if paid {
		// Actually create account in DB
		// ...
		newUser := UserProfile{
			Username: username,
			DisplayName: displayname, }

		u, err := newUser.Create(db, password)
		if err != nil { panic(err) }
		fmt.Println(u)

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

			switch(step) {
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
		ux.HandleLogin(res)
	case "logout":
		ux.HandleLogout(res)
	case "u":
		switch(len(args)) {
		case 1:
			uname := args[0]

			ux.HandleUserReq(res, uname)
		default:
			HandleWebError(w, r, http.StatusBadRequest)
		}
	case "static":
		HandleStatic(res)
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
