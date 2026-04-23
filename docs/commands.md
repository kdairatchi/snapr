# Command Reference

## `snapr snap`

Capture screenshots from a URL, a config file, or auto-detect `snapr.toml` in the current directory.

```
snapr snap [url] [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `-c, --config` | — | Path to `snapr.toml` |
| `-o, --out` | (from config) | Output directory |
| `-f, --format` | `png` | Output format: `png`, `pdf`, `webp` |
| `--width` | `1440` | Viewport width |
| `--height` | `900` | Viewport height |
| `--full` | `false` | Capture full scroll height |
| `--workers` | `4` | Concurrent capture workers |
| `--viewports` | — | Comma-separated presets: `mobile,tablet,desktop,wide` |

**Examples:**

```bash
snapr snap http://localhost:4000
snapr snap --config snapr.toml
snapr snap http://localhost:4000 --format pdf --full
snapr snap --config snapr.toml --viewports mobile,desktop --workers 6
```

After capture, writes `screenshots/manifest.json` with route metadata.

---

## `snapr crawl`

Auto-discover routes via link-following or sitemap, then capture them all.

```
snapr crawl <base-url> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `-o, --out` | `screenshots` | Output directory |
| `-f, --format` | `png` | Output format |
| `-m, --max` | `50` | Max pages to crawl |
| `--sitemap` | `false` | Try `/sitemap.xml` first, fall back to crawl |
| `--workers` | `4` | Concurrent capture workers |

**Examples:**

```bash
snapr crawl http://localhost:4000
snapr crawl https://prowlrbot.com --sitemap
snapr crawl http://localhost:3000 --max 20 --workers 8
```

Only follows same-origin links. Off-host URLs in sitemaps are skipped.

---

## `snapr serve`

Start a dev server process, wait for it to be ready, capture screenshots, then shut it down.

```
snapr serve [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--cmd` | (required) | Dev server command |
| `--port` | `4000` | Port to wait for |
| `-c, --config` | — | Path to `snapr.toml` |
| `--wait` | `30` | Seconds to wait for server startup |

**Examples:**

```bash
snapr serve --cmd "hwaro serve" --port 4000
snapr serve --cmd "npm run dev" --port 3000 --config snapr.toml
snapr serve --cmd "hugo server" --port 1313 --wait 10
snapr serve --cmd "python -m http.server 8000" --port 8000
```

The server process is killed after capture regardless of success or failure.

---

## `snapr diff`

Pixel-diff two directories of screenshots. Writes diff images for changed files.

```
snapr diff <dir-a> <dir-b> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--out` | `diff` | Directory to write diff images |
| `--threshold` | `0.1` | Pixelmatch sensitivity (0.0–1.0, lower = stricter) |
| `--fail-on-diff` | `false` | Exit with error if any diff or missing file found |
| `--min-percent` | `0.0` | Suppress output for diffs below this percentage |

**Output:**

```
snapr diff: screenshots/baseline/ vs screenshots/current/

  SAME    home.png
  DIFF    about.png  (142 px, 0.31%)  → diff/about.png
  MISS    contact.png

summary: 1 diff, 1 missing, 1 identical
```

**Examples:**

```bash
snapr diff screenshots/baseline screenshots/current
snapr diff screenshots/v1 screenshots/v2 --fail-on-diff --threshold 0.05
snapr diff screenshots/baseline screenshots/current --min-percent 1.0
```

---

## `snapr report`

Generate a self-contained HTML gallery from a screenshots directory.

```
snapr report [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--dir` | `screenshots` | Directory with screenshots |
| `--out` | `screenshots/report.html` | Output file |
| `--title` | `snapr report` | Page title |
| `--embed` | `true` | Base64-embed images (single portable file) |

**Examples:**

```bash
snapr report
snapr report --title "prowlrbot.com — deploy preview"
snapr report --dir screenshots --out docs/gallery.html --embed=false
```

With `--embed=true` (default), the HTML file is fully self-contained — images are base64-encoded inline. With `--embed=false`, the HTML uses relative paths and must stay alongside the screenshots directory.

If `manifest.json` exists in `--dir`, route URLs are included as metadata under each screenshot card.
