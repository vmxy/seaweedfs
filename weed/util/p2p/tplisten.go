package p2p

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/hashicorp/yamux"
)

type TpListen struct {
	listen *net.Listener
	stream chan net.Conn
}

func NewTpListen(lister *net.Listener) TpListen {
	ut := TpListen{
		listen: lister,
		stream: make(chan net.Conn, 1),
	}
	ut.Init()
	return ut
}
func (lister *TpListen) Init() {
	for {
		conn, err := (*lister.listen).Accept()
		if err != nil {
			log.Fatal(err)
		}

		// 创建 yamux 会话
		session, err := yamux.Server(conn, nil)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			for {
				// 接收新的流
				stream, err := session.AcceptStream()
				if err != nil {
					log.Println("Stream accept error:", err)
					session.Close()
					return
				}
				lister.stream <- stream
			}
		}()
		//go handleSession(session, handle)
	}
}
func (lister *TpListen) Accept() (conn net.Conn, err error) {
	conn = <-lister.stream
	return
}
func (lister *TpListen) Close() error {
	err := (*lister.listen).Close()
	return err
}
func (lister *TpListen) Addr() net.Addr {
	addr := (*lister.listen).Addr()
	return addr
}

type DailManager struct {
	connects Map[string, *yamux.Session]
}

func NewDailManager() DailManager {
	con := DailManager{
		connects: NewMap[string, *yamux.Session](),
	}

	return con
}

var manager DailManager = NewDailManager()

func DailTp(network string, host string, port int) (conn net.Conn, err error) {
	conn, err = manager.Dail(network, host, port)
	return
}

func (manager *DailManager) Dail(network string, host string, port int) (conn net.Conn, err error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	session, find := manager.connects.Get(addr)
	if !find {
		conn, err = net.Dial(network, addr)
		if err != nil {
			return
		}
		session, err1 := yamux.Server(conn, nil)
		if err1 != nil {
			return conn, err1
		}
		manager.connects.Set(addr, session)
	}

	conn, err = session.Open()
	return
}

func (manager *DailManager) HttpRequest(host string, port int) (req *http.Request, err error) {
	conn, err := manager.Dail("tcp", host, port)

	if err != nil {
		return &http.Request{}, err
	}
	// 使用 bufio.Reader 来读取 conn 中的数据
	reader := bufio.NewReader(conn)
	// 从 reader 中解析 HTTP 请求
	req, err = http.ReadRequest(reader)
	return req, err
}
