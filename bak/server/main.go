package main

import (
	"fmt"
	"github.com/net-byte/vtun/common/netutil"
	"github.com/patrickmn/go-cache"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

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
	conn, err := net.Listen("tcp", config.LocalAddr)
	if err != nil {
		log.Fatalln("failed to listen on udp socket:", err)
	}
	go func() {
		for { // 循环等待创建连接,实现可以接收多个client的连接功能
			fmt.Println("Accept")
			conn, err := conn.Accept() // 阻塞等待连接
			if err != nil {
				fmt.Println("server accept err :", err)
				return
			}
			fmt.Println("start Accept")
			// 并发创建用于完成server 与 client 数据通信的函数
			//go HandlerConnent(conn)  // 将conn socket传入
			go StartServer(config, iface, conn)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("接受字节数:", formatFileSize(int64(_totalReadBytes)))
	fmt.Println("发送字节数:", formatFileSize(int64(_totalWrittenBytes)))
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
func StartServer(config Config, iface *water.Interface, client net.Conn) {
	defer client.Close()
	log.Printf("vtun udp server started on %v", config.LocalAddr)
	s := &Server{config: config, iface: iface, localConn: client, connCache: cache.New(30*time.Minute, 10*time.Minute)}
	go s.tunToUdp()
	s.udpToTun()
}

// the server struct
type Server struct {
	config    Config
	iface     *water.Interface
	localConn net.Conn
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
			fmt.Println("key", key)
			s.localConn.Write(b)
			IncrWrittenBytes(n)
		}
		//s.localConn.Write(b)
		//IncrWrittenBytes(n)
	}
}

// udpToTun sends packets from udp to tun
func (s *Server) udpToTun() {
	packet := make([]byte, s.config.BufferSize)
	for {
		n, err := s.localConn.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		b := packet[:n]
		s.iface.Write(b)
		IncrReadBytes(n)
	}
}

// totalReadBytes is the total number of bytes read
var _totalReadBytes uint64 = 0

// totalWrittenBytes is the total number of bytes written
var _totalWrittenBytes uint64 = 0

// IncrReadBytes increments the number of bytes read
func IncrReadBytes(n int) {
	atomic.AddUint64(&_totalReadBytes, uint64(n))
}

// IncrWrittenBytes increments the number of bytes written
func IncrWrittenBytes(n int) {
	atomic.AddUint64(&_totalWrittenBytes, uint64(n))
}

// GetReadBytes returns the number of bytes read
func GetReadBytes() uint64 {
	return _totalReadBytes
}

// GetWrittenBytes returns the number of bytes written
func GetWrittenBytes() uint64 {
	return _totalWrittenBytes
}

// 字节的单位转换 保留两位小数
func formatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
