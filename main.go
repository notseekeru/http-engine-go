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
	reader := bufio.NewReader(conn)

	for {
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		queryParametersHashmap := make(map[string]string)
		headerHashmap := make(map[string]string)

		requestLine, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		println()
		print("Request Line: ", requestLine)

		requestLine = strings.TrimRight(requestLine, "\r\n")
		requestParts := strings.Split(requestLine, " ")

		requestQuery := requestParts[1]

		if strings.Contains(requestQuery, "?") {
			requestEndpoint, requestQueryStripped, found := strings.Cut(requestQuery, "?")
			if found {
				fmt.Println("Base Endpoint:    ", requestEndpoint)
				fmt.Println("Query String:", requestQueryStripped)
			}
			queryParametersHashmap["endpoint"] = requestEndpoint
			var requestParametersStripped []string
			requestParameters := requestQueryStripped
			if strings.Contains(requestParameters, "&") {
				requestParametersStripped = strings.Split(requestParameters, "&")
			}

			if slices.Contains(requestParametersStripped, "") {
				println("err: slice contained \"\"")
				return
			}

			fmt.Printf("requestParametersStripped: %s\n", requestParametersStripped)

			for _, value := range requestParametersStripped {
				result := strings.Split(value, "=")
				queryParametersHashmap[result[0]] = result[1]
				fmt.Printf("%s\n", queryParametersHashmap)
			}
		}

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
				fmt.Printf("ERR: Could not convert string %q to int: %v\n", strValue, err)
			}

			if intValue64 == 0 {
				println("WARN: Content-Length = 0")
				break
			}

			bodyReader := io.LimitReader(reader, intValue64)
			bodyBytes, err := io.ReadAll(bodyReader)
			if err != nil {
				panic(err)
			}
			fmt.Printf("INF: HTTP Body payload: %s\n", string(bodyBytes))

		} else {
			fmt.Println("INF: No HTTP Body payload found")
		}
		println(queryParametersHashmap["endpoint"])
		switch queryParametersHashmap["endpoint"] {
		case "/":
			MyHTTPMessage(conn, "200", "OK", "index.html File Sent", "html")
		case "/ping":
			MyHTTPMessage(conn, "200", "OK", "pong")
		default:
			MyHTTPMessage(conn, "404", "Not Found", "Not Found")
		}

		if strings.ToLower(headerHashmap["Connection"]) == "close" {
			return
		}
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
