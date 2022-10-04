package tcp

import (
	"fmt"
	"github.com/net-byte/vtun/common/netutil"
	"github.com/net-byte/water"
	"github.com/patrickmn/go-cache"
	"net"
	"sync/atomic"
	"vpn/log"
	"vpn/model"
)

type ServerNetWork struct {
	Cos       *model.CreateOptions
	connCache *cache.Cache
	UdpConn   *net.UDPConn
	TcpConn   net.Conn
	Net       *water.Interface
}

var (
	ServerSentBytes      uint64
	ServerReceivingBytes uint64
)

func NewServerNetWork(config *model.CreateOptions, Net *water.Interface) *ServerNetWork {
	snw := &ServerNetWork{
		Cos: config,
		Net: Net,
	}
	return snw
}

func (snw *ServerNetWork) ListenTCP() {
	ListenAddr, err := net.ResolveTCPAddr("tcp", snw.Cos.ServerAddress)
	if err != nil {
		log.Error("Failed to get tcp socket:", err)
		panic(err)
	}
	conn, err := net.ListenTCP("tcp", ListenAddr)
	if err != nil {
		log.Error("Failed to listen on tcp socket: ", err)
		panic(err)
	}
	for {
		log.Info("Waiting for connection")
		client, err := conn.Accept()
		if err != nil {
			log.Error("Server accept err : ", err)
			return
		}
		snw.TcpConn = client
		fmt.Println("Accept connections from : ", client.RemoteAddr())
		go snw.clientHandler(snw.Cos, snw.Net, client)
	}
}

// ClientHandler
func (snw *ServerNetWork) clientHandler(config *model.CreateOptions, water *water.Interface, client net.Conn) {
	go snw.readTunToTCPNetwork()
	go snw.readTCPNetworkToTUN()
}

func (snw *ServerNetWork) readTunToTCPNetwork() {
	buf := make([]byte, snw.Cos.BufferSize)
	for {
		n, err := snw.Net.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		b := buf[:n]
		if key := netutil.GetDstKey(b); key != "" {
			snw.Net.Write(b)
			snw.setSentBytes(n)
		}
	}
}

func (snw *ServerNetWork) readTCPNetworkToTUN() {
	buf := make([]byte, snw.Cos.BufferSize)
	for {
		n, err := snw.TcpConn.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		b := buf[:n]
		if key := netutil.GetSrcKey(b); key != "" {
			snw.Net.Write(b)
			snw.connCache.Set(key, snw.TcpConn.RemoteAddr(), cache.DefaultExpiration)
			snw.receivingBytes(n)
		}
	}
}

func (snw *ServerNetWork) setSentBytes(n int) {
	atomic.AddUint64(&ServerSentBytes, uint64(n))
}

func (snw *ServerNetWork) receivingBytes(n int) {
	atomic.AddUint64(&ServerReceivingBytes, uint64(n))
}
