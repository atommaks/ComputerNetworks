package main

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"html/template"
	"net/http"
	"time"
)

const COOKIE_NAME = "sessionId"

var (
	user, addr string
	config = &ssh.ClientConfig{}
	sources []command
	inMemorySession = NewSession()
)

type command struct {
	Cmd string
	Answer string
}

func main() {
	http.HandleFunc("/", LoginRouterhandler)
	http.HandleFunc("/term", TerminalRouterhandler)
	http.HandleFunc("/logout", LogoutRouterHandler)
	err := http.ListenAndServe(":9000", nil)
	check(err)
}

func LogoutRouterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cookie, _ := r.Cookie(COOKIE_NAME)
	id := cookie.Value

	c := &http.Cookie {
		Name:       COOKIE_NAME,
		Value:      "",
		HttpOnly: 	true,
		Expires:	time.Unix(0, 0),
	}
	inMemorySession.Close(id)

	http.SetCookie(w, c)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LoginRouterhandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		t, err := template.ParseFiles("login.html")
		check(err)
		t.ExecuteTemplate(w, "index", nil)
	}

	if r.Method == "POST" {
		u := r.FormValue("user")
		p := r.FormValue("pass")
		addr = r.FormValue("addr")

		if u != "" && p != "" && addr != "" {
			config =  &ssh.ClientConfig{
				User: 				u,
				HostKeyCallback: 	ssh.InsecureIgnoreHostKey(),
				Auth: 				[]ssh.AuthMethod{
					ssh.Password(p),
				},
			}
			client, err := ssh.Dial("tcp", addr, config)

			if err == nil {
				sessiodId := inMemorySession.Init(u)
				cookie := &http.Cookie{
					Name: 		COOKIE_NAME,
					Value: 		sessiodId,
					Expires: 	time.Now().Add(5 * time.Minute),
					HttpOnly: 	true,
				}
				http.SetCookie(w, cookie)
				user = u
				defer client.Close()
				http.Redirect(w, r, "/term", http.StatusSeeOther)
			} else {
				fmt.Fprint(w, "<p>Wrong login or password or address</p>")
			}
		} else {
			fmt.Fprint(w, "Enter login and password and address")
		}
	}
}

func TerminalRouterhandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cookie, _ := r.Cookie(COOKIE_NAME)

	if cookie != nil {
		id := cookie.Value
		client, err := ssh.Dial("tcp", addr, config)
		check(err)
		defer client.Close()

		t, err := template.ParseFiles("terminal.html")
		check(err)

		cmd_str := r.FormValue("cmdIn")
		if cmd_str != "" {
			answer_str := executeCmd(client, cmd_str)
			inMemorySession.getData(id).Sources = append(inMemorySession.getData(id).Sources, command{"> " + user + ": " + cmd_str ,answer_str})
		}
		t.ExecuteTemplate(w, "index", inMemorySession.getData(id).Sources)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func executeCmd(client *ssh.Client, cmd string) string {
	session, err := client.NewSession()
	check(err)
	defer session.Close()
	res, err := session.CombinedOutput(cmd)
	check(err)
	return string(res)
}

func check (err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//*********Session**********//

type sessionData struct {
	Username string
	Sources []command
}

type Session struct {
	data map[string]*sessionData
}

func NewSession() *Session {
	s := new(Session)
	s.data = make(map[string]*sessionData)
	return s
}

func (s *Session) Init(username string) string {
	sessionId := GenerateId()
	arr := make([]command, 0)
	data := &sessionData{Username: username, Sources: arr}
	s.data[sessionId] = data
	return sessionId
}

func (s *Session) Close(sessionId string) {
	delete(s.data, sessionId)
}

func (s *Session) getData(sessionId string) *sessionData {
	return s.data[sessionId]
}

func GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}