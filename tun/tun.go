package tun

import (
	"fmt"
	"github.com/net-byte/water"
	"net/netip"
	"runtime"
	"strconv"
	"vpn/model"
	"vpn/pkg/utils"
)

func CreateNetTUN(ClientAddress, ServerAddress []netip.Addr, mtu int) (*water.Interface, error) {
	tun, err := water.New(water.Config{
		DeviceType:             water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tun interface: %v", err)
	}
	return tun, nil
}

func SetServerTUN(config *model.CreateOptions) {
	utils.ExecCmd("ip", "link", "set", "dev", config.DeviceName, "mtu", strconv.Itoa(config.MTU))
	utils.ExecCmd("ip", "addr", "add", config.ServerAddress, "dev", config.DeviceName)
	utils.ExecCmd("ip", "link", "set", "dev", config.DeviceName, "up")
}

func SetClientTUN(config *model.CreateOptions) {
	switch runtime.GOOS {
	case "linux":
		utils.ExecCmd("ip", "link", "set", "dev", config.DeviceName, "mtu", strconv.Itoa(config.MTU))
		utils.ExecCmd("ip", "addr", "add", config.ClientAddress, "dev", config.DeviceName)
		utils.ExecCmd("ip", "link", "set", "dev", config.DeviceName, "up")
		utils.ExecCmd("route", "add", "0.0.0.0/1", "dev", config.DeviceName)
		utils.ExecCmd("route", "add", "128.0.0.0/1", "dev", config.DeviceName)
		utils.ExecCmd("route", "add", config.ServerAddress, "via", config.LocalGateway, "dev", config.PhysicalDevice)
		utils.ExecCmd("route", "add", config.DNS, "via", config.LocalGateway, "dev", config.PhysicalDevice)
	case "darwin":
		utils.ExecCmd("ifconfig", config.DeviceName, "inet", config.ClientAddress, config.ServerAddress, "up")
		utils.ExecCmd("route", "add", "default", config.ServerAddress)
		utils.ExecCmd("route", "change", "default", config.ServerAddress)
		utils.ExecCmd("route", "add", "0.0.0.0/1", "-interface", config.DeviceName)
		utils.ExecCmd("route", "add", "128.0.0.0/1", "-interface", config.DeviceName)
		utils.ExecCmd("route", "add", config.RemoteServerIP, config.LocalGateway)
		utils.ExecCmd("route", "add", config.DNS, config.LocalGateway)
	default:
		panic("Unsupported operating system")
	}
}

func ResetTUN(config *model.CreateOptions) {
	if config.OS == "darwin" {
		utils.ExecCmd("route", "add", "default", config.LocalGateway)
		utils.ExecCmd("route", "change", "default", config.LocalGateway)
	}
}
