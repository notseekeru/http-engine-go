package main

import (
	"fmt"
	"net"
)

func main() {
	mylisterer, err := net.Listen("tcp", ":8080")

	if err != nil {
		panic(err)
	}

	defer mylisterer.Close()

	fmt.Println("Listening to 8080")

	for {
		conn, err := mylisterer.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func(myconnection net.Conn) {
			defer myconnection.Close()

			response := "HTTP/1.1 200 OK\r\n" +
				"Content-Length: 22\r\n" +
				"\r\n" +
				"Hello from Go server!\n"

			myconnection.Write([]byte(response))
		}(conn)
	}
}
