package cmd

import (
	"sync"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/server"
	"github.com/spf13/cobra"
)

var httpServerOptions = server.HttpServerOptions{
	ServerName: envGet("HTTP_SERVER_NAME", "").(string),
	Listen:     envGet("HTTP_SERVER_LISTEN", ":80").(string),
	Tls:        envGet("HTTP_SERVER_TLS", false).(bool),
	Insecure:   envGet("HTTP_SERVER_INSECURE", false).(bool),
	Cert:       envGet("HTTP_SERVER_CERT", "").(string),
	Key:        envGet("HTTP_SERVER_KEY", "").(string),
	Chain:      envGet("HTTP_SERVER_CHAIN", "").(string),
	Timeout:    envGet("HTTP_SERVER_TIMEOUT", 30).(int),
}

func httpServerNew(stdout *common.Stdout) *server.HttpServer {

	common.Debug("HttpServer", httpServerOptions, stdout)
	return server.NewHttpServer(httpServerOptions, stdout)
}

func NewServerCommand(wg *sync.WaitGroup) *cobra.Command {

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Server tools",
	}

	httpServerCmd := &cobra.Command{
		Use:   "http",
		Short: "Run HTTP Server",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Running http server...")
			httpServerNew(stdout).Start(wg)
			wg.Wait()
		},
	}
	flags := httpServerCmd.PersistentFlags()
	flags.StringVar(&httpServerOptions.ServerName, "http-server-name", httpServerOptions.ServerName, "Http server name")
	flags.StringVar(&httpServerOptions.Listen, "http-server-listen", httpServerOptions.Listen, "Http server listen")
	flags.BoolVar(&httpServerOptions.Tls, "http-server-tls", httpServerOptions.Tls, "Http server TLS")
	flags.BoolVar(&httpServerOptions.Insecure, "http-server-insecure", httpServerOptions.Insecure, "Http server insecure skip verify")
	flags.StringVar(&httpServerOptions.Cert, "http-server-cert", httpServerOptions.Cert, "Http server cert file or content")
	flags.StringVar(&httpServerOptions.Key, "http-server-key", httpServerOptions.Key, "Http server key file or content")
	flags.StringVar(&httpServerOptions.Chain, "http-server-chain", httpServerOptions.Chain, "Http server CA chain file or content")
	flags.IntVar(&httpServerOptions.Timeout, "http-server-timeout", httpServerOptions.Timeout, "Http server timeout")
	serverCmd.AddCommand(httpServerCmd)

	return serverCmd
}
