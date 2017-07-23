package passunix

import (
	"errors"
	"net"
	"os"
	"syscall"
)

type filer interface {
	File() (*os.File, error)
}

// Send sends files to server on other side of unix conn. extra []byte is
// typically something you read from file[0], but it doesn't have to be. The
// Accept() function in this package will automagically create a conn that
// returns extra before reading file.
func Send(fileTo *net.UnixConn, extra []byte, file *os.File) error {
	fd := int(file.Fd())
	_, _, err := fileTo.WriteMsgUnix(extra, syscall.UnixRights(fd), nil)
	if err != nil {
		return err
	}
	return nil
}

// Accept gets a new fd off of c and returns the info.
// it has a few hard coded values. Beware if you are sending
// a giant http request with several k of cookies.
func Accept(c *net.UnixConn) ([]byte, []*os.File, error) {
	// no idea how big these need to be.
	b := make([]byte, 2024)
	oob := make([]byte, 2024)
	// TODO read b to get
	n, oobn, _, _, err := c.ReadMsgUnix(b, oob)
	if err != nil {
		return nil, nil, err
	}
	cmsgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, nil, err
	}
	if len(cmsgs) < 1 {
		return nil, nil, errors.New("need at least one cmsg")
	}
	fds, err := syscall.ParseUnixRights(&cmsgs[0])
	if err != nil {
		return nil, nil, err
	}
	files := make([]*os.File, len(fds))
	for i, fd := range fds {
		files[i] = os.NewFile(uintptr(fd), "receivedconn")
	}
	return b[:n], files, nil
}

// MakeConn will turn fd into a net.Conn. fd will be closed
// and a new net.Conn will rise from the ashes.
func MakeConn(fd *os.File) (net.Conn, error) {
	fileconn, err := net.FileConn(fd)
	if err != nil {
		return nil, err
	}
	fd.Close()
	return fileconn, nil
}
