package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/net-byte/vtun/common/counter"
	"github.com/net-byte/vtun/common/netutil"
	"github.com/net-byte/water"
)

// The config struct
type Config struct {
	DeviceName string
	LocalAddr  string
	ServerAddr string
	ServerIP   string
	BufferSize int
}

// The client struct
type Client struct {
	config     Config
	iface      *water.Interface
	localConn  *net.UDPConn
	serverAddr *net.UDPAddr
}

func main() {
	iface, err := water.New(water.Config{
		DeviceType:             water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{},
	})
	if err != nil {
		log.Fatalln("failed to create tun interface:", err)
	}
	config := Config{
		DeviceName: iface.Name(),
		LocalAddr:  ":3001",
		ServerIP:   "",
		BufferSize: 64 * 1024,
	}
	setTun(config.DeviceName)
	StartServer(config, iface)
	select {}
}

func setTun(name string) {
	ExecCmd("ip", "link", "set", "dev", name, "mtu", strconv.Itoa(1500))
	ExecCmd("ip", "addr", "add", "172.16.0.1/24", "dev", name)
	ExecCmd("ip", "link", "set", "dev", name, "up")
	//ExecCmd("route", "add", "default", "172.16.0.1")
	//ExecCmd("route", "change", "default", "172.16.0.1")
	//ExecCmd("route", "add", "0.0.0.0/1", "-interface", name)
	//ExecCmd("route", "add", "128.0.0.0/1", "-interface", name)
	//ExecCmd("route", "add", "45.195.69.18", "192.168.1.1")
	//ExecCmd("route", "add", "8.8.8.8", "192.168.1.1")
}

// ExecCmd executes the given command
func ExecCmd(c string, args ...string) string {
	fmt.Printf("exec %v %v\n", c, args)
	cmd := exec.Command(c, args...)
	out, err := cmd.Output()
	if err != nil {
		log.Println("failed to exec cmd:", err)
	}
	if len(out) == 0 {
		return ""
	}
	return strings.ReplaceAll(string(out), "\n", "")
}

// StartServer starts the udp server
func StartServer(config Config, iface *water.Interface) {
	log.Printf("vtun udp server started on %v", config.LocalAddr)
	localAddr, err := net.ResolveUDPAddr("udp", config.LocalAddr)
	if err != nil {
		log.Fatalln("failed to get udp socket:", err)
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatalln("failed to listen on udp socket:", err)
	}
	defer conn.Close()
	s := &Server{config: config, iface: iface, localConn: conn, connCache: cache.New(30*time.Minute, 10*time.Minute)}
	go s.tunToUdp()
	s.udpToTun()
}

// the server struct
type Server struct {
	config    Config
	iface     *water.Interface
	localConn *net.UDPConn
	connCache *cache.Cache
}

// tunToUdp sends packets from tun to udp
func (s *Server) tunToUdp() {
	packet := make([]byte, s.config.BufferSize)
	for {
		n, err := s.iface.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		b := packet[:n]
		if key := netutil.GetDstKey(b); key != "" {
			if v, ok := s.connCache.Get(key); ok {
				s.localConn.WriteToUDP(b, v.(*net.UDPAddr))
				counter.IncrWrittenBytes(n)
			}
		}
	}
}

// udpToTun sends packets from udp to tun
func (s *Server) udpToTun() {
	packet := make([]byte, s.config.BufferSize)
	for {
		n, cliAddr, err := s.localConn.ReadFromUDP(packet)
		if err != nil || n == 0 {
			continue
		}
		b := packet[:n]
		if key := netutil.GetSrcKey(b); key != "" {
			s.iface.Write(b)
			s.connCache.Set(key, cliAddr, cache.DefaultExpiration)
			counter.IncrReadBytes(n)
		}
	}
}
