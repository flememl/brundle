package main

import (
	"fmt"
	"github.com/scorredoira/email"
	"html/template"
	"io/ioutil"
	"net/http"
)

type Config struct {
	Port string
}

var config = Config{"8070"}

type UploadedFile struct {
	Data     []byte
	Filename string
}

type BugReport struct {
	Product     string
	Category    string
	Email       string
	Action      string
	Context     string
	Description string
	Screenshot  *UploadedFile
}

func getValues(r *http.Request) (*BugReport, error) {
	br := &BugReport{
		Product:     r.FormValue("product"),
		Category:    r.FormValue("category"),
		Email:       r.FormValue("email"),
		Action:      r.FormValue("action"),
		Context:     r.FormValue("context"),
		Description: r.FormValue("description"),
	}
	file, fheader, _ := r.FormFile("screenshot")
	if file != nil {
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		br.Screenshot = &UploadedFile{Data: data, Filename: fheader.Filename}
	}
	return br, nil
}

func (br *BugReport) send() error {
	title := fmt.Sprintf("[%s][%s] %s", br.Product, br.Category, br.Action)
	body := fmt.Sprintln("\nContext:", br.Context)
	body += fmt.Sprintln("\nDescription:", br.Description)
	m := email.NewMessage(title, body)
	m.To = []string{"bug@kpsule.me"}
	m.From = br.Email
	if br.Screenshot != nil {
		m.Attachments[br.Screenshot.Filename] = br.Screenshot.Data
	}
	return email.SendUnencrypted("aspmx.l.google.com:25", "", "", m)
}

var templates = template.Must(template.ParseFiles(
	"views/report.html",
	"views/success.html",
	"views/error.html"))

func renderTemplate(w http.ResponseWriter, tpl string, br *BugReport) {
	err := templates.ExecuteTemplate(w, tpl+".html", br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Path[1:]
	if len(page) == 0 {
		page = "report"
	}
	renderTemplate(w, page, nil)
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	br, err := getValues(r)
	if err != nil {
		/* XXX: Here we should highlight the errors instead */
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	err = br.send()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/success", http.StatusFound)
}

func main() {
	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/success", pageHandler)
	http.HandleFunc("/error", pageHandler)
	http.HandleFunc("/send", sendHandler)
	http.Handle("/views/style/", http.StripPrefix("/views/style/",
		http.FileServer(http.Dir("views/style"))))
	http.ListenAndServe(":"+config.Port, nil)
}
