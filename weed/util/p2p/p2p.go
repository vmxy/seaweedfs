package p2p

import (
	"context"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

type P2P struct {
	port   int
	ctx    context.Context
	host   *host.Host
	stream chan network.Stream
}

func NewP2P(port int) P2P {
	ctx := context.Background()
	/* 	host, err := libp2p.New(
	   	//libp2p.NoSecurity(cfg), // 禁用加密
	   	) */
	muBindUdpQuicWebtransport, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", port))
	muBindUdpQuicWebrtc, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webrtc-direct", port))
	host, err := libp2p.New(
		libp2p.NoSecurity,
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port),
		),
		libp2p.ListenAddrs(muBindUdpQuicWebtransport, muBindUdpQuicWebrtc),
		//libp2p.Identity(prvKey),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("host", host)
	for _, addr := range host.Addrs() {
		fmt.Println("addr===>", addr)
	}
	p2p := P2P{
		port: port,
		host: &host,
		ctx:  ctx,
	}
	host.SetStreamHandler("/asdf", func(s network.Stream) {
		p2p.stream <- s
	})
	return p2p
}
func (p2p P2P) Peer() string {
	return (*p2p.host).ID().String()
}
func (p2p P2P) Port() int {
	return p2p.port
}
func (p2p *P2P) Accept() (conn net.Conn, err error) {
	stream := <-p2p.stream
	p2pConn := P2PConn{
		stream: stream,
	}
	conn = &p2pConn
	return
}

/* func (p2p *P2P) Connect(addr string) (conn net.Conn, err error) {
	host := *p2p.host

	stream, err := host.NewStream(p2p.ctx, addr)
	p2pConn := P2PConn{
		stream: stream,
	}
	conn = &p2pConn

	return
} */
