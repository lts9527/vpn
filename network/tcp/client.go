package tcp

import (
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
	Net       *water.Interface
	TcpConn   net.Conn
}

var (
	ClientSentBytes      uint64
	ClientReceivingBytes uint64
)

func NewClientNetWork(config *model.CreateOptions, Net *water.Interface) *ClientNetWork {
	snw := &ClientNetWork{
		Cos: config,
		Net: Net,
	}
	return snw
}

func (cnw *ClientNetWork) ClientDial() {
	serverConn, err := net.Dial("tcp", cnw.Cos.RemoteServerIP)
	if err != nil {
		panic(err)
	}
	defer serverConn.Close()
	go cnw.readTCPNetworkToTUN()
	go cnw.readTunToTCPNetwork()
}

func (cnw *ClientNetWork) readTCPNetworkToTUN() {
	buf := make([]byte, 2000)
	for {
		n, err := cnw.TcpConn.Read(buf)
		if err != nil {
			log.Warn("client read err : ", err)
			return
		}
		b := buf[:n]
		cnw.Net.Write(b)
		cnw.receivingBytes(n)
	}
}

func (cnw *ClientNetWork) readTunToTCPNetwork() {
	buf := make([]byte, 2000)
	for {
		n, err := cnw.Net.Read(buf)
		if err != nil || n == 0 {
			log.Warn("client read err : ", err)
			continue
		}
		b := buf[:n]
		_, err = cnw.TcpConn.Write(b)
		if err != nil {
			log.Warn("client write err :", err)
			continue
		}
		cnw.setSentBytes(n)
	}
}

func (cnw *ClientNetWork) setSentBytes(n int) {
	atomic.AddUint64(&ClientSentBytes, uint64(n))
}

func (cnw *ClientNetWork) receivingBytes(n int) {
	atomic.AddUint64(&ClientReceivingBytes, uint64(n))
}
