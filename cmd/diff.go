package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kdairatchi/snapr/internal/diff"
	"github.com/spf13/cobra"
)

var (
	diffOutDir     string
	diffThreshold  float64
	diffFailOnDiff bool
	diffMinPercent float64
)

var diffCmd = &cobra.Command{
	Use:   "diff <dir-a> <dir-b>",
	Short: "Compare two screenshot directories and show pixel diffs",
	Example: `  snapr diff screenshots/baseline screenshots/current
  snapr diff screenshots/v1 screenshots/v2 --out diffs --fail-on-diff`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirA := args[0]
		dirB := args[1]

		fmt.Printf("snapr diff: %s/ vs %s/\n\n", dirA, dirB)

		results, err := diff.Compare(dirA, dirB, diffOutDir, diffThreshold)
		if err != nil {
			return err
		}

		var nDiff, nMissing, nSame int
		for _, r := range results {
			switch {
			case r.Missing:
				nMissing++
				fmt.Printf("  MISS    %s\n", r.Name)
			case r.DiffPixels > 0 && r.Percent >= diffMinPercent:
				nDiff++
				fmt.Printf("  DIFF    %-40s (%d px, %.2f%%)  → %s\n",
					r.Name, r.DiffPixels, r.Percent, filepath.ToSlash(r.DiffPath))
			case r.DiffPixels > 0:
				// below --min-percent threshold: treat as same for display
				nSame++
				fmt.Printf("  SAME    %s\n", r.Name)
			default:
				nSame++
				fmt.Printf("  SAME    %s\n", r.Name)
			}
		}

		fmt.Printf("\nsummary: %d diff, %d missing, %d identical\n", nDiff, nMissing, nSame)

		if diffFailOnDiff && (nDiff > 0 || nMissing > 0) {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	diffCmd.Flags().StringVar(&diffOutDir, "out", "diff", "directory to write diff images")
	diffCmd.Flags().Float64Var(&diffThreshold, "threshold", 0.1, "pixelmatch threshold (0.0–1.0)")
	diffCmd.Flags().BoolVar(&diffFailOnDiff, "fail-on-diff", false, "exit 1 if any diff or missing file is found")
	diffCmd.Flags().Float64Var(&diffMinPercent, "min-percent", 0.0, "suppress output for diffs below this percentage")
	rootCmd.AddCommand(diffCmd)
}
