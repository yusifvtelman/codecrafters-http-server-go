package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var directory *string

func main() {
	directory = flag.String("directory", ".", "Path to the directory")
	flag.Parse()
	logInfo("Directory passed as an argument: " + *directory)

	logInfo("Starting server...")
	startServer()
}

func startServer() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		logErrorAndExit("Failed to bind to port 4221", err)
	}
	defer l.Close()

	logInfo("Listening on 0.0.0.0:4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			logErrorAndExit("Error accepting connection", err)
		}
		logInfo("Accepted incoming connection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		logErrorAndExit("Error reading request data", err)
	}

	request := string(buffer[:n])
	logDebug("Received request:\n" + request)

	response := handleRequest(request)

	_, err = conn.Write([]byte(response))
	if err != nil {
		logErrorAndExit("Error sending response to client", err)
	}

	logInfo("Response sent successfully")
}

func handleRequest(request string) string {
	lines := strings.Split(request, "\r\n")
	if len(lines) == 0 {
		return "HTTP/1.1 400 Bad Request\r\n\r\n"
	}
	method := parseMethod(lines[0])
	logDebug("Parsed method: " + method)
	path := parsePath(lines[0])
	logDebug("Parsed path: " + path)

	switch {
	case path == "/":
		return "HTTP/1.1 200 OK\r\n\r\n"
	case strings.HasPrefix(path, "/echo/"):
		text := path[6:]
		logDebug("Echo text: " + text)
		return format200Response(text, "text/plain")
	case path == "/user-agent":
		userAgent := parseHeader(lines, "User-Agent:")
		logDebug("User-Agent: " + userAgent)
		return format200Response(userAgent, "text/plain")
	case strings.HasPrefix(path, "/files") && method == "GET":
		filename := path[6:]
		logDebug("[GET] File name passed: " + filename)
		file, err := getFile(filename)
		if err != nil {
			logDebug("Error finding a file")
			return format404Response()
		}
		return format200Response(file, "application/octet-stream")
	case strings.HasPrefix(path, "/files") && method == "POST":
		filename := path[6:]
		logDebug("[POST] File name passed: " + filename)
		data := lines[len(lines)-1]
		logDebug("Data to be written into file:" + data)
		writeFile(filename, data)
		return format201Response()
	default:
		return format404Response()
	}
}

func getFile(filename string) (string, error) {
	completeFilename := *directory + filename
	logDebug(completeFilename)

	body, err := os.ReadFile(completeFilename)

	if err == nil {
		logInfo("File: " + completeFilename + " is sent with Content: " + string(body))
	}

	return string(body), err
}

func writeFile(filename string, data string) {
	completeFilename := *directory + filename
	file, err := os.Create(completeFilename)
	file.WriteString(data)
	logInfo("File: " + completeFilename + "is created with Content: " + data)
	if err != nil {
		logErrorAndExit("Error writing a File: ", err)
	}
}

func parsePath(requestLine string) string {
	parts := strings.Split(requestLine, " ")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func parseMethod(requestLine string) string {
	parts := strings.Split(requestLine, " ")
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func parseHeader(lines []string, header string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, header) {
			return strings.TrimSpace(line[len(header):])
		}
	}
	return ""
}

func format404Response() string {
	return "HTTP/1.1 404 Not Found\r\n\r\n"
}

func format200Response(body string, contentType string) string {
	return fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s",
		contentType,
		len(body),
		body,
	)
}

func format201Response() string {
	return "HTTP/1.1 201 Created\r\n\r\n"
}

func logInfo(message string) {
	fmt.Println("[INFO] " + message)
}

func logDebug(message string) {
	fmt.Println("[DEBUG] " + message)
}

func logErrorAndExit(msg string, err error) {
	fmt.Println("[ERROR] " + msg + ": " + err.Error())
	os.Exit(1)
}
