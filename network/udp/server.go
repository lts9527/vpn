package udp

import (
	"github.com/net-byte/vtun/common/netutil"
	"github.com/net-byte/water"
	"github.com/patrickmn/go-cache"
	"net"
	"sync/atomic"
	"time"
	"vpn/log"
	"vpn/model"
)

type ServerNetWork struct {
	Cos       *model.CreateOptions
	connCache *cache.Cache
	UdpConn   *net.UDPConn
	Net       *water.Interface
}

var (
	ServerSentBytes      uint64
	ServerReceivingBytes uint64
)

func NewServerNetWork(config *model.CreateOptions, Net *water.Interface) *ServerNetWork {
	snw := &ServerNetWork{
		Cos:       config,
		Net:       Net,
		connCache: cache.New(30*time.Minute, 10*time.Minute),
	}
	return snw
}

func (snw *ServerNetWork) ListenUDP() {
	ListenAddr, err := net.ResolveUDPAddr("udp", ":"+snw.Cos.ListenPort)
	if err != nil {
		log.Error("Failed to get udp socket:", err)
		panic(err)
	}
	conn, err := net.ListenUDP("udp", ListenAddr)
	if err != nil {
		log.Error("Failed to listen on udp socket:", err)
		panic(err)
	}
	snw.UdpConn = conn
	defer conn.Close()
	go snw.readTunToUDPNetwork()
	go snw.readUDPNetworkToTUN()
}

// tunToUdp sends packets from tun to udp
func (snw *ServerNetWork) readTunToUDPNetwork() {
	buf := make([]byte, snw.Cos.BufferSize)
	for {
		n, err := snw.Net.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		b := buf[:n]
		if key := netutil.GetDstKey(b); key != "" {
			if v, ok := snw.connCache.Get(key); ok {
				snw.UdpConn.WriteToUDP(b, v.(*net.UDPAddr))
				snw.receivingBytes(n)
			}
		}
	}
}

// udpToTun sends packets from udp to tun
func (snw *ServerNetWork) readUDPNetworkToTUN() {
	buf := make([]byte, snw.Cos.BufferSize)
	for {
		n, clientAddr, err := snw.UdpConn.ReadFromUDP(buf)
		if err != nil || n == 0 {
			continue
		}
		b := buf[:n]
		if key := netutil.GetSrcKey(b); key != "" {
			snw.Net.Write(b)
			snw.connCache.Set(key, clientAddr, cache.DefaultExpiration)
			snw.setSentBytes(n)
		}
	}
}

func (snw *ServerNetWork) setSentBytes(n int) {
	atomic.AddUint64(&ServerSentBytes, uint64(n))
}

func (snw *ServerNetWork) receivingBytes(n int) {
	atomic.AddUint64(&ServerReceivingBytes, uint64(n))
}
