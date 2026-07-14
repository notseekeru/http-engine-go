package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {
	myListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		println(err.Error())
		return
	}

	defer myListener.Close()

	for {
		conn, err := myListener.Accept()
		if err != nil {
			println(err.Error())
			break
		}

		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		queryMap := make(map[string]string)
		headerMap := make(map[string]string)

		requestLine, err := reader.ReadString('\n')
		if err != nil {
			return
		}

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

		relativeURI := requestParts[1]

		if strings.Contains(relativeURI, "?") {
			endpoint, queryStripped, found := strings.Cut(relativeURI, "?")
			if found {
				queryMap["endpoint"] = endpoint
			}

			parameters := queryStripped

			for value := range strings.SplitSeq(parameters, "&") {
				resultKV := strings.Split(value, "=")
				if len(resultKV) != 2 || slices.Contains(resultKV, "") {
					MyHTTPMessage(conn, "400", "Bad Request", "Malformed query")
					return
				}
				queryMap[resultKV[0]] = resultKV[1]
			}

		} else {
			queryMap["endpoint"] = relativeURI
		}

		for {
			headerLine, err := reader.ReadString('\n')
			if err != nil {
				println(err.Error())
				return
			}
			headerLine = strings.TrimRight(headerLine, "\r\n")
			if headerLine == "" {
				println("INF: --End of Header--")
				break
			}
			headerParts := strings.SplitN(headerLine, ": ", 2)
			headerMap[headerParts[0]] = headerParts[1]
			fmt.Printf("INF: Header line: %q\n", headerLine)
		}

		if value, ok := headerMap["Content-Length"]; ok {
			fmt.Println("Content-Length Value:", value)
			strValue := headerMap["Content-Length"]
			intValue64, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				fmt.Printf("ERR: Could not convert string %q to int: %v\n", strValue, err)
			}

			if intValue64 == 0 {
				println("INF: Content-Length = 0, skipping body read")
			} else if intValue64 > 9999999 {
				println("INF: Content-Length = 9999999, skipping body read")
			} else {
				bodyReader := io.LimitReader(reader, intValue64)
				bodyBytes, err := io.ReadAll(bodyReader)
				if err != nil {
					println(err.Error())
					return
				}
				fmt.Printf("INF: HTTP Body payload: %s\n", string(bodyBytes))
			}

		} else {
			fmt.Println("INF: No HTTP Body payload found")
		}
		switch queryMap["endpoint"] {
		case "/":
			MyHTTPMessage(conn, "200", "OK", "index.html File Sent", "html")
		case "/ping":
			MyHTTPMessage(conn, "200", "OK", "pong")
		default:
			MyHTTPMessage(conn, "404", "Not Found", "Not Found")
		}

		if strings.ToLower(headerMap["Connection"]) == "close" {
			return
		}
	}
}

func MyHTTPMessage(myConnection net.Conn, statusCode string, statusPhrase string, messageBody string, contentType ...string) {
	// Server -> Client
	datenow := time.Now()
	server := "GoLang NixOS TCP/HTTP Engine"
	var body string
	var content string

	if len(contentType) > 0 && contentType[0] == "html" {
		content = "text/html"
		bodyBytes, err := os.ReadFile("index.html")
		if err != nil {
			println(err.Error())
			return
		}
		body = string(bodyBytes)
	} else {
		body = messageBody + "\n"
		content = "text/plain"
	}

	bodyLength := strconv.Itoa(len(body))

	serverResponse := "HTTP/1.1 " + statusCode + " " + statusPhrase + "\r\n" +
		"Date: " + datenow.UTC().Format(time.RFC1123) + "\r\n" +
		"Server: " + server + "\r\n" +
		"Content-Length: " + bodyLength + "\r\n" +
		"Content-Type: " + content + "\r\n" +
		"Connection: " + "keep-alive" + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}
