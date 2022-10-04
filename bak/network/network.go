package network

import "C"
import (
	"fmt"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"log"
	"os/exec"
	"strings"
)

func Create(name string) *tun.Device {
	tun, err := tun.CreateTUN(name, 0)
	if err != nil {
		panic(err)
	}
	dev := device.NewDevice(tun, conn.NewStdNetBind(), device.NewLogger(
		3,
		fmt.Sprintf("(%s) ", name),
	))
	err = dev.Up()
	if err != nil {
		panic(err)
	}
	ExecCmd("ifconfig", name, "inet", "172.16.0.10", "172.16.0.1", "up")
	//ExecCmd("route", "add", "default", "172.16.0.1")
	ExecCmd("route", "change", "default", "172.16.0.1")
	ExecCmd("route", "add", "0.0.0.0/1", "-interface", name)
	ExecCmd("route", "add", "128.0.0.0/1", "-interface", name)
	ExecCmd("route", "add", "45.195.69.18", "192.168.1.1")
	ExecCmd("route", "add", "8.8.8.8", "192.168.1.1")
	return &tun
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
