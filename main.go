package main

import (
	"flag"
	"fmt"
)

func usage() {
		fmt.Println(`
usage:
	quic-tunnel -m client -r xxx.xxx.xxx.xxx:60001 	//客户端运行模式
	quic-tunnel -m server -l 0.0.0.0:60001 	//服务端运行模式
		`)
}

func main() {

	remote := flag.String("r", "", "remote server IP address")
    mode := flag.String("m", "client", "run mode")
    local := flag.String("l", "", "run mode")

	flag.Parse()
	
	if *mode == "client" && *remote == "" {
		usage()
		return 
	} else if *mode == "server" && *local == "" {
		usage()
		return
	}

	if *mode == "client" {
		//客户端
		InitQuicConnect("0.0.0.0:60001", *remote)
	} else {
		//服务端
		InitQuicServer(*local, "0.0.0.0:60001")
	}

	select {}
}
