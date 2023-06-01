package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
	"time"
)

//ClientEndpoint ...
type ClientEndpoint struct {
	LocalSocket  string
	RemoteSocket string
	TlsConfig    *tls.Config
	session      quic.Connection
	iface        *water.Interface
}

//Run ...
func (c *ClientEndpoint) Run() error {
	var err error
	c.session, err = quic.DialAddr(context.Background(), c.RemoteSocket, c.TlsConfig, &quic.Config{KeepAlivePeriod: time.Second * 10})
	if err != nil {
		fmt.Println("dial err", err)
		fmt.Println(err)
		return err
	}
	
	fmt.Println("dail success")

	go TunToQuic()
	
	return nil
}

//export TunToQuic
//TunToQuic 将tun的数据发送到quic session
func TunToQuic() error {

	//打开一个quic stream, 用于隧道连接
	stream, err := client.session.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	defer stream.Close()

	p1die := make(chan struct{})
	//将tun的数据发送到quic stream
	go func() {
		for {
			buf := make([]byte, 1500)
			n, err := client.iface.Read(buf)
			fmt.Println("recv from tun", n)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("write to  stream", n)
			stream.Write(buf[:n])
		}
	}()

	p2die := make(chan struct{})
	//在quic stream上等待数据返回，并将数据写入tun
	go func() {
		for {
			buf := make([]byte, 1500)
			n, err := stream.Read(buf)
			fmt.Println("recv from stream", n)
			if err != nil {
				fmt.Println("read err", err)
				continue
			}
			fmt.Println("write to tun", n)
			client.iface.Write(buf[:n])
		}
	}()

	//等待两个协程退出
	select {
	case <-p1die:
	case <-p2die:
	}
	return nil
}

//InitQuicConnect 初始化quic连接
func InitQuicConnect(localSocket string, remoteSocket string) (err error) {
	client = &ClientEndpoint{
		LocalSocket:  localSocket,
		RemoteSocket: remoteSocket,
		TlsConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-tun"},
		},
	}

	client.TlsConfig.ServerName = "172.23.76.84"

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun0"
	client.iface, err = water.New(config)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("new tun", client.iface.Name())

	/*
	cmerr := exec.Command("ifconfig", client.iface.Name(), "13.0.0.2", "netmask 255.255.255.255").Run()
	if cmerr != nil {
		fmt.Println("ifconfig", cmerr)
		return err
	}

	fmt.Println("ifconfig success")

	cmerr = exec.Command("ip", "route", "add", "13.0.0.1", "dev", client.iface.Name()).Run()
	if cmerr != nil {
		fmt.Println("ip route ", cmerr)
		return err
	}
	*/
	
	fmt.Println("client run")

	err = client.Run()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

var (
	client *ClientEndpoint
)
