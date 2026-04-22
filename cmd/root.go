package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snapr",
	Short: "Screenshot any web project — local, live, or CI",
	Long: `snapr captures full-page screenshots and PDFs of web projects.

Works with local dev servers, deployed sites, and CI pipelines.
Configure routes in snapr.toml or pass a URL directly.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
