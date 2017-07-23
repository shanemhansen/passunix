package main

import (
	"github.com/pkg/errors"
	"github.com/shanemhansen/passunix"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	os.Remove("/tmp/foo.sock")
	l, err := net.Listen("unix", "/tmp/foo.sock")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	ul := l.(*net.UnixListener)
	for {
		conn, err := ul.AcceptUnix()
		if err != nil {
			panic(err)
		}
		if err := handleConn(conn); err != nil {
			log.Println(err)
		}
	}

}

func handleConn(conn *net.UnixConn) error {
	defer conn.Close()
	for {
		_, fds, err := passunix.Accept(conn)
		if err != nil {
			// ok, so we might have an EOF, but it's wrapped in a net.OpError
			// we should probably file a bug against this.
			// looking for: "read unix %s->@: EOF"
			if strings.Contains(err.Error(), "EOF") {
				return nil
			}
			// otherwise just a normal error.
			return errors.Wrapf(err, "unable to accept new conn")
		}
		conn2, err := passunix.MakeConn(fds[0])
		if err != nil {
			return err
		}
		func() {
			defer conn2.Close()
			// read data
			io.Copy(os.Stdout, conn2)
		}()
	}
}
