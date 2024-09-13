package p2p

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

type P2PConn struct {
	stream network.Stream
}

func (sc *P2PConn) Read(b []byte) (n int, err error) {
	return sc.stream.Read(b)
}

func (sc *P2PConn) Write(b []byte) (n int, err error) {
	return sc.stream.Write(b)
}

func (sc *P2PConn) Close() error {
	return sc.stream.Close()
}

func (sc *P2PConn) LocalAddr() net.Addr {
	addr, err := MultiaddrToNetAddr(sc.stream.Conn().LocalMultiaddr().String())
	if err != nil {
		return nil
	}
	return addr
}

func (sc *P2PConn) RemoteAddr() net.Addr {
	addr, err := MultiaddrToNetAddr(sc.stream.Conn().RemoteMultiaddr().String())
	if err != nil {
		return nil
	}
	return addr
}

func (sc *P2PConn) SetDeadline(t time.Time) error {
	return sc.stream.SetDeadline(t)
}

func (sc *P2PConn) SetReadDeadline(t time.Time) error {
	return sc.stream.SetReadDeadline(t)
}

func (sc *P2PConn) SetWriteDeadline(t time.Time) error {
	return sc.stream.SetWriteDeadline(t)
}

func isIpv4(host string) bool {
	iptype := net.ParseIP(host)
	return iptype != nil && iptype.To4() != nil
}
func isIpv6(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && ip.To16() != nil && ip.To4() == nil
}
func isDomain(host string) bool {
	re := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+(?:[a-zA-Z]{2,})$`)
	return re.MatchString(host)
}

// MultiaddrToNetAddr 将 Multiaddr 转换为 net.Addr
func MultiaddrToNetAddr(maddr string) (addr net.Addr, err error) {
	// 解析 Multiaddr，提取 IP 和端口信息
	addrInfos := strings.Split(maddr, "/")
	var ip string = ""
	var port int
	for _, addr := range addrInfos {
		if isIpv4(addr) || isIpv6(addr) || isDomain(addr) {
			ip = addr
			continue
		}
		v, err := strconv.Atoi(addr) // 尝试将字符串转换为整数
		if err == nil {
			port = v
			break
		}
	}
	if ip == "" {
		err = errors.New("no port")
		return
	}
	if port < 1 {
		err = errors.New("no port")
		return
	}
	addr = &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	return
}
