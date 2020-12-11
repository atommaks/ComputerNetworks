package main

import (
	//"bytes"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"github.com/gliderlabs/ssh"
	"os/exec"
	"strings"
)

var (
	u = "iu9_33_08"
	p = "Amnesia1488"
)

func main() {
	ssh.Handle(func(s ssh.Session) {
		term := terminal.NewTerminal(s, s.User() + ":~")
		for {
			line, _ := term.ReadLine()
			if line == "exit" {
				break
			} else {
				inp := strings.Split(line, " ")
				cmd := exec.Command(inp[0], inp[1:]...)
				out, err := cmd.CombinedOutput()
				if err != nil {
					if _, err := io.WriteString(s, "Incorrect command\n"); err != nil {
						log.Fatal("Failed: ", err)
					}

				} else {
					if _, err := io.WriteString(s, string(out)); err != nil {
						log.Fatal("Failed: ", err)
					}
				}
			}
		}
		//term := terminal.NewTerminal(s, "")
		//for {
		//	line, err := term.ReadLine()
		//	if err != nil || line == "exit"{
		//		break
		//	}
		//
		//	str := execute(line) + "\n"
		//	io.WriteString(s, str)
		//	log.Println(line)
		//}
		//str := execute(s.RawCommand())
		//io.WriteString(s, str)
		//term.Write([]byte(str))
		log.Println("terminal executed command")
	})
	log.Println("starting ssh-server on port 1969...")
	log.Fatal(ssh.ListenAndServe(":1969", nil, ssh.PasswordAuth(func (context ssh.Context, pass string) bool {
		return (pass == p && context.User() == u)
	},)))
}

//func execute(command string) string {
//	var buff bytes.Buffer
//	buff.WriteString(command)
//	cmd := exec.Command(command)
//	cmd.Stdin = &buff
//	res, err := cmd.CombinedOutput()
//	if err != nil {
//		return err.Error()
//	}
//	return string(res)
//}