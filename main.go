package main

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	myListener, err := net.Listen("tcp", ":8080")

	if err != nil {
		panic(err)
	}

	defer myListener.Close()

	for {
		conn, err := myListener.Accept()
		if err != nil {
			println(err)
			break
		}

		go func(myConnection net.Conn) {
			defer myConnection.Close()
			myConnection.SetDeadline(time.Now().Add(5 * time.Second))

			reader := bufio.NewReader(myConnection)

			requestLine, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			print("first line: ", requestLine)

			requestLine = strings.TrimRight(requestLine, "\r\n")
			requestParts := strings.Split(requestLine, " ")

			if len(requestParts) != 3 {
				MyHTTPMessage(conn, "400", "Bad Request", "Too many")
				return
			}
			if requestParts[2] != "HTTP/1.1" {
				MyHTTPMessage(conn, "501", "Bad Request", "Not HTTP/1.1")
				return
			}
			if requestParts[0] != "GET" && requestParts[0] != "POST" && requestParts[0] != "DELETE" && requestParts[0] != "PUT" {
				MyHTTPMessage(conn, "400", "Bad Request", "Not valid HTTP Method")
				return
			}

			switch requestParts[1] {
			case "/":
				MyHTTPMessage(conn, "200", "OK", "You re good!")
				return
			case "/ping":
				MyHTTPMessage(conn, "200", "OK", "pong")
				return
			default:
				MyHTTPMessage(conn, "404", "Not Found", "Not found")
				return
			}

		}(conn)

	}

}
func MyHTTPMessage(myConnection net.Conn, code string, res string, why string) {
	// Server -> Client. So we use server.
	datenow := time.Now()
	server := "GoLang NixOS"
	content := "text/plain"
	body := "Hello! " + why
	bodyLength := strconv.Itoa(len(body))
	connection := "close"

	serverResponse := "HTTP/1.1 " + code + " " + res + "\r\n" +
		"Date: " + datenow.Format(time.RFC1123) + "\r\n" +
		"Server: " + server + "\r\n" +
		"Content-Length: " + bodyLength + "\r\n" +
		"Content-Type: " + content + "\r\n" +
		"Connection: " + connection + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}
