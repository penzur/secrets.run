package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
)

var (
	secrets = map[string]string{}

	linkRx    = regexp.MustCompile(`https?:\/\/[a-zA-Z0-9\.:]+\/.{8,}`)
	spaceRx   = regexp.MustCompile(`\s+`)
	browserRx = regexp.MustCompile(`(?i)gecho|firefox|chrome|presto|applewebkit|trident/`)

	env  = os.Getenv("ENV")
	key  = os.Getenv("CSRF_KEY")
	port = ":" + os.Getenv("PORT")
)

func main() {
	isSecure := true
	if env == "" {
		isSecure = false
	}

	r := pat.New()
	r.StrictSlash(true)

	r.Post("/", create)
	r.Get("/{id}", view)
	r.Get("/", home)

	protect := csrf.Protect([]byte(key), csrf.Secure(isSecure))

	log.Fatal(http.ListenAndServe(
		port,
		handlers.LoggingHandler(os.Stdout, browser(protect(r))),
	))
}

func browser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent := r.UserAgent()

		if !browserRx.MatchString(agent) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func home(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")

	if linkRx.MatchString(link) {
		render(w, "secret", map[string]interface{}{
			"link": link,
		})
		return
	}

	render(w, "home", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func view(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	if id == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if secret, ok := secrets[id]; ok {
		render(w, "secret", map[string]interface{}{
			"secret": secret,
		})

		delete(secrets, id)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	secret := r.PostFormValue("secret")
	secret = spaceRx.ReplaceAllString(secret, " ")
	secret = strings.Trim(secret, " ")

	if secret == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	host := r.Host

	id := uuid.New().String()[:8]
	secrets[id] = secret

	link := fmt.Sprintf("%s://%s/%s", proto, host, id)
	link = url.QueryEscape(link)

	http.Redirect(w, r, "/?link="+link, http.StatusSeeOther)
}

func render(w http.ResponseWriter, secret string, ctx interface{}) {
	tmpl, _ := template.New("").ParseFiles(
		"tmpl/layout.tmpl",
		fmt.Sprintf("tmpl/%s.tmpl", secret),
		"tmpl/css/main.css",
	)
	_ = tmpl.ExecuteTemplate(w, "layout", ctx)
}
