package config

import (
	"fmt"
	"os"

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
	out = append(out, c.Routes...)
	for _, p := range c.Projects {
		for _, r := range p.Routes {
			out = append(out, Route{
				URL:  p.Base + r,
				Name: p.Name + "-" + sanitize(r),
			})
		}
	}
	return out
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
