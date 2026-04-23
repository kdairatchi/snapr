package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
)

type Output struct {
	Dir      string `toml:"dir"`
	Format   string `toml:"format"`
	Width    int    `toml:"width"`
	Height   int    `toml:"height"`
	FullPage bool   `toml:"full_page"`
}

type Route struct {
	URL  string `toml:"url"`
	Name string `toml:"name"`
}

type Project struct {
	Name   string   `toml:"name"`
	Base   string   `toml:"base"`
	Routes []string `toml:"routes"`
}

type Config struct {
	Output   Output    `toml:"output"`
	Routes   []Route   `toml:"routes"`
	Projects []Project `toml:"projects"`
}

func (c *Config) AllRoutes() []Route {
	var out []Route
	for _, r := range c.Routes {
		out = append(out, Route{URL: r.URL, Name: safeName(r.Name)})
	}
	for _, p := range c.Projects {
		base := safeName(p.Name)
		for _, r := range p.Routes {
			out = append(out, Route{
				URL:  p.Base + r,
				Name: base + "-" + safeName(sanitize(r)),
			})
		}
	}
	return out
}

var unsafeNameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// safeName strips path traversal and non-filesystem-safe characters from a
// user-supplied route or project name before using it as an output filename.
func safeName(s string) string {
	s = filepath.Base(s) // collapse any directory components (e.g. ../../etc)
	s = unsafeNameRe.ReplaceAllString(s, "-")
	if s == "" || s == "." || s == ".." {
		return "unnamed"
	}
	return s
}

func sanitize(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '/' || c == '?' || c == '=' || c == '&' {
			result = append(result, '-')
		} else {
			result = append(result, c)
		}
	}
	if len(result) > 0 && result[0] == '-' {
		result = result[1:]
	}
	if len(result) == 0 {
		return "index"
	}
	return string(result)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	cfg.Output = Output{
		Dir:    "screenshots",
		Format: "png",
		Width:  1440,
		Height: 900,
	}
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func Default() *Config {
	return &Config{
		Output: Output{
			Dir:    "screenshots",
			Format: "png",
			Width:  1440,
			Height: 900,
		},
	}
}
