package main

import (
	"config"
	"net"
	"os"
	"runtime"
	"socket"
	"sslog"
)

func main() {

	//加载数据库会议相关数据信息到内存中
	if err := config.LoadGlobaldata(); err != nil {
		os.Exit(1)
	}

	go FlashSocket()
	WebSocket()
}

func FlashSocket() {
	service := ":8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		sslog.LoggerErr("ResolveTCPAddr: %s", err.Error())
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		sslog.LoggerErr("ListenTCP: %s", err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerDebug("[%s:%d] LocalPort %d accept  a client comming %s", file, line+1, tcpAddr.Port, conn.RemoteAddr().String())
		session := socket.NewTcpSession(tcpAddr.Port, conn)
		go socket.ProcessTcpSession(session)
	}
}

func WebSocket() {
	service := ":80"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		sslog.LoggerErr("ResolveTCPAddr: %s", err.Error())
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		sslog.LoggerErr("ListenTCP: %s", err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		_, file, line, _ := runtime.Caller(0)
		sslog.LoggerDebug("[%s:%d] LocalPort %d accept  a client comming %s", file, line+1, tcpAddr.Port, conn.RemoteAddr().String())
		session := socket.NewTcpSession(tcpAddr.Port, conn)
		go socket.ProcessTcpSession(session)
	}

}
