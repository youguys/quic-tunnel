package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
	"time"
)

//ServerEndpoint ...
type ServerEndpoint struct {
	Address   string
	TLSConfig *tls.Config
	listener  *quic.Listener
	iface     water.Interface
}

//Run ...
func (s *ServerEndpoint) Run() error {
	var err error

	s.listener, err = quic.ListenAddr(s.Address, s.TLSConfig, &quic.Config{KeepAlivePeriod: time.Second * 10})
	if err != nil {
        fmt.Println("listen err", err)
		return err
	}


    fmt.Println("listen success")

	for {

        fmt.Println("begin to accept")
		session, err := s.listener.Accept(context.Background())
		if err != nil {
            fmt.Println("accept err", err)
			return err
		}
    
        fmt.Println("accept success")

		go func() {
			//打开一个quic stream, 用于隧道连接
			stream, err := session.AcceptStream(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}

			defer stream.Close()

			p1die := make(chan struct{})
			//将tun的数据发送到quic stream
			go func() {
				for {
					buf := make([]byte, 1500)
					n, err := server.iface.Read(buf)
                    fmt.Println("recv from tun", n)
					if err != nil {
						fmt.Println(err)
						continue
					}

                    fmt.Println("write to stream", n)
					stream.Write(buf[:n])
				}
			}()

			p2die := make(chan struct{})
			//在quic stream上等待数据返回，并将数据写入tun
			go func() {
				for {
					buf := make([]byte, 1500)
					n, err := stream.Read(buf)
					if err != nil {
						fmt.Println(err)
						continue
					}
                    
                    fmt.Println("stram read len", n)
                    fmt.Println(server)
                    
                    fmt.Println("write to tun", n)
					server.iface.Write(buf[:n])
				}
			}()

			//等待两个协程退出
			select {
			case <-p1die:
			case <-p2die:
			}
		}()
	}

	return nil
}

//export InitQuicServer
//InitQuicServer 初始化quic连接
func InitQuicServer(localSocket string, remoteSocket string) (err error) {

	server = &ServerEndpoint{
		Address: localSocket,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
            NextProtos:         []string{"quic-tun"},
		},
	}
    
    tlsCert, err := tls.LoadX509KeyPair("./server.crt", "./server.key")
    if err != nil {
        fmt.Println("load cert err", err)
        return err
    }
    server.TLSConfig.Certificates = []tls.Certificate{tlsCert}


	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun0"

    var iface *water.Interface
	iface, err = water.New(config)
    server.iface = *iface
	if err != nil {
		return err
	}

    /*
	out, cmerr := exec.Command("ifconfig", server.iface.Name(), "13.0.0.1", "netmask 25.255.255.255").Output()
    if cmerr != nil {
        fmt.Println(cmerr,out)
		return cmerr
    }
    
    out, cmerr = exec.Command("ip", "route", "add", "13.0.0.1", "dev", server.iface.Name()).Output()
    if cmerr != nil {
        fmt.Println(cmerr,out)
		return cmerr
    }
    */

    fmt.Println("server run")
	err = server.Run()
	if err != nil {
		return err
	}

	return nil
}

var (
	server *ServerEndpoint
)
