package main

import (
	"bytes"
	"fmt"
	"github.com/jlaffaye/ftp"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
	"crypto/rand"
	"net/http"
	"html/template"
)

type command struct {
	Cmd string
	Answer string
}

const COOKIE_NAME = "sessionId"

var (
	addr, user, password string
	connection *ftp.ServerConn = nil
	inMemorySession = NewSession()
)

func main () {
	http.HandleFunc("/", LoginRouterhandler)
	http.HandleFunc("/term", TerminalRouterhandler)
	http.HandleFunc("/logout", LogoutRouterHandler)
	err := http.ListenAndServe(":9000", nil)
	check(err)
}

func LoginRouterhandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		t, err := template.ParseFiles("login.html")
		check(err)
		t.ExecuteTemplate(w, "index", nil)
	}

	if r.Method == "POST" {
		addr = r.FormValue("addr")
		user = r.FormValue("user")
		password = r.FormValue("pass")
		if user != "" && password != "" && addr != "" {
			c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
			if err == nil {
				err = c.Login(user, password)
				if err == nil {
					sessiodId := inMemorySession.Init(user)
					cookie := &http.Cookie{
						Name: 		COOKIE_NAME,
						Value: 		sessiodId,
						Expires: 	time.Now().Add(5 * time.Minute),
						HttpOnly: 	true,
					}
					connection = c
					http.SetCookie(w, cookie)
					http.Redirect(w, r, "/term", http.StatusSeeOther)
				} else {
					fmt.Fprint(w, "<p>Wrong login or password or address</p>")
				}
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
		t, err := template.ParseFiles("terminal.html")
		check(err)

		cmd_str := r.FormValue("cmdIn")
		if cmd_str != "" {
			cmd := command{Cmd: cmd_str, Answer: "",}
			arr := strings.Split(cmd_str, " ")
			if len(arr) == 1 {
				if arr[0] == "quit" {
					cmd.Answer = "Press Exit button"
				} else if arr[0] == "list" {
					res_str := ""
					curDir, _ := connection.CurrentDir()
					for _, v := range getDirectory(curDir) {
						res_str += (v.Name + "\n")
					}
					cmd.Answer = res_str
				} else {
					cmd.Answer = "Wrong command"
				}
			} else if len(arr) == 2 {
				if arr[0] == "stor" {
					curDir, _ := connection.CurrentDir()
					uploadFile(arr[1], curDir)
				} else if arr[0] == "retr" {
					downloadFile(arr[1], "/Users/ATOM/Desktop")
				} else if arr[0] == "mkd" {
					createDirectory(arr[1])
				} else if arr[0] == "dele" {
					deleteFile(arr[1])
				} else if arr[0] == "rmd" {
					err = connection.RemoveDir(arr[1])
					check(err)
				} else if arr[0] == "list" {
					res_str := ""
					for _, v := range getDirectory(arr[1]) {
						res_str += (v.Name + "\n")
					}
					cmd.Answer = res_str
				} else {
					cmd.Answer = "Wrong command"
				}
			} else {
				cmd.Answer = "Wrong command"
			}
			inMemorySession.getData(id).Sources = append(inMemorySession.getData(id).Sources, cmd)
		}
		t.ExecuteTemplate(w, "index", inMemorySession.getData(id).Sources)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func LogoutRouterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	err := connection.Quit()
	check(err)
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

//проверка на ошибку
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

//получить файл в виде набора байтов
func getFile(path string) *bytes.Reader {
	data, err := ioutil.ReadFile(path)
	check(err)
	return bytes.NewReader(data)
}

//выход
func exit () {
	err := connection.Quit()
	check(err)
}

//создание директории
func createDirectory (path string) {
	err:= connection.MakeDir(path)
	check(err)
}

//загрузка файла
func uploadFile(path string, ftp_path string) {
	err := connection.ChangeDir(ftp_path)
	check(err)
	err = connection.Stor(path, getFile(path))
	check(err)
}

//удаление файла на ftp-сервере
func deleteFile(path string) {
	err := connection.Delete(path)
	check(err)
}

//скачивание файла
func downloadFile (path string, local_path string) {
	r, err := connection.Retr(path)
	check(err)
	err = os.Chdir(local_path)
	check(err)
	name := path[strings.LastIndex(path, "/") + 1:]
	file, err := os.Create(name)
	check(err)
	buf, err := ioutil.ReadAll(r)
	check(err)
	file.Write(buf)
	file.Close()
	r.Close()
}

//получение содержимого директории
func getDirectory(path string) (data []*ftp.Entry) {
	data, err := connection.List(path)
	check(err)
	return
}

//185.20.227.83:1969
//admin
//123456

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