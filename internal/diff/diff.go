package diff

import (
	"fmt"
	"image"
	"image/png"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/orisano/pixelmatch"
)

// Result holds the comparison outcome for a single image pair.
type Result struct {
	Name       string
	DiffPixels int
	Total      int
	Percent    float64
	DiffPath   string // path to diff image, empty if identical
	Missing    bool   // present in one directory but not the other
}

func isImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return img, nil
}

func writeDiffPNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// Compare reads image files from dirA, matches each by name to dirB,
// runs pixelmatch on each pair, and writes diff PNGs to outDir.
// threshold is a perceptual sensitivity value between 0.0 and 1.0.
func Compare(dirA, dirB, outDir string, threshold float64) ([]Result, error) {
	entriesA, err := os.ReadDir(dirA)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dirA, err)
	}

	// Build a set of image names in dirB for fast lookup and missing-from-A detection.
	entriesB, err := os.ReadDir(dirB)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dirB, err)
	}
	inB := make(map[string]struct{}, len(entriesB))
	for _, e := range entriesB {
		if !e.IsDir() && isImage(e.Name()) {
			inB[e.Name()] = struct{}{}
		}
	}

	var results []Result

	// Process files present in dirA.
	inA := make(map[string]struct{}, len(entriesA))
	for _, e := range entriesA {
		if e.IsDir() || !isImage(e.Name()) {
			continue
		}
		name := e.Name()
		inA[name] = struct{}{}

		pathA := filepath.Join(dirA, name)
		pathB := filepath.Join(dirB, name)

		if _, ok := inB[name]; !ok {
			results = append(results, Result{Name: name, Missing: true})
			continue
		}

		imgA, err := loadImage(pathA)
		if err != nil {
			return nil, err
		}
		imgB, err := loadImage(pathB)
		if err != nil {
			return nil, err
		}

		boundsA := imgA.Bounds()
		boundsB := imgB.Bounds()
		total := boundsA.Dx() * boundsA.Dy()

		// Dimensions mismatch — treat every pixel as different.
		if !boundsA.Eq(boundsB) {
			results = append(results, Result{
				Name:       name,
				DiffPixels: total,
				Total:      total,
				Percent:    100.0,
			})
			continue
		}

		var diffImg image.Image
		diffPixels, err := pixelmatch.MatchPixel(
			imgA, imgB,
			pixelmatch.Threshold(threshold),
			pixelmatch.WriteTo(&diffImg),
		)
		if err != nil {
			return nil, fmt.Errorf("pixelmatch %s: %w", name, err)
		}

		var percent float64
		if total > 0 {
			percent = float64(diffPixels) / float64(total) * 100.0
		}

		r := Result{
			Name:       name,
			DiffPixels: diffPixels,
			Total:      total,
			Percent:    percent,
		}

		if diffPixels > 0 && diffImg != nil {
			// Write diff image with .png extension regardless of source format.
			base := strings.TrimSuffix(name, filepath.Ext(name)) + ".png"
			diffPath := filepath.Join(outDir, base)
			if err := writeDiffPNG(diffPath, diffImg); err != nil {
				return nil, fmt.Errorf("write diff image %s: %w", diffPath, err)
			}
			r.DiffPath = diffPath
		}

		results = append(results, r)
	}

	// Collect files present in dirB but not in dirA.
	for _, e := range entriesB {
		if e.IsDir() || !isImage(e.Name()) {
			continue
		}
		if _, ok := inA[e.Name()]; !ok {
			results = append(results, Result{Name: e.Name(), Missing: true})
		}
	}

	return results, nil
}
