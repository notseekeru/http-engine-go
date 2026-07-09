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
				headerLine = strings.TrimRight(headerLine, "\r\n")
				if headerLine == "" {
					println("End")
					break
				}
				result := strings.Split(headerLine, ": ")
				headers[result[0]] = result[1]
				fmt.Printf("Read line: %q\n", headerLine)
			}

			if value, ok := headers["Content-Length"]; ok {
				fmt.Println("Found value:", value)
			} else {
				fmt.Println("Key not found")
				return
			}

			strValue := headers["Content-Length"]
			intValue64, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				fmt.Printf("PARSING ERROR: Could not convert string %q to int: %v\n", strValue, err)
			}
			nextByte, _ := reader.Peek(1)
			fmt.Printf("Cursor is currently resting on character: %q\n", string(nextByte))
			bodyReader := io.LimitReader(reader, intValue64)

			bodyBytes, err := io.ReadAll(bodyReader)
			if err != nil {
				fmt.Printf("CRITICAL ERROR DURING READ: %v\n", err)
			}
			fmt.Printf("Body payload read via io: %s\n", string(bodyBytes))

		}(conn)

	}
}
