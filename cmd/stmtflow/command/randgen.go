package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PingCAP-QE/clustered-index-rand-test/sqlgen"
	"github.com/spf13/cobra"
)

const (
	DropDBIfExistSQL      = "/* init */ drop database if exists test;"
	CreateDBIfNotExistSQL = "/* init */ create database if not exists test;"
	UseDBSQL              = "/* init */ use test;"
)

type options struct {
	caseName    string
	withResults bool
}

func genCase(basicPath string, initSQls []string, testSQLs []string, opts options) (string, error) {
	testCaseFileName := filepath.Join(basicPath, opts.caseName+stdTestExt)
	f, err := os.OpenFile(testCaseFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n", DropDBIfExistSQL)
	fmt.Fprintf(f, "%s\n", CreateDBIfNotExistSQL)
	fmt.Fprintf(f, "%s\n", UseDBSQL)
	for _, sql := range initSQls {
		_, err = fmt.Fprintf(f, "/* init */ %s;\n", strings.TrimSpace(sql))
		if err != nil {
			return "", err
		}
	}
	for _, sql := range testSQLs {
		_, err = fmt.Fprintf(f, "/* s1 */ %s;\n", strings.TrimSpace(sql))
		if err != nil {
			return "", err
		}
	}
	return testCaseFileName, nil
}

func RandGen(c *CommonOptions) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:           "randgen <case_name>",
		Short:         "Generate rand test cases using cluster index rand gen",
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			state := sqlgen.NewState()
			gen := sqlgen.NewGenerator(state)
			count := 50
			initSQLs := make([]string, 0, 50)
			testSQLs := make([]string, 0, 50)
			for state.IsInitializing() {
				ss := gen()
				for _, s := range strings.Split(ss, " ; ") {
					initSQLs = append(initSQLs, s)
				}
			}

			for i := 0; i < count; i++ {
				ss := gen()
				for _, s := range strings.Split(ss, " ; ") {
					testSQLs = append(testSQLs, s)
				}
			}
			path := args[0]
			casePath, err := genCase(path, initSQLs, testSQLs, opts)
			if err != nil {
				return err
			}
			if opts.withResults {
				ctx := context.Background()
				playOps := PlayOpts{Write: true}
				playOps.TextDumpOptions.Verbose = true
				err = GenResultFile(ctx, c, casePath, playOps)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.caseName, "case_name", "sample_case", "case_name for the generated cases")
	cmd.Flags().BoolVarP(&opts.withResults, "with_results", "r", false, "write to result file")
	return cmd
}
