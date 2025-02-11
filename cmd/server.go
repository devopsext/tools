package cmd

import (
	"strings"
	"sync"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/server"
	"github.com/spf13/cobra"
)

var httpServerOptions = server.HttpServerOptions{
	ServerName:      envGet("HTTP_SERVER_NAME", "").(string),
	Listen:          envGet("HTTP_SERVER_LISTEN", ":80").(string),
	Tls:             envGet("HTTP_SERVER_TLS", false).(bool),
	Insecure:        envGet("HTTP_SERVER_INSECURE", false).(bool),
	CA:              envGet("HTTP_SERVER_CA", "").(string),
	Crt:             envGet("HTTP_SERVER_CRT", "").(string),
	Key:             envGet("HTTP_SERVER_KEY", "").(string),
	Timeout:         envGet("HTTP_SERVER_TIMEOUT", 30).(int),
	Methods:         strings.Split(envGet("HTTP_SERVER_METHODS", "POST").(string), ","),
	SensitiveFields: strings.Split(envGet("HTTP_SERVER_SENSITIVE_FIELDS", "password,user,pass,username,token,secret").(string), ","),
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
	flags.StringVar(&httpServerOptions.CA, "http-server-ca", httpServerOptions.CA, "Http server ca file or content")
	flags.StringVar(&httpServerOptions.Crt, "http-server-crt", httpServerOptions.Crt, "Http server crt file or content")
	flags.StringVar(&httpServerOptions.Key, "http-server-key", httpServerOptions.Key, "Http server key file or content")
	flags.IntVar(&httpServerOptions.Timeout, "http-server-timeout", httpServerOptions.Timeout, "Http server timeout")
	flags.StringSliceVar(&httpServerOptions.Methods, "http-server-methods", httpServerOptions.Methods, "Http server methods")
	flags.StringSliceVar(&httpServerOptions.SensitiveFields, "http-server-sensitive-fields", httpServerOptions.SensitiveFields, "Http server sensitive fields")

	serverCmd.AddCommand(httpServerCmd)

	return serverCmd
}
