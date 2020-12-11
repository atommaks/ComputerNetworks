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
	"bufio"
)

var (
	addr, user, password string
)

func main () {
	fmt.Print("Enter address: ")
	fmt.Scan(&addr)
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Print("Enter login: ")
	fmt.Scan(&user)
	fmt.Print("Enter password: ")
	fmt.Scan(&password)
	err = c.Login(user, password)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		command := strings.Split(scanner.Text(), " ")
		if len(command) == 1 {
			if command[0] == "quit" {
				break
			} else if command[0] == "list" {
				curDir, _ := c.CurrentDir()
				for _, v := range getDirectory(curDir, c) {
					fmt.Println(v.Name)
				}
			} else {
				fmt.Println("Wrong command")
			}
		} else if len(command) == 2 {
			if command[0] == "stor" {
				curDir, _ := c.CurrentDir()
				uploadFile(command[1], curDir, c)
			} else if command[0] == "retr" {
				downloadFile(command[1], "/Users/ATOM/Desktop", c)
			} else if command[0] == "mkd" {
				createDirectory(command[1], c)
			} else if command[0] == "dele" {
				deleteFile(command[1], c)
			} else if command[0] == "list" {
				for _, v := range getDirectory(command[1], c) {
					fmt.Println(v.Name)
				}
			} else if command[0] == "rmd" {
				err = c.RemoveDir(command[1])
				check(err)
			} else {
				fmt.Println("Wrong command")
			}
		} else {
			fmt.Println("Wrong command")
		}
	}
	err = c.Quit()
	check(err)

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

//создание директории
func createDirectory (path string, c *ftp.ServerConn) {
	err:= c.MakeDir(path)
	check(err)
}

//загрузка файла
func uploadFile(path string, ftp_path string, c *ftp.ServerConn) {
	err := c.ChangeDir(ftp_path)
	check(err)
	err = c.Stor(path, getFile(path))
	check(err)
}

//удаление файла на ftp-сервере
func deleteFile(path string, c *ftp.ServerConn) {
	err := c.Delete(path)
	check(err)
}

//скачивание файла
func downloadFile (path string, local_path string, c *ftp.ServerConn) {
	r, err := c.Retr(path)
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
func getDirectory(path string, c *ftp.ServerConn) (data []*ftp.Entry) {
	data, err := c.List(path)
	check(err)
	return
}

//185.20.227.83:1969
//admin
//123456