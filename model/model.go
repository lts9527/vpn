package model

type CreateOptions struct {
	ServerMode     bool   `json:"server_mode,omitempty"`
	BufferSize     int    `json:"buffer_size,omitempty"`
	MTU            int    `json:"mtu,omitempty"`
	OS             string `json:"os,omitempty"`
	Key            string `json:"key,omitempty"`
	DNS            string `json:"dns,omitempty"`
	Mode           string `json:"mode,omitempty"`
	NetworkMode    string `json:"network_mode,omitempty"`
	ClientAddress  string `json:"client_address,omitempty"`
	Gateway        string `json:"gateway,omitempty"`
	ListenPort     string `json:"listen_port,omitempty"`
	DeviceName     string `json:"device_name,omitempty"`
	PhysicalDevice string `json:"physical_device,omitempty"`
	RemoteServerIP string `json:"remote_server_ip,omitempty"`
	ServerAddress  string `json:"server_address,omitempty"`
	LocalGateway   string `json:"local_gateway,omitempty"`
}
