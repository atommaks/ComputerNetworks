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
)

func main(){
	c, err := ftp.Dial("185.20.227.83:1969", ftp.DialWithTimeout(5*time.Second))
	check(err)

	err = c.Login("admin", "123456")
	check(err)

	test(c)

	err = c.Quit()
	check(err)
}

//юнит-тесты
func test(c *ftp.ServerConn) {
	createDirectory("/Maxim_dir", c)
	uploadFile("brad-pitt.jpeg", "/Maxim_dir", c)
	downloadFile("/Maxim_dir/brad-pitt.jpeg", "/Users/ATOM/Desktop", c)
	createDirectory("/Maxim_dir/new", c)
	for _, v := range getDirectory("/Maxim_dir", c) {
		fmt.Println(v.Name)
	}
	deleteFile("/Maxim_dir/brad-pitt.jpeg", c)
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