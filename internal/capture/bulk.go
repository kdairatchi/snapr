package capture

import (
	"context"
	"sync"
)

type Job struct {
	URL          string
	Name         string
	ViewportName string // when set, overrides opts Width/Height with Viewports[ViewportName]
}

type BulkResult struct {
	Result *Result
	Err    error
	Job    Job
}

// BulkSnap runs Snap concurrently with a worker pool. workers=1 is sequential.
// Results preserve input order.
func BulkSnap(ctx context.Context, jobs []Job, outDir string, opts Options, workers int) []BulkResult {
	if workers < 1 {
		workers = 1
	}
	results := make([]BulkResult, len(jobs))
	jobCh := make(chan int, len(jobs))
	for i := range jobs {
		jobCh <- i
	}
	close(jobCh)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobCh {
				j := jobs[i]
				r, err := Snap(ctx, j.URL, outDir, j.Name, opts)
				results[i] = BulkResult{Result: r, Err: err, Job: j}
			}
		}()
	}
	wg.Wait()
	return results
}
