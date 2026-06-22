package main

import (
	"net"
	"time"
)

func main() {
	mylisterer, err := net.Listen("tcp", ":8080")

	if err != nil {
		panic(err)
	}

	defer mylisterer.Close()

	println("Listening to 8080")

	for {
		conn, err := mylisterer.Accept()
		if err != nil {
			println(err)
			continue
		}

		go func(myconnection net.Conn) {
			defer myconnection.Close()
			myconnection.SetDeadline(time.Now().Add(5 * time.Second)) // Sets deadline

			myresponse := "HTTP/1.1 200 OK\r\n" +
				"Content-Length: 22\r\n" +
				"\r\n" +
				"Hello from Go server!\n"

			myconnection.Write([]byte(myresponse))
		}(conn)
	}
}
