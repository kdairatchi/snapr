package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kdairatchi/snapr/internal/capture"
	"github.com/kdairatchi/snapr/internal/config"
	"github.com/kdairatchi/snapr/internal/routes"
	"github.com/spf13/cobra"
)

var (
	crawlOutDir  string
	crawlFormat  string
	crawlMax     int
	crawlSitemap bool
	crawlWorkers int
)

var crawlCmd = &cobra.Command{
	Use:   "crawl <base-url>",
	Short: "Auto-discover routes then screenshot them all",
	Example: `  snapr crawl http://localhost:4000
  snapr crawl https://prowlrbot.com --sitemap
  snapr crawl http://localhost:3000 --max 20`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		base := args[0]
		outDir := crawlOutDir
		if outDir == "" {
			outDir = "screenshots"
		}
		format := crawlFormat
		if format == "" {
			format = "png"
		}

		fmt.Printf("snapr crawl: discovering routes at %s\n", base)

		ctx := context.Background()
		var discovered []routes.Route
		var err error

		if crawlSitemap {
			discovered, err = routes.FromSitemap(base)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sitemap failed (%v), falling back to crawl\n", err)
			}
		}
		if len(discovered) == 0 {
			discovered, err = routes.Crawl(ctx, base, crawlMax)
			if err != nil {
				return fmt.Errorf("crawl failed: %w", err)
			}
		}

		fmt.Printf("found %d route(s)\n\n", len(discovered))

		cfg := config.Default()
		opts := capture.Options{
			Width:    cfg.Output.Width,
			Height:   cfg.Output.Height,
			FullPage: true,
			Format:   format,
			Timeout:  capture.DefaultOptions().Timeout,
		}

		jobs := make([]capture.Job, len(discovered))
		for i, r := range discovered {
			jobs[i] = capture.Job{URL: r.URL, Name: r.Name}
		}

		results := capture.BulkSnap(ctx, jobs, outDir, opts, crawlWorkers)

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
		fmt.Printf("done: %d screenshot(s) in %s/\n", len(discovered), outDir)
		return nil
	},
}

func init() {
	crawlCmd.Flags().StringVarP(&crawlOutDir, "out", "o", "screenshots", "output directory")
	crawlCmd.Flags().StringVarP(&crawlFormat, "format", "f", "png", "output format: png|pdf|webp")
	crawlCmd.Flags().IntVarP(&crawlMax, "max", "m", 50, "max pages to crawl")
	crawlCmd.Flags().BoolVar(&crawlSitemap, "sitemap", false, "try /sitemap.xml first")
	crawlCmd.Flags().IntVar(&crawlWorkers, "workers", 4, "number of concurrent capture workers")
	rootCmd.AddCommand(crawlCmd)
}
