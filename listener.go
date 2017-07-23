package passunix

import (
	"bytes"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
	"strings"
)

// Listen returns a Listener that returns net.Conn's after Accept() is called
// addr should be a filesystem path. There's some cleverness in the API,
// it assumes that the fd has been read a little already and it prepends
// those bytes. For example if you are sniffing a connection before
// passing the fd
func Listen(addr string) (net.Listener, error) {
	listener, err := net.Listen("unix", addr)
	if err != nil {
		panic(err)
	}
	ul := listener.(*net.UnixListener)
	connChan := make(chan net.Conn)
	go func() {
		for {
			conn, err := ul.AcceptUnix()
			if err != nil {
				panic(err)
			}
			go func() {
				if err := handleConn(conn, connChan); err != nil {
					log.Printf("Error from handleConn %s", err)
				}
			}()
		}
	}()
	return &chanListener{
		connChan: connChan,
		addr:     ul.Addr(),
		inner:    ul,
	}, nil
}

type chanListener struct {
	addr     net.Addr
	connChan chan net.Conn
	inner    net.Listener
}

type readerConn struct {
	net.Conn
	multiReader io.Reader
}

func (this *readerConn) Read(buf []byte) (int, error) {
	return this.multiReader.Read(buf)
}

func (this *chanListener) Close() error {
	this.inner.Close()
	close(this.connChan)

	return nil
}
func (this *chanListener) Addr() net.Addr {
	return this.addr
}

var ErrChanClosed = errors.New("can't accept on channel, closed")

func (this *chanListener) Accept() (net.Conn, error) {
	conn, ok := <-this.connChan
	if ok {
		return conn, nil
	}
	return nil, ErrChanClosed
}

// handleConn will accept new conns until read returns EOF.
func handleConn(conn *net.UnixConn, connChan chan net.Conn) error {
	defer conn.Close()
	for {
		data, fds, err := Accept(conn)
		if err != nil {
			// ok, so we might have an EOF, but it's wrapped in a net.OpError
			// we should probably file a bug against this.
			// looking for: "read unix %s->@: EOF"
			if strings.Contains(err.Error(), "EOF") {
				return nil
			}
			// otherwise just a normal error.
			return errors.Wrapf(err, "unable to accept new conn %+T", err)
		}
		if len(fds) > 1 {
			log.Printf("Got %d fds, we don't know how to handle that because we prepend the data to the first fd.", len(fds))
		}
		conn2, err := MakeConn(fds[0])
		if err != nil {
			return errors.Wrap(err, "unable to turn conn into con")
		}
		rcon := &readerConn{Conn: conn2, multiReader: io.MultiReader(bytes.NewReader(data), conn2)}
		connChan <- rcon
	}
}
