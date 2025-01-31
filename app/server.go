package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()
	readbuffer := make([]byte, 1024)
	n, err := conn.Read(readbuffer)
	if err != nil {
		fmt.Println("Failed to read data")
		os.Exit(1)
	}

	request := string(readbuffer[:n])
	fmt.Println("Received: ", request)

	// parse the request string "GET / HTTP/1.1\r\nUser-Agent: Go-http-client/1.1\r\n"
	requestLine := strings.Split(request, "\r\n")[0]
	requestLineParts := strings.Split(requestLine, " ")
	if len(requestLineParts) != 3 {
		fmt.Println("Invalid request")
		os.Exit(1)
	}
	requestMethod := requestLineParts[0]
	path := requestLineParts[1]
	fmt.Println("Path: ", path)
	if path == "/" {
		fmt.Println("Responding...")
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err != nil {
			fmt.Println("Failed to write data")
			os.Exit(1)
		}
	} else if strings.HasPrefix(path, "/echo/") { // STAGE 4
		pathParts := strings.Split(path, "/echo/")
		if len(pathParts) < 2 {
			fmt.Println("Invalid path")
			fmt.Println("Responding...")
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			os.Exit(1)
		}

		// for a request path such as /echo/abc/def/ -- everything after "/echo/" is the word
		word := strings.Join(pathParts[1:], "")
		contentLength := fmt.Sprintf("Content-Length: %d", len(word))
		headers := []string{"HTTP/1.1 200 OK", "Content-Type: text/plain", contentLength}
		response := strings.Join(headers, "\r\n") + "\r\n\r\n" + word
		fmt.Printf("Response: %s\n", response)
		fmt.Println("Responding...")
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Failed to write payload data")
			os.Exit(1)
		}
	} else if path == "/user-agent" {
		headersChunk := strings.Split(request, "\r\n\r\n")
		headers := strings.Split(headersChunk[0], "\r\n")
		for _, header := range headers {
			if strings.HasPrefix(header, "User-Agent:") {
				userAgent := strings.Split(header, " ")[1]
				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: text/plain\r\n\r\n%s", len(userAgent), userAgent)
				fmt.Println("Responding...")
				_, err = conn.Write([]byte(response))
				if err != nil {
					fmt.Println("Failed to write data")
					os.Exit(1)
				}
				break
			}
		}
	} else if strings.HasPrefix(path, "/files/") { // STAGE 7
		filename := strings.Split(path, "/files/")[1]
		fullFilename := directory + filename
		fmt.Println("Full filename: ", fullFilename)
		fmt.Println("Request method: ", requestMethod)
		if requestMethod == "POST" {
			// Parse request content
			requestBody := strings.Split(request, "\r\n\r\n")[1]
			if err := os.WriteFile(fullFilename, []byte(requestBody), 0644); err == nil {
				fmt.Println("File written")
				fmt.Println("Responding...")
				conn.Write([]byte("HTTP/1.1 201 OK\r\n\r\n"))
			} else {
				fmt.Println("Failed to write file")
				os.Exit(1)
			}

		} else if requestMethod == "GET" {
			if _, err := os.Stat(fullFilename); err == nil {
				content, err := os.ReadFile(fullFilename)
				if err != nil {
					fmt.Println("Failed to read file")
					os.Exit(1)
				}
				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: application/octet-stream\r\n\r\n%s", len(content), content)
				fmt.Println("Responding...")
				conn.Write([]byte(response))
			} else if os.IsNotExist(err) {
				fmt.Println("File not found")
				fmt.Println("Responding...")
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			} else {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Invalid method")
			fmt.Println("Responding...")
			conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
		}
	} else {
		fmt.Println("Invalid path")
		fmt.Println("Responding...")
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		os.Exit(1)
	}

	conn.Close()
}

func main() {
	// parse command line args
	var directory string
	flag.StringVar(&directory, "directory", "", "Directory to serve")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, directory)
	}
}
