package main

import (
	"net"
	"strconv"
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
			break
		}

		go func(myconnection net.Conn) {
			defer myconnection.Close()
			myconnection.SetDeadline(time.Now().Add(5 * time.Second)) // Sets deadline

			mybody := "Hello from go Server!\n"
			mybodylength := strconv.Itoa(len(mybody))

			myresponse := "HTTP/1.1 200 OK\r\n" +
				"Content-Length: " + mybodylength + "\r\n" +
				"\r\n" +
				mybody

			myconnection.Write([]byte(myresponse))
		}(conn)
	}
}
