package p2p

import (
	"fmt"
	"testing"
)

func TestAddr(t *testing.T) {

	// 将 Multiaddr 转换为 net.Addr
	netAddr, err := MultiaddrToNetAddr("/ip4/127.0.0.1/tcp/8080")
	fmt.Println("netAddr======>", netAddr, err)
}
