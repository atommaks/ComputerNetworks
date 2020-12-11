package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

var (
	sources  []command
	cmd, sshCmd, u, address, p string
	config = &ssh.ClientConfig {}
)

type command struct {
	cmd string
	answer string
}

func main() {
	logged := false
	for {
		fmt.Scan(&sshCmd)
		if sshCmd == "ssh" {
			fmt.Scan(&address)
			if address != "" {
				fmt.Scan(&u)
				if u != "" {
					fmt.Print("Enter password: ")
					for i := 0; i < 3; i++ {
						fmt.Scan(&p)
						config = &ssh.ClientConfig{
							User: u,
							HostKeyCallback: ssh.InsecureIgnoreHostKey(),
							Auth: []ssh.AuthMethod{
								ssh.Password(p),
							},
						}
						client, err := ssh.Dial("tcp", address, config)
						if err == nil {
							defer client.Close()
							logged = true
							fmt.Println("Welcome " + u)
							break
						} else {
							fmt.Println("Try password again")
						}
					}
					if logged {
						break
					} else {
						fmt.Println(u + ": Permission denied (publickey,password)")
						os.Exit(1)
					}
				} else {
					fmt.Println("Help: ssh addres username")
				}
			} else {
				fmt.Println("Help: ssh addres username")
			}
		} else {
			fmt.Println("Help: ssh addres username")
		}
	}
	for {
		fmt.Print("> " + u + ": ")
		myscanner := bufio.NewScanner(os.Stdin)
		myscanner.Scan()
		cmd := myscanner.Text()
		if cmd == "exit" {
			os.Exit(1)
		} else {
			fmt.Println(executeCmd(cmd))
		}
	}
}

func executeCmd(cmd string) string {
	client, err := ssh.Dial("tcp", address, config)
	check(err)
	defer client.Close()
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