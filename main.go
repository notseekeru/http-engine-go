package main

import (
	"bufio"
	"fmt"
	"io"
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
			println()
			print("Request Line: ", requestLine)

			requestLine = strings.TrimRight(requestLine, "\r\n")
			requestParts := strings.Split(requestLine, " ")

			headerHashmap := make(map[string]string)

			for {
				headerLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				headerLine = strings.TrimRight(headerLine, "\r\n")
				if headerLine == "" {
					println("--End of Header--")
					break
				}
				result := strings.Split(headerLine, ": ")
				headerHashmap[result[0]] = result[1]
				fmt.Printf("Header line: %q\n", headerLine)
			}
			if value, ok := headerHashmap["Content-Length"]; ok {
				fmt.Println("Content-Length Value:", value)
			} else {
				fmt.Println("Key not found")
				return
			}

			// CONVERT HASHMAP CONTENTLENGTH TO AN INTEGER64
			strValue := headerHashmap["Content-Length"]
			intValue64, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				fmt.Printf("PARSING ERROR: Could not convert string %q to int: %v\n", strValue, err)
			}
			bodyReader := io.LimitReader(reader, intValue64)

			// READING THE BODY
			bodyBytes, err := io.ReadAll(bodyReader)
			if err != nil {
				fmt.Printf("CRITICAL ERROR DURING READ: %v\n", err)
			}
			fmt.Printf("HTTP Body payload: %s\n", string(bodyBytes))

			// ROUTING LOGIC
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
				MyHTTPMessage(conn, "200", "OK", "You've arrived at: /")
				return
			case "/ping":
				MyHTTPMessage(conn, "200", "OK", "pong")
				return
			default:
				MyHTTPMessage(conn, "404", "Not Found", "You've have not arrived due to: Not Found")
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
	body := why
	bodyLength := strconv.Itoa(len(body))
	connection := "close"

	serverResponse := "HTTP/1.1 " + code + " " + res + "\r\n" +
		"Date: " + datenow.UTC().Format(time.RFC1123) + "\r\n" +
		"Server: " + server + "\r\n" +
		"Content-Length: " + bodyLength + "\r\n" +
		"Content-Type: " + content + "\r\n" +
		"Connection: " + connection + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}
