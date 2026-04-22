package capture

import "strings"

var Viewports = map[string][2]int{
	"mobile":  {375, 812},
	"tablet":  {768, 1024},
	"desktop": {1440, 900},
	"wide":    {1920, 1080},
}

// ParseViewports splits "mobile,desktop" into viewport sizes.
// Unknown names are skipped. Empty string returns nil.
func ParseViewports(s string) [][2]int {
	if s == "" {
		return nil
	}
	var out [][2]int
	for _, name := range strings.Split(s, ",") {
		name = strings.TrimSpace(name)
		if v, ok := Viewports[name]; ok {
			out = append(out, v)
		}
	}
	return out
}
