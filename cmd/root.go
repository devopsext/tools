package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var version = "unknown"
var APPNAME = "TOOLS"
var stdout *common.Stdout
var mainWG sync.WaitGroup

var stdoutOptions = common.StdoutOptions{
	Format:          envGet("STDOUT_FORMAT", "template").(string),
	Level:           envGet("STDOUT_LEVEL", "info").(string),
	Template:        envGet("STDOUT_TEMPLATE", "{{.file}} {{.msg}}").(string),
	TimestampFormat: envGet("STDOUT_TIMESTAMP_FORMAT", time.RFC3339Nano).(string),
	TextColors:      envGet("STDOUT_TEXT_COLORS", true).(bool),
}

func getOnlyEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	return fmt.Sprintf("$%s", key)
}

func envGet(s string, def interface{}) interface{} {
	return utils.EnvGet(fmt.Sprintf("%s_%s", APPNAME, s), def)
}

func envStringExpand(s string, def string) string {
	snew := envGet(s, def).(string)
	return os.Expand(snew, getOnlyEnv)
}

func envFileContentExpand(s string, def string) string {
	snew := envGet(s, def).(string)
	bytes, err := utils.Content(snew)
	if err != nil {
		return def
	}
	return os.Expand(string(bytes), getOnlyEnv)
}

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "tools",
		Short: "Tools",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			stdout = common.NewStdout(stdoutOptions)
			stdout.SetCallerOffset(1)
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringVar(&stdoutOptions.Format, "stdout-format", stdoutOptions.Format, "Stdout format: json, text, template")
	flags.StringVar(&stdoutOptions.Level, "stdout-level", stdoutOptions.Level, "Stdout level: info, warn, error, debug, panic")
	flags.StringVar(&stdoutOptions.Template, "stdout-template", stdoutOptions.Template, "Stdout template")
	flags.StringVar(&stdoutOptions.TimestampFormat, "stdout-timestamp-format", stdoutOptions.TimestampFormat, "Stdout timestamp format")
	flags.BoolVar(&stdoutOptions.TextColors, "stdout-text-colors", stdoutOptions.TextColors, "Stdout text colors")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	})

	rootCmd.AddCommand(NewSlackCommand())
	rootCmd.AddCommand(NewTelegramCommand())
	rootCmd.AddCommand(NewGraylogCommand())
	rootCmd.AddCommand(NewJiraCommand())
	rootCmd.AddCommand(NewGrafanaCommand())
	rootCmd.AddCommand(NewJSONCommand())
	rootCmd.AddCommand(NewGitlabCommand())
	rootCmd.AddCommand(NewGoogleCommand())
	rootCmd.AddCommand(NewPrometheusCommand())
	rootCmd.AddCommand(NewObserviumCommand())
	rootCmd.AddCommand(NewZabbixCommand())
	rootCmd.AddCommand(NewVCenterCommand())
	rootCmd.AddCommand(NewPagerDutyCommand())
	rootCmd.AddCommand(NewAWSCommand())
	rootCmd.AddCommand(NewSite24x7Command())
	rootCmd.AddCommand(NewCatchpointCommand())
	rootCmd.AddCommand(NewVirusTotalCommand())
	rootCmd.AddCommand(NewCryptoCommand())
	rootCmd.AddCommand(NewNetboxCommand())
	rootCmd.AddCommand(NewK8sCommand())
	rootCmd.AddCommand(NewTeleportCommand())

	rootCmd.AddCommand(NewTemplateCommand())
	rootCmd.AddCommand(NewDateCommand())

	rootCmd.AddCommand(NewServerCommand(&mainWG))

	if err := rootCmd.Execute(); err != nil {
		stdout.Error(err)
		os.Exit(1)
	}
}
