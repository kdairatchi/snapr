package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Options struct {
	Width    int
	Height   int
	FullPage bool
	Format   string // png | pdf | webp
	Timeout  time.Duration
}

type Result struct {
	Path string
	URL  string
	Name string
}

func DefaultOptions() Options {
	return Options{
		Width:   1440,
		Height:  900,
		Format:  "png",
		Timeout: 30 * time.Second,
	}
}

func Snap(ctx context.Context, url, outDir, name string, opts Options) (*Result, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", outDir, err)
	}

	ext := strings.ToLower(opts.Format)
	if ext == "pdf" {
		return snapPDF(ctx, url, outDir, name, opts)
	}
	return snapImage(ctx, url, outDir, name, opts, ext)
}

func snapImage(ctx context.Context, url, outDir, name string, opts Options, ext string) (*Result, error) {
	allocCtx, cancel := chromedp.NewExecAllocator(ctx,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.WindowSize(opts.Width, opts.Height),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, opts.Timeout)
	defer cancel()

	var buf []byte
	tasks := chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(
			int64(opts.Width), int64(opts.Height), 1, false,
		),
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
	}

	if opts.FullPage {
		tasks = append(tasks, chromedp.FullScreenshot(&buf, 100))
	} else {
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
	}

	if err := chromedp.Run(taskCtx, tasks...); err != nil {
		return nil, fmt.Errorf("chromedp %s: %w", url, err)
	}

	outPath := filepath.Join(outDir, name+"."+ext)
	if err := os.WriteFile(outPath, buf, 0o644); err != nil {
		return nil, fmt.Errorf("write %s: %w", outPath, err)
	}
	return &Result{Path: outPath, URL: url, Name: name}, nil
}

func snapPDF(ctx context.Context, url, outDir, name string, opts Options) (*Result, error) {
	allocCtx, cancel := chromedp.NewExecAllocator(ctx,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.WindowSize(opts.Width, opts.Height),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, opts.Timeout)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("chromedp pdf %s: %w", url, err)
	}

	outPath := filepath.Join(outDir, name+".pdf")
	if err := os.WriteFile(outPath, buf, 0o644); err != nil {
		return nil, fmt.Errorf("write %s: %w", outPath, err)
	}
	return &Result{Path: outPath, URL: url, Name: name}, nil
}
