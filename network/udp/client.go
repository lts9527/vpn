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
	tun        *water.Interface
	serverAddr *net.UDPAddr
}

func NewClientNetWork(config *model.CreateOptions, Tun *water.Interface) *ClientNetWork {
	return &ClientNetWork{
		Cos: config,
		tun: Tun,
	}
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
		cnw.tun.Write(b)
		cnw.setSentBytes(n)
	}
}

func (cnw *ClientNetWork) readTunToUDPNetwork() {
	buf := make([]byte, 2000)
	for {
		n, err := cnw.tun.Read(buf)
		if err != nil || n == 0 {
			fmt.Println("err ", err)
			continue
		}
		b := buf[:n]
		cnw.UdpConn.WriteToUDP(b, cnw.serverAddr)
		cnw.setReceivingBytes(n)
	}
}

func (cnw *ClientNetWork) GetSend() uint64 {
	return ClientSentBytes
}

func (cnw *ClientNetWork) GetReceiving() uint64 {
	return ClientReceivingBytes
}

var (
	ClientSentBytes      uint64
	ClientReceivingBytes uint64
)

func (cnw *ClientNetWork) setSentBytes(n int) {
	atomic.AddUint64(&ClientSentBytes, uint64(n))
}

func (cnw *ClientNetWork) setReceivingBytes(n int) {
	atomic.AddUint64(&ClientReceivingBytes, uint64(n))
}
