package udp

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
	Cos        *model.CreateOptions
	connCache  *cache.Cache
	UdpConn    *net.UDPConn
	Net        *water.Interface
	serverAddr *net.UDPAddr
}

func NewClientNetWork(config *model.CreateOptions, Net *water.Interface) *ClientNetWork {
	snw := &ClientNetWork{
		Cos: config,
		Net: Net,
	}
	return snw
}

func (cnw *ClientNetWork) ClientDial() {
	serverAddr, err := net.ResolveUDPAddr("udp", cnw.Cos.RemoteServerIP+":"+cnw.Cos.ListenPort)
	if err != nil {
		log.Error(err.Error())
		return
	}
	cnw.serverAddr = serverAddr
	localAddr, err := net.ResolveUDPAddr("udp", ":"+"3000")
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		panic(err)
	}
	cnw.UdpConn = conn
	defer conn.Close()
	go cnw.readUDPNetworkToTUN()
	cnw.readTunToUDPNetwork()
}

func (cnw *ClientNetWork) readUDPNetworkToTUN() {
	buf := make([]byte, 2000)
	for {
		n, _, err := cnw.UdpConn.ReadFromUDP(buf)
		if err != nil || n == 0 {
			continue
		}
		b := buf[:n]
		cnw.Net.Write(b)
		cnw.setSentBytes(n)
	}
}

func (cnw *ClientNetWork) readTunToUDPNetwork() {
	buf := make([]byte, 2000)
	for {
		n, err := cnw.Net.Read(buf)
		if err != nil || n == 0 {
			fmt.Println("err ", err)
			continue
		}
		b := buf[:n]
		cnw.UdpConn.WriteToUDP(b, cnw.serverAddr)
		cnw.receivingBytes(n)
	}
}

var (
	ClientSentBytes      uint64
	ClientReceivingBytes uint64
)

func (cnw *ClientNetWork) setSentBytes(n int) {
	atomic.AddUint64(&ClientSentBytes, uint64(n))
}

func (cnw *ClientNetWork) receivingBytes(n int) {
	atomic.AddUint64(&ClientReceivingBytes, uint64(n))
}
