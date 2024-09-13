package p2p

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestP2p(t *testing.T) {
	/*
		 	peer := NewP2P(P2POptions{
				Port: 8383,
			})
	*/
	peer1 := NewP2P(P2POptions{
		Port: 8383,
	})
	time.Sleep(1 * time.Second)
	peer2 := NewP2P(P2POptions{
		Port:       8484,
		ServerPort: 9001,
	})
	time.Sleep(1 * time.Second)

	peer2.HandleHttp(func(req http.Request) {
		fmt.Println("request", req.URL)
	})
	if len(peer2.GetPeerAddr()) < 1 {

		return
	}
	p2id := peer2.GetPeerAddr()[0]
	fmt.Println("p2id=", p2id)
	resp, err := peer1.OpenHttpRequest(p2id, "GET", "http://192.168.66.7:9001/4,033e1678de6a", nil)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("body", resp.StatusCode, resp.Header.Get("content-length"))
	bytes := make([]byte, 2048)
	for {
		size, err := resp.Body.Read(bytes)
		fmt.Println("-------------------->client", size, err)
		if err != nil {
			break
		}
	}
	time.Sleep(60 * time.Second)
}
