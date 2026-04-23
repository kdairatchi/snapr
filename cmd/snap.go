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
	snapCfgPath   string
	snapOutDir    string
	snapFormat    string
	snapWidth     int
	snapHeight    int
	snapFull      bool
	snapWorkers   int
	snapViewports string
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

		// Build jobs, optionally expanding per viewport.
		// Use ParseViewportNames (not ParseViewports) to preserve input order and carry names.
		var jobs []capture.Job
		vpNames := capture.ParseViewportNames(snapViewports)
		if len(vpNames) > 0 {
			for _, r := range routes {
				for _, vpName := range vpNames { // ordered — no map iteration
					jobs = append(jobs, capture.Job{
						URL:          r.URL,
						Name:         r.Name + "-" + vpName,
						ViewportName: vpName,
					})
				}
			}
		} else {
			for _, r := range routes {
				jobs = append(jobs, capture.Job{URL: r.URL, Name: r.Name})
			}
		}

		fmt.Printf("snapr: capturing %d job(s) → %s/\n", len(jobs), outDir)

		ctx := context.Background()

		// Viewport jobs each need their own opts; run sequentially with per-job opts.
		// Non-viewport jobs use BulkSnap's worker pool.
		var results []capture.BulkResult
		if len(vpNames) > 0 {
			results = make([]capture.BulkResult, len(jobs))
			for i, j := range jobs {
				jobOpts := opts
				if vp, ok := capture.Viewports[j.ViewportName]; ok {
					jobOpts.Width = vp[0]
					jobOpts.Height = vp[1]
				}
				r, err := capture.Snap(ctx, j.URL, outDir, j.Name, jobOpts)
				results[i] = capture.BulkResult{Result: r, Err: err, Job: j}
			}
		} else {
			results = capture.BulkSnap(ctx, jobs, outDir, opts, snapWorkers)
		}

		var failed []string
		for _, br := range results {
			if br.Err != nil {
				fmt.Fprintf(os.Stderr, "  FAIL  %s: %v\n", br.Job.URL, br.Err)
				failed = append(failed, br.Job.URL)
				continue
			}
			if br.Result != nil {
				fmt.Printf("  OK    %s → %s\n", br.Job.URL, filepath.Base(br.Result.Path))
			}
		}

		manifestPath := filepath.Join(outDir, "manifest.json")
		if err := capture.WriteManifest(outDir, results); err != nil {
			fmt.Fprintf(os.Stderr, "manifest: %v\n", err)
		} else {
			fmt.Printf("\nmanifest: %s\n", manifestPath)
		}

		if len(failed) > 0 {
			return fmt.Errorf("%d route(s) failed", len(failed))
		}
		fmt.Printf("done: %d screenshot(s) in %s/\n", len(jobs), outDir)
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
	snapCmd.Flags().IntVar(&snapWorkers, "workers", 4, "number of concurrent capture workers")
	snapCmd.Flags().StringVar(&snapViewports, "viewports", "", "comma-separated viewports: mobile,tablet,desktop,wide")
	rootCmd.AddCommand(snapCmd)
}
