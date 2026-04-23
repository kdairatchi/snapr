package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kdairatchi/snapr/internal/capture"
	"github.com/kdairatchi/snapr/internal/config"
	"github.com/spf13/cobra"
)

var (
	serveCommand string
	servePort    int
	serveCfgPath string
	serveWait    int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a dev server, snap it, then shut it down",
	Example: `  snapr serve --cmd "hwaro serve" --port 4000
  snapr serve --cmd "npm run dev" --port 3000 --config snapr.toml
  snapr serve --cmd "hugo server" --port 1313 --wait 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if serveCommand == "" {
			return fmt.Errorf("--cmd is required")
		}

		var cfg *config.Config
		var err error
		if serveCfgPath != "" {
			cfg, err = config.Load(serveCfgPath)
			if err != nil {
				return err
			}
		} else if _, statErr := os.Stat("snapr.toml"); statErr == nil {
			cfg, err = config.Load("snapr.toml")
			if err != nil {
				return err
			}
		} else {
			cfg = config.Default()
			cfg.Routes = []config.Route{{
				URL:  fmt.Sprintf("http://localhost:%d", servePort),
				Name: "index",
			}}
		}

		parts := strings.Fields(serveCommand)
		if len(parts) == 0 {
			return fmt.Errorf("--cmd produced empty command")
		}
		proc := exec.Command(parts[0], parts[1:]...)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr

		fmt.Printf("snapr serve: starting `%s`\n", serveCommand)
		if err := proc.Start(); err != nil {
			return fmt.Errorf("start server: %w", err)
		}
		defer func() {
			_ = proc.Process.Kill()
			_ = proc.Wait() // reap zombie; ignore error (process already killed)
		}()

		fmt.Printf("waiting for localhost:%d", servePort)
		if err := waitForPort(servePort, time.Duration(serveWait)*time.Second); err != nil {
			return fmt.Errorf("server never came up: %w", err)
		}
		fmt.Println(" ready")

		opts := capture.Options{
			Width:    cfg.Output.Width,
			Height:   cfg.Output.Height,
			FullPage: cfg.Output.FullPage,
			Format:   cfg.Output.Format,
			Timeout:  capture.DefaultOptions().Timeout,
		}

		routes := cfg.AllRoutes()
		if len(routes) == 0 {
			routes = []config.Route{{
				URL:  fmt.Sprintf("http://localhost:%d", servePort),
				Name: "index",
			}}
		}

		fmt.Printf("capturing %d route(s) → %s/\n\n", len(routes), cfg.Output.Dir)

		ctx := context.Background()
		var failed []string
		for _, r := range routes {
			result, err := capture.Snap(ctx, r.URL, cfg.Output.Dir, r.Name, opts)
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
		fmt.Printf("\ndone: %d screenshot(s) in %s/\n", len(routes), cfg.Output.Dir)
		return nil
	},
}

func waitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := fmt.Sprintf("localhost:%d", port)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}
	return fmt.Errorf("timeout after %.0fs", timeout.Seconds())
}

func init() {
	serveCmd.Flags().StringVar(&serveCommand, "cmd", "", "dev server command to run (required)")
	serveCmd.Flags().IntVar(&servePort, "port", 4000, "port to wait for")
	serveCmd.Flags().StringVarP(&serveCfgPath, "config", "c", "", "path to snapr.toml")
	serveCmd.Flags().IntVar(&serveWait, "wait", 30, "seconds to wait for server startup")
	_ = serveCmd.MarkFlagRequired("cmd")
	rootCmd.AddCommand(serveCmd)
}
