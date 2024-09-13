package p2p

import (
	"net"

	"github.com/quic-go/quic-go"
)

/*
	Read(b []byte) (n int, err error)

Write(b []byte) (n int, err error)
Close() error
LocalAddr() Addr
RemoteAddr() Addr
SetDeadline(t time.Time) error
SetReadDeadline(t time.Time) error
SetWriteDeadline(t time.Time) error
*/
type TpConn struct {
	stream      *net.Conn
	quickStream *quic.Stream
}

func (tp *TpConn) Read(b []byte) (n int, err error) {
	stream := *tp.quickStream
	return stream.Read(b)
}
func (tp *TpConn) Write(b []byte) (n int, err error) {
	stream := *tp.quickStream

	return stream.Write(b)
}
func (tp *TpConn) Close() error {
	stream := *tp.quickStream
	return stream.Close()
}
