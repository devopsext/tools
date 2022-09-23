package cmd

import (
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

type DateOptions struct {
	Value  string
	Offset string
	Format string
}

var dateOptions = DateOptions{
	Value:  envGet("DATE_VALUE", "").(string),
	Offset: envGet("DATE_OFFSET", "").(string),
	Format: envGet("DATE_FORMAT", time.RFC3339Nano).(string),
}

var dateOutput = common.OutputOptions{
	Output: envGet("DATE_OUTPUT", "").(string),
}

func dateCalculate(opts DateOptions) (*time.Time, error) {

	t, err := time.Parse(opts.Format, opts.Value)
	if err != nil {
		return nil, err
	}
	if !utils.IsEmpty(opts.Offset) {
		d, err := time.ParseDuration(opts.Offset)
		if err != nil {
			return nil, err
		}
		t = t.Add(d)
	}
	return &t, nil
}

func NewDateCommand() *cobra.Command {

	dateCmd := &cobra.Command{
		Use:   "date",
		Short: "Date time tools",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Date calculation...")

			time, err := dateCalculate(dateOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			ts := time.Format(dateOptions.Format)
			common.OutputRaw(dateOutput.Output, []byte(ts), stdout)
		},
	}
	flags := dateCmd.PersistentFlags()
	flags.StringVar(&dateOptions.Value, "date-value", dateOptions.Value, "Date value")
	flags.StringVar(&dateOptions.Offset, "date-offset", dateOptions.Offset, "Date offset")
	flags.StringVar(&dateOptions.Format, "date-format", dateOptions.Format, "Date format")
	flags.StringVar(&dateOutput.Output, "date-output", dateOutput.Output, "Date output")

	return dateCmd
}
