package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type Data struct {
	Mail string `json:"mail"`
	Host string `json:"host"`
	Port int    `json:"port"`
	Hash []byte `json:"hash"`
}

var to, subject, msg string

func main () {
	data := Data{}
	getData(&data)
	key := ""
	fmt.Print("Enter key phrase: ")
	fmt.Scan(&key)
	password, err := decrypt(data.Hash, key)
	check(err)
	fmt.Println(string(password))
	fmt.Print("To: ")
	fmt.Scan(&to)
	fmt.Print("Subject: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	subject += scanner.Text()
	fmt.Print("Message:\n")
	var builder strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal(err)
			}
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	msg = builder.String()
	fmt.Println("Sending ...")
	a := smtp.PlainAuth("", data.Mail, string(password), data.Host)
	str := "Subject: " + subject + "\r\n" + msg
	err = smtp.SendMail(fmt.Sprintf("%s:%d", data.Host, data.Port), a, data.Mail, []string{to}, bytes.NewBufferString(str).Bytes())
	check(err)
	fmt.Println("Sent")
}

func getData (data *Data) {
	b, err := ioutil.ReadFile("data.json")
	check(err)
	err = json.Unmarshal(b, &data)
	check(err)
}

func decrypt (data []byte, passPhrase string) ([]byte, error) {
	key := []byte(createHash(passPhrase))
	block, err := aes.NewCipher(key)
	check(err)
	gcm, err := cipher.NewGCM(block)
	check(err)
	nonceSize := gcm.NonceSize()
	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	return plainText, err
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func check (err error) {
	if err != nil {
		log.Fatal(err)
	}
}