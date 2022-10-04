package service

import (
	"github.com/net-byte/water"
	"net/netip"
	"vpn/model"
	"vpn/tun"
)

type Service struct {
	DeviceList map[string]*model.CreateOptions
}

func NewService() *Service {
	return &Service{
		DeviceList: make(map[string]*model.CreateOptions),
	}
}

func (s *Service) CreateClientTUN(config *model.CreateOptions) (Tun *water.Interface, err error) {
	Tun, err = tun.CreateNetTUN([]netip.Addr{netip.MustParseAddr(config.ClientAddress)}, []netip.Addr{netip.MustParseAddr(config.DNS)}, config.MTU)
	if err != nil {
		return nil, err
	}
	//s.DeviceList[Tun.Name()] = config
	return Tun, nil
}

func (s *Service) SetTUN(config *model.CreateOptions) {
	switch {
	case config.OS == "linux" && config.ServerMode:
		tun.SetServerTUN(config)
	case config.OS == "darwin" || config.OS == "linux":
		tun.SetClientTUN(config)
	default:
		panic("模式错误")
	}
}

func (s *Service) ResetTUN(config *model.CreateOptions) {
	tun.ResetTUN(config)
}
