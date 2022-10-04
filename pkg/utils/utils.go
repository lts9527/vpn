package utils

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"vpn/log"
)

// ExecCmd executes the given command
func ExecCmd(c string, args ...string) string {
	log.Info("exec %v %v", c, args)
	cmd := exec.Command(c, args...)
	out, _ := cmd.Output()
	if len(out) == 0 {
		return ""
	}
	return strings.ReplaceAll(string(out), "\n", "")
}

// FormatFileSize 字节的单位转换 保留两位小数
func FormatFileSize(fileSize int64) (size string) {
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

func DiscoverGatewayOSSpecificIPv4() (ip net.IP, err error) {
	if runtime.GOOS == "linux" {
		ipstr := ExecCmd("route -n | grep -A 1 'Gateway' | awk 'NR==2{print $2}'")
		ipv4 := net.ParseIP(ipstr)
		if ipv4 == nil {
			return nil, errors.New("can't parse string output")
		}
		return ipv4, nil
	}
	ipstr := ExecCmd("sh", "-c", "route -n get default | grep 'gateway' | awk 'NR==1{print $2}'")
	ipv4 := net.ParseIP(ipstr)
	if ipv4 == nil {
		return nil, errors.New("can't parse string output")
	}
	return ipv4, nil
}
