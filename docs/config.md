# Configuration

snapr uses a `snapr.toml` file to define output settings, routes, and projects.

Start from the example:

```bash
cp snapr.toml.example snapr.toml
```

---

## `[output]`

Controls where and how screenshots are saved.

| Key | Type | Default | Description |
|---|---|---|---|
| `dir` | string | `screenshots` | Output directory |
| `format` | string | `png` | `png`, `pdf`, or `webp` |
| `width` | int | `1440` | Viewport width in pixels |
| `height` | int | `900` | Viewport height in pixels |
| `full_page` | bool | `false` | Capture full scroll height |

```toml
[output]
dir       = "screenshots"
format    = "png"
width     = 1440
height    = 900
full_page = true
```

---

## `[[routes]]`

Individual URL → name mappings.

| Key | Type | Description |
|---|---|---|
| `url` | string | Full URL to capture |
| `name` | string | Output filename (without extension) |

```toml
[[routes]]
url  = "http://localhost:4000"
name = "home"

[[routes]]
url  = "http://localhost:4000/about"
name = "about"

[[routes]]
url  = "http://localhost:4000/blog"
name = "blog"
```

Route names are sanitized — path separators and special characters are replaced with `-`. Names like `../../etc/passwd` are collapsed to `passwd` via `filepath.Base`.

---

## `[[projects]]`

Shorthand for a base URL + list of paths. Useful for capturing an entire deployed site.

| Key | Type | Description |
|---|---|---|
| `name` | string | Prefix for all output filenames |
| `base` | string | Base URL (no trailing slash) |
| `routes` | []string | Path list |

```toml
[[projects]]
name   = "prowlrbot"
base   = "https://prowlrbot.com"
routes = ["/", "/tools", "/about", "/blog"]
```

This produces: `prowlrbot-index.png`, `prowlrbot-tools.png`, `prowlrbot-about.png`, `prowlrbot-blog.png`.

---

## CLI flag overrides

All output settings can be overridden at the command line without editing the config:

```bash
snapr snap --config snapr.toml \
  --out dist/screenshots \
  --format pdf \
  --width 1280 \
  --height 800 \
  --full \
  --workers 6
```

Flags always take precedence over config values.

---

## Full example

```toml
[output]
dir       = "screenshots"
format    = "png"
width     = 1440
height    = 900
full_page = true

[[routes]]
url  = "http://localhost:4000"
name = "home"

[[routes]]
url  = "http://localhost:4000/about"
name = "about"

[[routes]]
url  = "http://localhost:4000/blog"
name = "blog"

[[projects]]
name   = "live"
base   = "https://prowlrbot.com"
routes = ["/", "/tools", "/about"]
```
