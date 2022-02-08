package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zyguan/tidb-test-util/pkg/fs"
)

const (
	EnvFSHost   = "DODO_FS_HOST"
	EnvFSPort   = "DODO_FS_PORT"
	EnvFSFBHost = "DODO_FS_FBHOST"
	EnvFSFBPort = "DODO_FS_FBPORT"
	EnvFSFBUser = "DODO_FS_FBUSER"
	EnvFSFBPass = "DODO_FS_FBPASS"
)

func FileServer() *cobra.Command {
	cli := fs.Default()
	cmd := &cobra.Command{
		Use:   "fs",
		Short: "File server utils",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if v := os.Getenv(EnvFSHost); len(v) > 0 {
				cli.Host = v
			}
			if v := os.Getenv(EnvFSFBHost); len(v) > 0 {
				cli.FBHost = v
			}
			if v := os.Getenv(EnvFSFBUser); len(v) > 0 {
				cli.FBUser = v
			}
			if v := os.Getenv(EnvFSFBPass); len(v) > 0 {
				cli.FBPass = v
			}
			if v := os.Getenv(EnvFSPort); len(v) > 0 {
				p, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("invalid port: %v", err)
				}
				cli.Port = p
			}
			if v := os.Getenv(EnvFSFBPort); len(v) > 0 {
				p, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("invalid port: %v", err)
				}
				cli.FBPort = p
			}
			return cli.Auth()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(fsGet(cli))
	cmd.AddCommand(fsPut(cli))
	cmd.AddCommand(fsDel(cli))
	cmd.AddCommand(fsMove(cli))
	cmd.AddCommand(fsCopy(cli))
	cmd.AddCommand(fsStat(cli))
	cmd.AddCommand(fsWhereIs(cli))

	return cmd
}

func fsGet(cli *fs.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "get <remote> [local]",
		Short:         "Download a file",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			remote := args[0]
			local := filepath.Base(remote)
			if len(args) > 1 {
				local = args[1]
			}
			fmt.Fprintf(os.Stderr, "downloading from %s ...\n", cli.DownloadURL(remote))
			return cli.GetFile(remote, local)
		},
	}
	return cmd
}

func fsPut(cli *fs.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "put <remote> <local>",
		Short:         "Upload a file",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			remote, local := args[0], args[1]
			f, err := os.Open(local)
			if err != nil {
				return err
			}
			if fi, err := f.Stat(); err != nil {
				return err
			} else if fi.IsDir() {
				return errors.New("cannot upload a directory yet")
			}
			fmt.Fprintf(os.Stderr, "uploading to %s ...\n", remote)
			err = cli.PutFile(remote, fs.File(f))
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, cli.DownloadURL(remote))
			return nil
		},
	}

	return cmd
}

func fsDel(cli *fs.Client) *cobra.Command {
	var opts struct {
		Force bool
	}
	cmd := &cobra.Command{
		Use:           "del <remote>",
		Short:         "Delete a file or directory",
		SilenceUsage:  true,
		SilenceErrors: true,
		Hidden:        true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			remote := args[0]
			info, err := cli.Stat(remote, false)
			if err != nil {
				if strings.Contains(err.Error(), remote+": 404 Not Found") {
					return nil
				}
				return err
			}
			if info.Dir && len(info.Items) > 0 && !opts.Force {
				return errors.New("cannot delete a non-empty directory without force option")
			}
			if info.Dir && len(info.Items) > 0 {
				var confirm string
				fmt.Fprint(os.Stdout, "Please type 'yes' to confirm: ")
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Fprintln(os.Stdout, "Abort by user")
					return nil
				}
			}
			fmt.Fprintf(os.Stderr, "deleting %s ...\n", remote)
			return cli.DelFile(remote, opts.Force)
		},
	}
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "force delete a file or directory")

	return cmd
}

func fsMove(cli *fs.Client) *cobra.Command {
	var opts struct {
		Force          bool
		IgnoreChecksum bool
	}
	cmd := &cobra.Command{
		Use:           "mv <src> <dst>",
		Short:         "Move a file or directory",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst := args[0], args[1]
			info, err := cli.Stat(src, false)
			if err != nil {
				return err
			}
			_, err = cli.Stat(dst, false)
			if err == nil && !opts.Force {
				return fmt.Errorf("cannot move because %s exists", dst)
			}
			if info.Dir {
				fmt.Fprintf(os.Stderr, "moving diretory %s to %s ...\n", src, dst)
				return cli.Rename(src, dst)
			} else {
				fmt.Fprintf(os.Stderr, "moving file %s to %s ...\n", src, dst)
				err = cli.MoveFile(src, dst)

				if opts.IgnoreChecksum && err != nil && strings.Contains(err.Error(), fs.ExtChecksum+": 404 Not Found") {
					return cli.Rename(src, dst)
				}
				return err
			}

		},
	}
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "ignore existence of destination and force move")
	cmd.Flags().BoolVarP(&opts.IgnoreChecksum, "ignore-checksum", "i", false, "move a file without checksum")

	return cmd
}

func fsCopy(cli *fs.Client) *cobra.Command {
	var opts struct {
		Force          bool
		IgnoreChecksum bool
	}
	cmd := &cobra.Command{
		Use:           "cp <src> <dst>",
		Short:         "Copy a file or directory",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst := args[0], args[1]
			info, err := cli.Stat(src, false)
			if err != nil {
				return err
			}
			_, err = cli.Stat(dst, false)
			if err == nil && !opts.Force {
				return fmt.Errorf("cannot copy because %s exists", dst)
			}
			if info.Dir {
				fmt.Fprintf(os.Stderr, "copying diretory %s to %s ...\n", src, dst)
				return cli.Copy(src, dst)
			} else {
				fmt.Fprintf(os.Stderr, "copying file %s to %s ...\n", src, dst)
				err = cli.CopyFile(src, dst)

				if opts.IgnoreChecksum && err != nil && strings.Contains(err.Error(), fs.ExtChecksum+": 404 Not Found") {
					return cli.Copy(src, dst)
				}
				return err
			}

		},
	}
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "ignore existence of destination and force copy")
	cmd.Flags().BoolVarP(&opts.IgnoreChecksum, "ignore-checksum", "i", false, "copy a file without checksum")

	return cmd
}

func fsStat(cli *fs.Client) *cobra.Command {
	var opts struct {
		OutFmt string
	}
	cmd := &cobra.Command{
		Use:           "stat <path>",
		Short:         "Show file info",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			info, err := cli.Stat(path, false)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, cli.Format(info, opts.OutFmt))
			if info.Dir && strings.HasSuffix(path, "/") {
				for _, item := range info.Items {
					fmt.Fprintln(os.Stdout, cli.Format(&item, opts.OutFmt))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.OutFmt, "output", "o", "detail", "output format [name|url|detail]")
	return cmd
}

func fsWhereIs(cli *fs.Client) *cobra.Command {
	var opts struct {
		OutFmt string
	}
	cmd := &cobra.Command{
		Use:           "whereis <name> [ref]",
		Short:         "Where is the component archive built by CI",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, ref := args[0], "master"
			if len(args) > 1 {
				ref = args[1]
			}
			remote := cli.WhereIsComponent(name, ref)
			if len(remote) == 0 {
				return errors.New("not found")
			}
			switch opts.OutFmt {
			case "url":
				fmt.Fprintln(os.Stdout, cli.DownloadURL(remote))
			case "path":
				fmt.Fprintln(os.Stdout, remote)
			default:
				fmt.Fprintln(os.Stdout, cli.DownloadURL(remote))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.OutFmt, "output", "o", "url", "output format [url|path]")

	return cmd
}
