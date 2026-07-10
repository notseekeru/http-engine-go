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

			headerHashmap := make(map[string]string)
			reader := bufio.NewReader(myConnection)

			// FIRST REQUEST LINE PARSING OF HTTP MESSAGE
			requestLine, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			println()
			print("Request Line: ", requestLine)

			requestLine = strings.TrimRight(requestLine, "\r\n")
			requestParts := strings.Split(requestLine, " ")

			// ROUTING LOGIC
			if len(requestParts) != 3 {
				MyHTTPMessage(conn, "400", "Bad Request", "Too many")
				return
			}
			if requestParts[2] != "HTTP/1.1" {
				MyHTTPMessage(conn, "400", "Bad Request", "Only HTTP/1.1 supported")
				return
			}
			if requestParts[0] != "GET" && requestParts[0] != "POST" {
				MyHTTPMessage(conn, "405", "Method Not Allowed", "Unsupported HTTP Method")
				return
			}

			// LOOP THROUGH UNPREDICATABLE HEADER LINE MAP IT USING A HASHMAP
			for {
				headerLine, err := reader.ReadString('\n')
				if err != nil {
					println(err)
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
				// CONVERT HASHMAP CONTENTLENGTH TO AN INTEGER64
				strValue := headerHashmap["Content-Length"]
				intValue64, err := strconv.ParseInt(strValue, 10, 64)
				if err != nil {
					fmt.Printf("PARSING ERROR: Could not convert string %q to int: %v\n", strValue, err)
				}

				// READING THE BODY
				bodyReader := io.LimitReader(reader, intValue64)
				bodyBytes, err := io.ReadAll(bodyReader)
				if err != nil {
					panic(err)
				}
				fmt.Printf("HTTP Body payload: %s\n", string(bodyBytes))

			} else {
				fmt.Println("Key not found")
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
	body := why + "\n"
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
