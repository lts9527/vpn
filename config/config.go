package config

import (
	"fmt"
	"github.com/spf13/viper"
)

var (
	Conf *config
)

type config struct {
	viper *viper.Viper
}

type configInfo struct {
	ConfName string
	ConfType string
	ConfPath string
}

func init() {
	c1 := &configInfo{
		ConfName: "config",
		ConfType: "yaml",
		ConfPath: "config/",
	}
	WorkConf = &config{
		viper: getConf(c1),
	}
}

func getConf(c1 *configInfo) *viper.Viper {
	v := viper.New()
	v.SetConfigName(c1.ConfName)
	v.SetConfigType(c1.ConfType)
	v.AddConfigPath(c1.ConfPath)
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	return v
}

func (c *config) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *config) GetStringSlice(key string) []string {
	return c.viper.GetStringSlice(key)
}

func (c *config) GetInt(key string) int {
	return c.viper.GetInt(key)
}
