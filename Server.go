package main

import (
	"log"
	"net/http"
	"html/template"
	_ "strconv"
	"net/url"
	"fmt"
	"strings"
	_ "golang.org/x/crypto/bcrypt"
)

var Settings Config

func PageDependencies(page string) ([]string) {
	for _, p := range Settings.Templates {
		if p.Name == page { return p.Dependencies }
	}
	return nil
}

type IndexPage struct {
	Settings *Config
	LoggedIn bool }

type UserPage struct {
	Settings *Config
	User WebUserProfile
	LoggedIn bool
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
	Error *SignupError}

type SignupCreatePage struct {
	Username string
	DisplayName string
	Password string
	LoggedIn bool
	Settings *Config }

type SignupReceiptPage struct {
	LoggedIn bool
	Settings *Config
}

func HandleWebIndex(w http.ResponseWriter, r *http.Request) {
	page := "tmpl/index.html"

	incl := PageDependencies(page)
	tmpl := template.Must(template.ParseFiles(page))
	for _, t := range incl {
		template.Must(tmpl.ParseFiles(t)) }
	tmpl.Execute(w, IndexPage{
		Settings: &Settings,
		LoggedIn: false});
}

func HandleUserReq(w http.ResponseWriter, r *http.Request, uname string) {
	db, _ := DBConnect(&Settings)
	page := "tmpl/user.html"

	incl := PageDependencies(page)
	tmpl := template.Must(template.ParseFiles(page))
	user, err := UserByName(db, uname)
	if err != nil { HandleWebError(w, r, http.StatusNotFound)
		return }

	webuser := user.AsWebEntity()
	webuser.Bookmarks = user.Bookmarks(db)
	webuser.ThisIsMe = false

	for _, t := range incl { tmpl.ParseFiles(t) }
	tmpl.Execute(w, UserPage{
		Settings: &Settings,
		User: webuser,
		LoggedIn: false,
		Title: user.DisplayName + " (" + uname + ") - Bookmarks" })
}

func HandleStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/" + strings.TrimPrefix(
		r.URL.Path, "/static/"))
}

// 1st step in acc creation...
func HandleSignupNew(w http.ResponseWriter, r *http.Request, e *SignupError) {
	page := "tmpl/signup-new.html"

	incl := PageDependencies(page)
	tmpl := template.Must(template.ParseFiles(page))
	for _, t := range incl { tmpl.ParseFiles(t) }
	tmpl.Execute(w, SignupNewPage{
		Error: e,
		LoggedIn: false,
		Settings: &Settings })
}

// 2nd step in acc creation...
func HandleSignupCreate(w http.ResponseWriter, r *http.Request) {
	db, _ := DBConnect(&Settings)
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
		HandleSignupNew(w, r, &SignupError{ShortPassword: true})
		return
	}

	if !ValidUsername(username) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		HandleSignupNew(w, r, &SignupError{BadUName: true})
		return
	}

	existing, err := UserByName(db, username)
	if err == nil && existing.Username == username {
		w.WriteHeader(http.StatusConflict)
		HandleSignupNew(w, r, &SignupError{Taken: true})
		return
	}

	if confirmpassword != password {
		w.WriteHeader(http.StatusUnprocessableEntity)
		HandleSignupNew(w, r, &SignupError{Mismatch: true})
		return
	}

	incl := PageDependencies(page)
	tmpl := template.Must(template.ParseFiles(page))
	for _, t := range incl { tmpl.ParseFiles(t) }
	tmpl.Execute(w, SignupCreatePage{
		Username: username,
		DisplayName: displayname,
		Password: password,
		LoggedIn: false,
		Settings: &Settings })
}

// 3rd step in acc creation...
func HandleSignupPay(w http.ResponseWriter, r *http.Request) {

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
		db, _ := DBConnect(&Settings)
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
func HandleSignupReceipt(w http.ResponseWriter, r *http.Request) {
	page := "tmpl/signup-receipt.html"

	incl := PageDependencies(page)
	tmpl := template.Must(template.ParseFiles(page))
	for _, t := range incl { tmpl.ParseFiles(t) }
	tmpl.Execute(w, SignupReceiptPage{
		LoggedIn: false,
		Settings: &Settings })
}

/* func HandleSignup(w http.ResponseWriter, r *http.Request, step string) {
	page := "tmpl/signup.html"

	incl := PageDependencies(page)
	tmpl := template.Must(template.ParseFiles(page))
	for _, t := range incl { tmpl.ParseFiles(t) }
	tmpl.Execute(w, SignupPage{
		LoggedIn: false,
		Settings: &Settings }) */

func HandleReq(w http.ResponseWriter, r *http.Request) {
	const serveDir string = "/"
	dispatchers := map[string]bool {
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

	// Top-level index page should redirect to a search bar
	if dispatcher == "" {
		/* http.Redirect(w, r,
			serveDir + "about", http.StatusMovedPermanently)
		return */
		HandleWebIndex(w, r)
		return
	}

	// Trailing slashes are non-canonical resources
	if hasTrailingSlash {
		http.Redirect(w, r, strings.TrimRight(r.URL.Path, "/"),
			http.StatusMovedPermanently)
		return
	}

	validCall := dispatchers[dispatcher]
	// Undefined func call
	if !validCall {
		HandleWebError(w, r, http.StatusNotFound)
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
				HandleSignupNew(w, r, nil)
			case "create":
				HandleSignupCreate(w, r)
			case "pay":
				HandleSignupPay(w, r)
			case "receipt":
				HandleSignupReceipt(w, r)
			default:
				HandleWebError(w, r, http.StatusBadRequest)
			}
		}
	case "u":
		switch(len(args)) {
		case 1:
			uname := args[0]

			HandleUserReq(w, r, uname)
		default:
			HandleWebError(w, r, http.StatusBadRequest)
		}
	case "static":
		HandleStatic(w, r)
	default:
		HandleWebError(w, r, http.StatusNotFound)
	}
}

func main() {
	err := ReadDefaultConfig(&Settings)
	if err != nil { panic(err) }

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
