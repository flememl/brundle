package main

import (
	"fmt"
	"github.com/scorredoira/email"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Port        string
	Logpath     string
	Logfilename string
}

var config = Config{
	Port:        "8070",
	Logpath:     "/var/log/brundle/",
	Logfilename: "brundle.log",
}
var logger *log.Logger

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
		logger.Println("[ERROR] Could not render template:", err)
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
		logger.Println("[ERROR] Could not get values:", err)
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	err = br.send()
	if err != nil {
		logger.Println("[ERROR] Could not send email:", err)
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	logger.Println("Success - mail sent:", br)
	http.Redirect(w, r, "/success", http.StatusFound)
}

func setupLogging() {
	err := os.MkdirAll(config.Logpath, 0755)
	if err != nil {
		log.Fatal(err)
	}
	w, err := os.OpenFile(config.Logpath+config.Logfilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(w, "", log.Ldate|log.Ltime)
	logger.Println("Brundle Start")
}

func main() {
	setupLogging()
	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/success", pageHandler)
	http.HandleFunc("/error", pageHandler)
	http.HandleFunc("/send", sendHandler)
	http.Handle("/favicon.ico", http.StripPrefix("/",
		http.FileServer(http.Dir("views/images"))))
	http.Handle("/views/style/", http.StripPrefix("/views/style/",
		http.FileServer(http.Dir("views/style"))))
	http.ListenAndServe(":"+config.Port, nil)
}
