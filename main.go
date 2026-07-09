package main

import (
	"bufio"
	"fmt"
	_ "io"
	"net"
	"strconv"
	"strings"
	"time"
)

// Testing grounds for my implementation of solutions

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

			headers := make(map[string]string)

			for {
				headerLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				if headerLine == "\r\n" {
					println("End")
					break
				}
				result := strings.Split(headerLine, ": ")
				headers[result[0]] = result[1]
				headerLine = strings.TrimRight(headerLine, "\r\n")
				fmt.Printf("Read line: %q\n", headerLine)
			}

			if value, ok := headers["Content-Length"]; ok {
				fmt.Println("Found value:", value)
			} else {
				fmt.Println("Key not found")
				return
			}

			strValue := headers["Content-Length"]
			intValueContentLength, _ := strconv.Atoi(strValue)

			bufferHTTPBody := make([]byte, intValueContentLength)
			fmt.Printf("%q", bufferHTTPBody)

		}(conn)

	}
}
