package tcp

import (
	"fmt"
	"github.com/net-byte/water"
	"github.com/patrickmn/go-cache"
	"net"
	"sync/atomic"
	"vpn/log"
	"vpn/model"
)

type ClientNetWork struct {
	Cos       *model.CreateOptions
	connCache *cache.Cache
	UdpConn   *net.UDPConn
	tun       *water.Interface
	TcpConn   net.Conn
}

var (
	ClientSentBytes      uint64
	ClientReceivingBytes uint64
)

func NewClientNetWork(config *model.CreateOptions, Tun *water.Interface) *ClientNetWork {
	return &ClientNetWork{
		Cos: config,
		tun: Tun,
	}
}

func (cnw *ClientNetWork) ClientDial() {
	conn, err := net.Dial("tcp", cnw.Cos.RemoteServerIP+":"+cnw.Cos.ListenPort)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Warn(fmt.Sprintf("Establish a connection from: %s", conn.RemoteAddr()))
	defer conn.Close()
	cnw.TcpConn = conn
	go cnw.readTCPNetworkToTUN()
	cnw.readTunToTCPNetwork()
}

func (cnw *ClientNetWork) readTCPNetworkToTUN() {
	buf := make([]byte, cnw.Cos.BufferSize)
	for {
		n, err := cnw.TcpConn.Read(buf)
		if err != nil {
			log.Warn(fmt.Sprintf("client read err : %v", err))
			return
		}
		b := buf[:n]
		cnw.tun.Write(b)
		cnw.receivingBytes(n)
	}
}

func (cnw *ClientNetWork) readTunToTCPNetwork() {
	buf := make([]byte, cnw.Cos.BufferSize)
	for {
		n, err := cnw.tun.Read(buf)
		if err != nil || n == 0 {
			log.Warn("client read err : ", err)
			continue
		}
		b := buf[:n]
		cnw.TcpConn.Write(b)
		cnw.setSentBytes(n)
	}
}

func (cnw *ClientNetWork) GetSend() uint64 {
	return ClientSentBytes
}

func (cnw *ClientNetWork) GetReceiving() uint64 {
	return ClientReceivingBytes
}

func (cnw *ClientNetWork) setSentBytes(n int) {
	atomic.AddUint64(&ClientSentBytes, uint64(n))
}

func (cnw *ClientNetWork) receivingBytes(n int) {
	atomic.AddUint64(&ClientReceivingBytes, uint64(n))
}
