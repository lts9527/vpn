package main

import (
	"fmt"
	"github.com/net-byte/water"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
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
	serverConn net.Conn
	serverAddr *net.TCPAddr
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
		//	LocalAddr:  ":3741",
		ServerAddr: "45.195.69.18:3001",
		ServerIP:   "",
		BufferSize: 64 * 1024,
	}
	setTun(config)
	go StartClient(config, iface)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	StopApp()
	fmt.Println("接受字节数:", formatFileSize(int64(_totalReadBytes)))
	fmt.Println("发送字节数:", formatFileSize(int64(_totalWrittenBytes)))
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

func setTun(config Config) {
	ExecCmd("ifconfig", config.DeviceName, "172.16.0.10", "172.16.0.1", "up")
	ExecCmd("route", "add", "default", "172.16.0.1")
	ExecCmd("route", "change", "default", "172.16.0.1")
	ExecCmd("route", "add", "0.0.0.0/1", "-interface", config.DeviceName)
	ExecCmd("route", "add", "128.0.0.0/1", "-interface", config.DeviceName)
	ExecCmd("route", "add", "45.195.69.18", "192.168.1.1")
	ExecCmd("route", "add", "8.8.8.8", "192.168.1.1")
}

func StopApp() {
	ExecCmd("route", "add", "default", "192.168.1.1")
	ExecCmd("route", "change", "default", "192.168.1.1")
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

// StartClient starts the udp client
func StartClient(config Config, dev *water.Interface) {
	serverConn, err := net.Dial("tcp", config.ServerAddr)
	if err != nil {
		serverConn.Close()
		return
	}
	fmt.Println("serverConn", serverConn.RemoteAddr())
	defer serverConn.Close()
	c := &Client{config: config, iface: dev, serverConn: serverConn}
	go c.udpToTun()
	c.tunToUdp()
}

// udpToTun sends packets from udp to tun
func (c *Client) udpToTun() {
	packet := make([]byte, 2000)
	for {
		n, err := c.serverConn.Read(packet)
		if err != nil {
			fmt.Println("client read err :", err)
			return
		}
		b := packet[:n]
		c.iface.Write(b)
		IncrReadBytes(n)
	}
}

// tunToUdp sends packets from tun to udp
func (c *Client) tunToUdp() {
	packet := make([]byte, 2000)
	for {
		n, err := c.iface.Read(packet)
		if err != nil || n == 0 {
			fmt.Println("err ", err)
			continue
		}
		b := packet[:n]
		// 1.主动向server端发送数据
		_, err = c.serverConn.Write(b)
		if err != nil {
			fmt.Println("client write err :", err)
			return
		}
		IncrWrittenBytes(n)
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
