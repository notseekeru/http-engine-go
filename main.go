package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
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

		go handleConnection(conn)

	}

}
func handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	headerHashmap := make(map[string]string)
	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	println()
	print("Request Line: ", requestLine)

	requestLine = strings.TrimRight(requestLine, "\r\n")
	requestParts := strings.Split(requestLine, " ")

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
		strValue := headerHashmap["Content-Length"]
		intValue64, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			fmt.Printf("PARSING ERROR: Could not convert string %q to int: %v\n", strValue, err)
		}

		bodyReader := io.LimitReader(reader, intValue64)
		bodyBytes, err := io.ReadAll(bodyReader)
		if err != nil {
			panic(err)
		}
		fmt.Printf("HTTP Body payload: %s\n", string(bodyBytes))

	} else {
		fmt.Println("INF: No HTTP Body payload found")
	}

	switch requestParts[1] {
	case "/":
		MyHTTPMessage(conn, "200", "OK", "index.html File Sent", "html")
		return
	case "/ping":
		MyHTTPMessage(conn, "200", "OK", "pong")
		return
	default:
		MyHTTPMessage(conn, "404", "Not Found", "Not Found")
		return
	}
}

func MyHTTPMessage(myConnection net.Conn, statusCode string, resCode string, messageBody string, contentType ...string) {
	// Server -> Client
	datenow := time.Now()
	server := "GoLang NixOS TCP/HTTP Engine"
	connection := "keep-alive"
	var body string
	var bodyBytes []byte
	var content string

	if len(contentType) > 0 && contentType[0] == "html" {
		content = "text/html"
		bodyBytes, _ = os.ReadFile("index.html")
		body = string(bodyBytes)
	} else {
		body = messageBody + "\n"
		content = "text/plain"
	}

	bodyLength := strconv.Itoa(len(body))

	serverResponse := "HTTP/1.1 " + statusCode + " " + resCode + "\r\n" +
		"Date: " + datenow.UTC().Format(time.RFC1123) + "\r\n" +
		"Server: " + server + "\r\n" +
		"Content-Length: " + bodyLength + "\r\n" +
		"Content-Type: " + content + "\r\n" +
		"Connection: " + connection + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}
