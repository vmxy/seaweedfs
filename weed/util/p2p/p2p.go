package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

type P2P struct {
	port       int
	serverPort int
	ctx        context.Context
	host       *host.Host
	stream     chan network.Stream
	acceptHttp chan http.Request
}
type P2POptions struct {
	Port       int
	ServerPort int
	Master     bool
}

const ProtocolId = "/weed-node"

func NewP2P(options P2POptions) P2P {
	port := options.Port
	isMaster := options.Master
	ctx := context.Background()
	/* 	host, err := libp2p.New(
	   	//libp2p.NoSecurity(cfg), // 禁用加密
	   	) */
	//muBindUdpQuicWebtransport, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", port))
	//muBindUdpQuicWebrtc, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webrtc-direct", port))
	var enrelay libp2p.Option
	if isMaster {
		enrelay = libp2p.EnableRelay()
	}
	host, err := libp2p.New(
		libp2p.NoSecurity,
		enrelay,
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port),
		),
		//libp2p.ListenAddrs(muBindUdpQuicWebtransport, muBindUdpQuicWebrtc),
		//libp2p.Identity(prvKey),
	)
	if err != nil {
		panic(err)
	}
	p2p := P2P{
		port:       port,
		host:       &host,
		ctx:        ctx,
		serverPort: options.ServerPort,
		stream:     make(chan network.Stream, 1),
		acceptHttp: make(chan http.Request),
	}
	host.SetStreamHandler(ProtocolId, func(s network.Stream) {
		p2p.stream <- s
	})
	go p2p.handle()
	return p2p
}
func (p2p P2P) ID() string {
	return (*p2p.host).ID().String()
}
func (p2p P2P) Open() bool {
	return true
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

func (p2p *P2P) HandleHttp(handle func(request http.Request)) {
	go func() {
		for {
			accept := <-p2p.acceptHttp
			go handle(accept)
		}
	}()
}

func (p2p *P2P) handle() {
	for {
		conn, err := p2p.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}
		//reader := bufio.NewReader(conn)
		//writer := bufio.NewWriter(conn)
		//rw := bufio.NewReadWriter(reader, writer)
		readBytes := make([]byte, 2048)
		size, err := conn.Read(readBytes)
		if err != nil || size < 1 {
			conn.Close()
			continue
		}
		readBytes = readBytes[0:size]

		if isHttp(readBytes) {
			p2p.handleHttp(readBytes, conn)
		}
	}
}
func isHttp(bytes []byte) bool {
	if len(bytes) < 9 {
		return false
	}
	input := string(bytes[0:7])
	vs := strings.Split(input, " ")
	if len(vs) < 2 {
		return false
	}
	reg := `^(GET|HEAD|PUT|POST|DELETE|OPTIONS|TRACE|CONNECT)`
	r := regexp.MustCompile(reg)
	return r.MatchString(input)
}
func (p2p *P2P) handleHttp(bytes []byte, local net.Conn) {
	reader := bufio.NewReader(strings.NewReader(string(bytes)))
	var headers map[string]string = make(map[string]string)
	var list []string = make([]string, 0)
	reHost := regexp.MustCompile("^(?i)Host$")
	var host string = fmt.Sprintf("127.0.0.1:%d", p2p.serverPort)
	for i := 0; true; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		if line == "\r\n" {
			break
		}
		line = strings.TrimSpace(line)
		if i > 0 {
			kv := strings.Split(line, ": ")
			if reHost.MatchString(kv[0]) {
				hp := strings.Split(kv[1], ":")
				if len(hp) < 2 {
					kv[1] = kv[1] + ":80"
				}
				host = kv[1]
			}

			headers[kv[0]] = kv[1]
		}
		list = append(list, line)
	}

	bytes = []byte(strings.Join(list, "\r\n") + "\r\n\r\n")
	conn, err := net.Dial("tcp", host)
	if err != nil {
		local.Close()
		return
	}
	go func() {
		bytes := make([]byte, 2048)
		for {
			size, err := local.Read(bytes)
			if err != nil {
				conn.Close()
				break
			}
			conn.Write(bytes[0:size])
		}
	}()
	go func() {
		for {
			bytes := make([]byte, 2048)
			size, err := conn.Read(bytes)
			if err != nil {
				local.Close()
				if err == io.EOF {
					break
				}
				conn.Close()
				break
			}
			local.Write(bytes[0:size])
		}
	}()
	conn.Write(bytes)
}
func (p2p *P2P) Connect(peerAddr string) {
	host := *p2p.host
	maddr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		log.Println(err)
		return
	}
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return
	}
	err = host.Connect(p2p.ctx, *info)
	if err != nil {
		log.Println("p2p connect error", peerAddr, err)
	}
}

// CustomTransport 实现了 http.RoundTripper 接口
type CustomTransport struct {
	Conn              net.Conn
	DisableKeepAlives bool
}

// RoundTrip 实现了 http.RoundTripper 接口
func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 将请求写入自定义的 net.Conn
	if err := req.Write(t.Conn); err != nil {
		return nil, err
	}

	// 读取响应
	resp, err := http.ReadResponse(bufio.NewReader(t.Conn), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
func NewHttpRequest(method string, url string, body io.Reader, conn net.Conn) (resp *http.Response, err error) {
	// 创建自定义的 RoundTripper
	transport := &CustomTransport{
		Conn:              conn,
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return
	}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return
	}
	fmt.Println("Response Status:", resp.Status)
	return resp, err
}
func (p2p *P2P) OpenHttpRequest(peerAddr string, method string, url string, body io.Reader) (resp *http.Response, err error) {
	conn, err := p2p.OpenStream(peerAddr)
	if err != nil {
		return
	}
	resp, err = NewHttpRequest(method, url, body, conn)
	return
}
func (p2p *P2P) OpenStream(peerAddr string) (conn net.Conn, err error) {
	host := *p2p.host
	maddr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// Start a stream with the destination.
	// Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
	stream, err := host.NewStream(context.Background(), info.ID, ProtocolId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	p2pConn := P2PConn{
		stream: stream,
	}
	conn = &p2pConn
	// Create a buffered stream so that read and writes are non-blocking.
	//rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	return
}
func (p2p *P2P) GetPeerAddr() []string {
	var addrs []string
	host := *p2p.host
	for _, a := range host.Addrs() {
		line := a.String()
		if !strings.Contains(line, "/127.0.0.1") || strings.Contains(line, "/web") {
			continue
		}
		/* 	if strings.Contains(line, "quic-v1") {

		} */
		addr := fmt.Sprintf("%s/p2p/%s", line, host.ID())
		addrs = append(addrs, addr)
	}
	return addrs
}

/*
	 func (p2p *P2P) Connect(addr string) (conn net.Conn, err error) {
		host := *p2p.host

		stream, err := host.NewStream(p2p.ctx, addr)
		p2pConn := P2PConn{
			stream: stream,
		}
		conn = &p2pConn

		return
	}
*/
type MultiaddrStrings []multiaddr.Multiaddr

// 实现 sort.Interface
func (m MultiaddrStrings) Len() int      { return len(m) }
func (m MultiaddrStrings) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m MultiaddrStrings) Less(i, j int) bool {
	/*
		 	line := m[i].String()
			if strings.Contains(line, "/127.0.0.1") {
				return false
			}
			if strings.Contains(line, "/quic-v1") {
				sj := m[j].String()
				if strings.Contains(sj, "/quic-v1") {
					return len(line) < len(sj)
				}
				return true
			}
			return m[i].String() < m[j].String()
	*/
	return true
}
