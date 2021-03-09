package command

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zyguan/sqlz/stmtflow"
	"github.com/zyguan/tidb-test-util/cmd/stmtflow/core"
)

type testOptions struct {
	stmtflow.EvalOptions
	Filter  string
	DryRun  bool
	Diff    bool
	DiffCmd string
}

func Test(c *CommonOptions) *cobra.Command {
	opts := testOptions{}
	cmd := &cobra.Command{
		Use:           "test [tests.jsonnet ...]",
		Short:         "Run tests",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			opts.EvalOptions = c.EvalOptions()
			ctx := context.Background()
			errCnt, skippedCnt := 0, 0
			for _, path := range args {
				log.Printf("[%s] load tests", path)
				tests, err := core.Load(path, opts.Filter)
				if err != nil {
					return err
				}
				// TODO: support concurrent execution
				for _, t := range tests {
					if opts.DryRun {
						log.Printf("[%s#%s] type:%s labels:%s", path, t.Name, t.AssertMethod, t.Labels)
						continue
					}
					repeat := 1
					if repeat < t.Repeat {
						repeat = t.Repeat
					}
					var (
						db      *sql.DB
						skipped bool
					)
					for i := 0; i < repeat; i++ {
						db, err = c.OpenDB()
						if err != nil {
							return err
						}
						err = validateTiDBVersion(db, t)
						if err != nil {
							db.Close()
							skipped = true
							break
						}
						err = testOne(c.WithTimeout(ctx), db, t, opts)
						db.Close()
						if err != nil {
							break
						}
					}
					if err != nil {
						if skipped {
							log.Printf("[%s#%s] skipped: %v", path, t.Name, err)
							skippedCnt += 1
						} else {
							log.Printf("[%s#%s] failed:  %+v", path, t.Name, err)
							errCnt += 1
						}
					} else {
						log.Printf("[%s#%s] passed", path, t.Name)
					}
				}
			}
			if errCnt > 0 {
				plural := ""
				if errCnt > 1 {
					plural = "s"
				}
				return fmt.Errorf("%d test%s failed", errCnt, plural)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.Filter, "filter", "f", "", "filter tests by a jsonnet expr, eg. std.startsWith(test.name, 'foo')")
	cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "n", false, "just list tests to be run")
	cmd.Flags().BoolVar(&opts.Diff, "diff", false, "diff text output if available")
	cmd.Flags().StringVar(&opts.DiffCmd, "diff-cmd", "diff -u -N --color", "diff command to use")

	return cmd
}

func testOne(ctx context.Context, db *sql.DB, test core.Test, opts testOptions) (err error) {
	var actual stmtflow.History
	evalOpts := opts.EvalOptions
	evalOpts.Callback = actual.Collect
	err = stmtflow.Run(ctx, db, test.Test, evalOpts)
	if err != nil {
		return errors.Wrap(err, "run test")
	}
	err = test.Assert(actual)
	if err == nil || !opts.Diff {
		return
	}
	exp, ok := test.ExpectedText()
	if !ok {
		return
	}
	buf := new(bytes.Buffer)
	if e := actual.DumpText(buf, stmtflow.TextDumpOptions{Verbose: true}); e != nil {
		return
	}
	_ = core.LocalDiff(os.Stdout, test.Name, exp, buf.String(), strings.Split(opts.DiffCmd, " "))
	return
}

func validateTiDBVersion(db *sql.DB, test core.Test) error {
	if len(test.VersionConstraint) == 0 {
		return nil
	}
	var ver string
	err := db.QueryRow("select version()").Scan(&ver)
	if err != nil {
		return errors.Wrap(err, "query for version")
	}
	const verPrefix = "5.7.25-TiDB-"
	if strings.HasPrefix(ver, verPrefix) {
		ver = ver[len(verPrefix):]
	}
	return test.ValidateVersion(ver)
}
