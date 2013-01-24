package main

import (
	"fmt"
	"html/template"
	"net/http"
	"github.com/scorredoira/email"
)

type BugReport struct {
	Product     string
	Category    string
	Email       string
	Action      string
	Context     string
	Description string
}

func (br *BugReport) send() error {
	title := fmt.Sprintf("[%s][%s] %s", br.Product, br.Category, br.Action)
	body := fmt.Sprintln("Context:", br.Context)
	body += fmt.Sprintln("\nDescription:", br.Description)
	m := email.NewMessage(title, body)
	m.To = []string{"bug@kpsule.me"}
	m.From = br.Email
	return email.SendUnencrypted("aspmx.l.google.com:25", "", "", m)
}

var templates = template.Must(template.ParseFiles("views/report.html", "views/success.html"))

func renderTemplate(w http.ResponseWriter, tpl string, br *BugReport) {
	err := templates.ExecuteTemplate(w, tpl+".html", br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "report", &BugReport{})
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "success", nil)
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	br := &BugReport{
		Product:     r.FormValue("product"),
		Category:    r.FormValue("category"),
		Email:       r.FormValue("email"),
		Action:      r.FormValue("action"),
		Context:     r.FormValue("context"),
		Description: r.FormValue("description"),
	}
	err := br.send()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/success", http.StatusFound)
}

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/success", successHandler)
	http.HandleFunc("/send", sendHandler)
	http.Handle("/views/style/", http.StripPrefix("/views/style/",
		http.FileServer(http.Dir("views/style"))))
	http.ListenAndServe(":8080", nil)
}
