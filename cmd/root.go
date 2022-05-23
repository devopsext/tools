package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var version = "unknown"
var APPNAME = "TOOLS"
var stdout *common.Stdout

var stdoutOptions = common.StdoutOptions{
	Format:          envGet("STDOUT_FORMAT", "template").(string),
	Level:           envGet("STDOUT_LEVEL", "info").(string),
	Template:        envGet("STDOUT_TEMPLATE", "{{.file}} {{.msg}}").(string),
	TimestampFormat: envGet("STDOUT_TIMESTAMP_FORMAT", time.RFC3339Nano).(string),
	TextColors:      envGet("STDOUT_TEXT_COLORS", true).(bool),
}

func envGet(s string, d interface{}) interface{} {
	return utils.EnvGet(fmt.Sprintf("%s_%s", APPNAME, s), d)
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
	rootCmd.AddCommand(NewCmdbCommand())

	if err := rootCmd.Execute(); err != nil {
		stdout.Error(err)
		os.Exit(1)
	}
}
