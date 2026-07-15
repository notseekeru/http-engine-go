package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {
	myListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err.Error())
	}

	defer myListener.Close()

	for {
		conn, err := myListener.Accept()
		if err != nil {
			log.Print(err.Error())
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
				log.Print(err.Error())
				MyHTTPMessage(conn, "400", "Bad Request")
				return
			}
			headerLine = strings.TrimRight(headerLine, "\r\n")
			if headerLine == "" {
				break
			}
			headerParts := strings.SplitN(headerLine, ": ", 2)
			headerMap[headerParts[0]] = headerParts[1]
		}

		if _, ok := headerMap["Content-Length"]; ok {
			strValue := headerMap["Content-Length"]
			intValue64, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				log.Printf("Could not convert string %q to int: %v", strValue, err)
				MyHTTPMessage(conn, "500", "Internal Server Error")
				return
			}

			if intValue64 > 0 && intValue64 <= 9999999 && requestParts[0] == "POST" {
				bodyBytes, err := io.ReadAll(io.LimitReader(reader, intValue64))
				if err != nil {
					log.Print(err.Error())
					MyHTTPMessage(conn, "400", "Bad Request")
					return
				}
				log.Printf("DEBUG: POST body: %s", bodyBytes)
			}

		}

		switch queryMap["endpoint"] {
		case "/":
			HTTPFileServe(conn, "200", "OK", "index.html")
		case "/styles.css":
			HTTPFileServe(conn, "200", "OK", "styles.css")
		case "/ping":
			MyHTTPMessage(conn, "200", "OK", "pong")
		default:
			MyHTTPMessage(conn, "404", "Not Found", "Endpoint Not Found")
		}

		if strings.ToLower(headerMap["Connection"]) == "close" {
			return
		}
	}
}

func MyHTTPMessage(myConnection net.Conn, statusCode string, statusPhrase string, messageBody ...string) {

	contentType := "text/plain"

	body := strings.Join(messageBody, "")
	body = body + "\n"
	contentLength := strconv.Itoa(len(body))

	// Server -> Client
	serverResponse := "HTTP/1.1 " + statusCode + " " + statusPhrase + "\r\n" +
		"Date: " + time.Now().UTC().Format(time.RFC1123) + "\r\n" +
		"Server: " + "GoLang NixOS TCP/HTTP Engine" + "\r\n" +
		"Content-Length: " + contentLength + "\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"Connection: " + "keep-alive" + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}

func HTTPFileServe(myConnection net.Conn, statusCode string, statusPhrase, filePath string) {
	var body string
	var contentType string

	body, contentType = fileReadingHelper(myConnection, filePath)

	contentLength := strconv.Itoa(len(body))

	// Server -> Client
	serverResponse := "HTTP/1.1 " + statusCode + " " + statusPhrase + "\r\n" +
		"Date: " + time.Now().UTC().Format(time.RFC1123) + "\r\n" +
		"Server: " + "GoLang NixOS TCP/HTTP Engine" + "\r\n" +
		"Content-Length: " + contentLength + "\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"Connection: " + "keep-alive" + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}

func fileReadingHelper(myConnection net.Conn, filePath string) (string, string) {
	sanitizedFilePath := filepath.Base(filePath)
	println("DEBUG: sanitizedFilePath: " + sanitizedFilePath)
	contentSlice := strings.SplitN(sanitizedFilePath, ".", 2)
	println("DEBUG: contentSlice: " + strings.Join(contentSlice, ", "))

	bodyBytes, err := os.ReadFile(sanitizedFilePath)
		if err != nil {
			log.Print(err.Error())
			MyHTTPMessage(myConnection, "404", "Not Found", "File not found")
			return "", ""
		}
	body := string(bodyBytes)
	contentType := "text/" + contentSlice[1]
	return body, contentType
}
