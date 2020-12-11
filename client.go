package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	// Подключаемся к сокету
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")
	var r float64
	for {
		// Чтение входных данных от stdin
		fmt.Print("Enter R: ")
		fmt.Scanf("%f", &r)
		// Отправляем в socket
		fmt.Fprintf(conn, fmt.Sprintf("%f", r) + "\n")
		// Прослушиваем ответ
		answer, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print(answer)
	}
}