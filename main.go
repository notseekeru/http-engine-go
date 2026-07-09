package main

import (
	"bufio"
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

			for {
				headerLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				println(headerLine)
			}

		}(conn)

	}
}
