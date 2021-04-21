package command

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/zyguan/sqlz/stmtflow"
	"github.com/zyguan/tidb-test-util/cmd/stmtflow/core"
)

type PlayOpts struct {
	stmtflow.TextDumpOptions
	Write bool
}

func GenResultFile(ctx context.Context, c *CommonOptions, path string, playOpts PlayOpts) error {
	var (
		in       io.ReadCloser
		result   stmtflow.History
		done     func()
		evalOpts = c.EvalOptions()
	)
	db, err := c.OpenDB()
	if err != nil {
		return err
	}
	fmt.Printf("after open err=%v c.dsn=%v, path=%v\n", err, c.DSN, path)
	in, err = os.Open(path)
	if err != nil {
		return err
	}
	if playOpts.Write {
		textOut, err := os.OpenFile(resultPathForText(path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		jsonOut, err := os.OpenFile(resultPathForJson(path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		textWriter := stmtflow.TextDumper(io.MultiWriter(os.Stdout, textOut), playOpts.TextDumpOptions)
		evalOpts.Callback = stmtflow.ComposeHandler(result.Collect, textWriter)
		done = func() {
			result.DumpJson(jsonOut, stmtflow.JsonDumpOptions{})
			jsonOut.Close()
			textOut.Close()
			in.Close()
			db.Close()
		}
	} else {
		evalOpts.Callback = stmtflow.TextDumper(os.Stdout, playOpts.TextDumpOptions)
		done = func() {
			in.Close()
			db.Close()
		}
	}

	if err = stmtflow.Run(c.WithTimeout(ctx), db, core.ParseSQL(in), evalOpts); err != nil {
		return err
	}

	done()
	return nil
}

func Play(c *CommonOptions) *cobra.Command {
	opts := PlayOpts{}
	cmd := &cobra.Command{
		Use:           "play [test.sql ...]",
		Short:         "Try tests",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			ctx := context.Background()
			for _, path := range args {
				fmt.Println("# " + path)
				err := GenResultFile(ctx, c, path, opts)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&opts.Write, "write", "w", false, "write to expected result files")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", true, "verbose output")
	cmd.Flags().BoolVar(&opts.WithLat, "with-lat", false, "record latency of each statement")
	return cmd
}
