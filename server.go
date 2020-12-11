package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"math"
)

func main() {
	fmt.Println("Launching server...")
	// Устанавливаем прослушивание порта
	ln, _ := net.Listen("tcp", ":8081")
	// Открываем порт
	conn, _ := ln.Accept()

	for {
		// Будем прослушивать все сообщения разделенные \n
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// Распечатываем полученое сообщение
		fmt.Print("Message Received:", string(message))
		// Процесс выборки для полученной строки
		msg := message[:len(message) - 1]
		R, _ := strconv.ParseFloat(string(msg), 64)
		S := R * R * math.Pi
		newmessage := fmt.Sprintf("%f", S)
		// Отправить новую строку обратно клиенту
		conn.Write([]byte(newmessage + "\n"))
	}
}