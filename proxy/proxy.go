package proxy

import (
	"log"
	"net"

	"github.com/juliotorresmoreno/turn/manager"
)

type Proxy struct {
	listener net.Listener
}

func SetupProxy(addr string) *Proxy {
	proxy := Proxy{}
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	proxy.listener = listener

	return &proxy
}

func (proxy *Proxy) Listen(manager *manager.Manager) {
	for {
		conn, err := proxy.listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		manager.Pipe <- conn
	}
}
