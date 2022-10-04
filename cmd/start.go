package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"vpn/log"
	"vpn/model"
	"vpn/server"
)

var create = &model.CreateOptions{}

func runCreate(ctx context.Context, c *model.CreateOptions) {
	srv := server.NewServer(create)
	srv.Init()
	go srv.Start()
	//defer func() {
	//	if err := recover(); err != nil {
	//		fmt.Println(err)
	//	}
	//}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Warn("stop")
	srv.Stop()
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start",
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Println("start", create)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		runCreate(ctx, create)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&create.ServerMode, "server", "S", false, "运行的模式 [server client]")
	startCmd.Flags().StringVarP(&create.NetworkMode, "network", "n", "udp", "运行的网络模式 [tcp udp]")
	startCmd.Flags().StringVarP(&create.ListenPort, "port", "p", "9527", "监听的端口")
	startCmd.Flags().StringVarP(&create.RemoteServerIP, "remoteIP", "r", "45.195.69.18", "远程服务器真实ip")
	startCmd.Flags().StringVarP(&create.ClientAddress, "clientIP", "c", "172.16.0.10", "客户端ip")
	startCmd.Flags().StringVarP(&create.ServerAddress, "serverIP", "s", "172.16.0.1", "服务端ip")
	startCmd.Flags().StringVarP(&create.DNS, "dns", "", "8.8.8.8", "dns")
}
