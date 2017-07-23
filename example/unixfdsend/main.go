package main

import (
	"net"

	"github.com/shanemhansen/passunix"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/foo.sock")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	google, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		panic(err)
	}
	// make request in this program, read response in another program
	// magic!
	request := []byte("GET / HTTP/1.1\r\nHost: www.google.com\r\nConnection: close\r\n\r\n")
	if _, err := google.Write(request); err != nil {
		panic(err)
	}
	fd, err := google.(*net.TCPConn).File()
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	if err := passunix.Send(conn.(*net.UnixConn), request, fd); err != nil {
		panic(err)
	}
}
