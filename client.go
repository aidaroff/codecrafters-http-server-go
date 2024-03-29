package main

import (
	"fmt"
	"net"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:4221")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send the HTTP request to the server
	request := "GET /user-agent HTTP/1.1\r\nHost: localhost\r\nUser-Agent: curl/7.64.1\r\n\r\n"
	_, err = conn.Write([]byte(request))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Receive the server's response
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(buffer))

	// Close the connection
	conn.Close()
}
