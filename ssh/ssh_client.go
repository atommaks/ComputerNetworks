package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
)

var (
	user      = "iu9_33_08"
	password  = "Amnesia1488"
	addr      = "185.20.227.83:1969"
	sources  []command
)

type command struct {
	cmd string
	answer string
}

func main() {
	config := &ssh.ClientConfig{
		User: user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	client, err := ssh.Dial("tcp", addr, config)
	check(err)
	defer client.Close()
	test(client)
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

func test(client *ssh.Client) {
	fmt.Println("> " + user + ": " + "ls")
	fmt.Println(executeCmd(client, "ls"))
	fmt.Println("> " + user + ": " + "echo ATOOOOOOM >test.txt")
	fmt.Println(executeCmd(client, "echo ATOOOOOOM >test.txt"))
	fmt.Println("> " + user + ": " + "ls")
	fmt.Println(executeCmd(client, "ls"))
	fmt.Println("> " + user + ": " + "rm test.txt")
	fmt.Println(executeCmd(client, "rm test.txt"))
	fmt.Println("> " + user + ": " + "ls")
	fmt.Println(executeCmd(client, "ls"))
}