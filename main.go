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

	println("Listening to 8080")

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

			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			println(line)

			line = strings.TrimRight(line, "\r\n")
			parts := strings.Split(line, " ")

			for index, value := range parts {
				println(index, value)
			}

			if len(parts) != 3 {
				MyHttpMessage(conn, "400", "Bad Request", "Too many")
				return
			}
			if parts[2] != "HTTP/1.1" {
				MyHttpMessage(conn, "501", "Bad Request", "Not HTTP/1.1")
				return
			}
			if parts[0] != "GET" && parts[0] != "POST" && parts[0] != "DELETE" && parts[0] != "PUT" {
				MyHttpMessage(conn, "400", "Bad Request", "Not valid HTTP Method")
				return
			}
			if parts[1] == "/ping" {
				MyHttpMessage(conn, "200", "OK", "pong")
				return
			} else if parts[1] != "/" {
				MyHttpMessage(conn, "404", "Not Found", "Not found")
				return
			} else {
				MyHttpMessage(conn, "200", "OK", "You're good!")
				return
			}

			newreader := bufio.NewReader(myConnection)
			var fullHeader string

			for {
				line, err := newreader.ReadString('\n')
				if err != nil {
					break
				}

				fullHeader += line

				// If a line is just "\r\n" or "\n", we have reached the end of the headers
				if line == "\r\n" || line == "\n" {
					break
				}
			}

		}(conn)

	}

}
func MyHttpMessage(myConnection net.Conn, code string, res string, why string) {
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
