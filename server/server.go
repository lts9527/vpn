package server

import (
	"github.com/net-byte/water"
	"runtime"
	"vpn/model"
	"vpn/network/tcp"
	"vpn/network/udp"
	"vpn/pkg/utils"
	"vpn/service"
)

type Server struct {
	svc    *service.Service
	Tun    *water.Interface
	config *model.CreateOptions
}

func NewServer(config *model.CreateOptions) *Server {
	return &Server{
		svc:    service.NewService(),
		config: config,
	}
}

func (s *Server) Init() {
	var err error
	if !s.config.ServerMode {
		s.config.LocalGateway = s.getGateway()
	}
	s.config.OS = runtime.GOOS
	s.config.BufferSize = 64 * 1024
	s.Tun, err = s.CreateTUN(s.config)
	if err != nil {
		panic(err)
	}
	s.config.DeviceName = s.Tun.Name()
	s.SetTUN(s.config)

}

func (s *Server) Start() {
	switch s.config.NetworkMode {
	case "udp":
		if s.config.ServerMode {
			net := s.NewUDPServerNetWork()
			net.ListenUDP()
		} else {
			net := s.NewUDPClientNetWork()
			net.ClientDial()
		}
	case "tcp":
		if s.config.ServerMode {
			net := s.NewTCPServerNetWork()
			net.ListenTCP()
		} else {
			net := s.NewTCPClientNetWork()
			net.ClientDial()
		}
	default:
		panic("Select the correct mode")
	}
}

func (s *Server) Stop() {
	s.ResetTUN(s.config)
}

func (s *Server) NewTCPServerNetWork() *tcp.ServerNetWork {
	return tcp.NewServerNetWork(s.config, s.Tun)
}

func (s *Server) NewTCPClientNetWork() *tcp.ClientNetWork {
	return tcp.NewClientNetWork(s.config, s.Tun)
}

func (s *Server) NewUDPServerNetWork() *udp.ServerNetWork {
	return udp.NewServerNetWork(s.config, s.Tun)
}

func (s *Server) NewUDPClientNetWork() *udp.ClientNetWork {
	return udp.NewClientNetWork(s.config, s.Tun)
}

func (s *Server) CreateTUN(config *model.CreateOptions) (*water.Interface, error) {
	return s.svc.CreateClientTUN(config)
}

func (s *Server) SetTUN(config *model.CreateOptions) {
	s.svc.SetTUN(config)
}

func (s *Server) ResetTUN(config *model.CreateOptions) {
	s.svc.ResetTUN(config)
}

func (s *Server) getGateway() string {
	str, _ := utils.DiscoverGatewayOSSpecificIPv4()
	return str.String()
}
