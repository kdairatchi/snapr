package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kdairatchi/snapr/internal/capture"
	"github.com/kdairatchi/snapr/internal/config"
	"github.com/spf13/cobra"
)

var (
	snapCfgPath string
	snapOutDir  string
	snapFormat  string
	snapWidth   int
	snapHeight  int
	snapFull    bool
)

var snapCmd = &cobra.Command{
	Use:   "snap [url]",
	Short: "Capture screenshots from a URL or config file",
	Example: `  snapr snap http://localhost:4000
  snapr snap --config snapr.toml
  snapr snap http://localhost:4000 --format pdf --full`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfg *config.Config
		var err error

		if snapCfgPath != "" {
			cfg, err = config.Load(snapCfgPath)
			if err != nil {
				return err
			}
		} else if len(args) == 1 {
			cfg = config.Default()
			cfg.Routes = []config.Route{{URL: args[0], Name: "snap"}}
		} else {
			if _, err := os.Stat("snapr.toml"); err == nil {
				cfg, err = config.Load("snapr.toml")
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("provide a URL or --config path (or add snapr.toml here)")
			}
		}

		// CLI flags override config
		outDir := cfg.Output.Dir
		if snapOutDir != "" {
			outDir = snapOutDir
		}
		format := cfg.Output.Format
		if snapFormat != "" {
			format = snapFormat
		}
		opts := capture.Options{
			Width:    cfg.Output.Width,
			Height:   cfg.Output.Height,
			FullPage: cfg.Output.FullPage,
			Format:   format,
			Timeout:  capture.DefaultOptions().Timeout,
		}
		if snapWidth > 0 {
			opts.Width = snapWidth
		}
		if snapHeight > 0 {
			opts.Height = snapHeight
		}
		if cmd.Flags().Changed("full") {
			opts.FullPage = snapFull
		}

		routes := cfg.AllRoutes()
		if len(routes) == 0 {
			return fmt.Errorf("no routes defined")
		}

		fmt.Printf("snapr: capturing %d route(s) → %s/\n", len(routes), outDir)

		ctx := context.Background()
		var failed []string
		for _, r := range routes {
			result, err := capture.Snap(ctx, r.URL, outDir, r.Name, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  FAIL  %s: %v\n", r.URL, err)
				failed = append(failed, r.URL)
				continue
			}
			fmt.Printf("  OK    %s → %s\n", r.URL, filepath.Base(result.Path))
		}

		if len(failed) > 0 {
			return fmt.Errorf("%d route(s) failed", len(failed))
		}
		fmt.Printf("\ndone: %d screenshot(s) in %s/\n", len(routes), outDir)
		return nil
	},
}

func init() {
	snapCmd.Flags().StringVarP(&snapCfgPath, "config", "c", "", "path to snapr.toml")
	snapCmd.Flags().StringVarP(&snapOutDir, "out", "o", "", "output directory (overrides config)")
	snapCmd.Flags().StringVarP(&snapFormat, "format", "f", "", "output format: png|pdf|webp")
	snapCmd.Flags().IntVar(&snapWidth, "width", 0, "viewport width")
	snapCmd.Flags().IntVar(&snapHeight, "height", 0, "viewport height")
	snapCmd.Flags().BoolVar(&snapFull, "full", false, "capture full page (scroll height)")
	rootCmd.AddCommand(snapCmd)
}
