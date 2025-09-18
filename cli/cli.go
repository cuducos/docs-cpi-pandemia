package cli

import (
	"path/filepath"
	"time"

	"github.com/cuducos/docs-cpi-pandemia/downloader"
	"github.com/cuducos/docs-cpi-pandemia/filesystem"
	"github.com/cuducos/docs-cpi-pandemia/unzip"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const help = "Downloads all public documents received by the CPI da Pandemia"

var (
	conns     uint
	retries   uint
	directory string
	cleanUp   bool
	timeout   string
	tolerant  bool
)

var cmd = &cobra.Command{
	Use:   "docs-cpi-pandemia",
	Short: help,
	RunE: func(_ *cobra.Command, _ []string) error {
		if cleanUp {
			filesystem.CleanDir(directory)
		}
		dur, err := time.ParseDuration(timeout)
		if err != nil {
			return err
		}
		if err := downloader.Download(directory, conns, retries, dur, tolerant); err != nil {
			return err

		}
		g := new(errgroup.Group)
		g.Go(func() error { return unzip.UnzipAll(directory) })
		g.Go(func() error { return filesystem.CleanDir(filepath.Join(directory, ".cache")) })
		return g.Wait()
	},
}

func CLI() *cobra.Command {
	cmd.Flags().BoolVarP(
		&cleanUp,
		"cleanup",
		"c",
		false,
		"Cleans up the target directory, including resetting the cache",
	)
	cmd.Flags().StringVarP(
		&directory,
		"directory",
		"d",
		"data",
		"Target directory to download the files",
	)
	cmd.Flags().UintVarP(
		&conns,
		"parallel",
		"p",
		32,
		"Maximum parallels downloads allowed",
	)
	cmd.Flags().UintVarP(
		&retries,
		"retries",
		"r",
		16,
		"Maximum retries for the same URL",
	)
	cmd.Flags().StringVarP(
		&timeout,
		"timeout",
		"t",
		"25m",
		"Timeout for each download",
	)
	cmd.Flags().BoolVarP(
		&tolerant,
		"tolerant",
		"x",
		false,
		"Tolerate download errors by logging them instead of crashing",
	)
	return cmd
}
