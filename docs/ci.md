# CI / CD

## GitHub Action

snapr ships as a reusable GitHub Action. Add it to any workflow:

```yaml
- name: Screenshot
  uses: kdairatchi/snapr@v1
  with:
    config: snapr.toml
    upload: true
```

### Inputs

| Input | Default | Description |
|---|---|---|
| `config` | `snapr.toml` | Path to config file |
| `url` | — | Single URL (use instead of config) |
| `format` | `png` | `png`, `pdf`, or `webp` |
| `out` | `screenshots` | Output directory |
| `full` | `true` | Full-page capture |
| `upload` | `true` | Upload as workflow artifact (30-day retention) |
| `version` | `latest` | snapr version to pin |

### Outputs

| Output | Description |
|---|---|
| `screenshots-dir` | Path to the directory containing screenshots |

---

## Recipes

### Capture on every push

```yaml
name: screenshots
on: [push]

jobs:
  snap:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: kdairatchi/snapr@v1
        with:
          config: snapr.toml
          upload: true
```

### Visual regression — fail on diff

```yaml
name: visual-regression
on: [pull_request]

jobs:
  diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      # restore baseline screenshots from main
      - uses: actions/download-artifact@v4
        with:
          name: snapr-baseline
          path: screenshots/baseline
        continue-on-error: true

      # capture current branch
      - uses: kdairatchi/snapr@v1
        with:
          config: snapr.toml
          out: screenshots/current
          upload: false

      # diff and fail if anything changed
      - run: snapr diff screenshots/baseline screenshots/current --fail-on-diff

      # update baseline on main
      - if: github.ref == 'refs/heads/main'
        uses: actions/upload-artifact@v4
        with:
          name: snapr-baseline
          path: screenshots/current/
```

### Capture a live deployed site after deploy

```yaml
- name: Wait for deployment
  run: sleep 30

- name: Screenshot deployed site
  uses: kdairatchi/snapr@v1
  with:
    url: https://yoursite.com
    format: png
    full: true
    upload: true
```

### Generate and publish an HTML report

```yaml
- uses: kdairatchi/snapr@v1
  with:
    config: snapr.toml
    upload: false

- run: snapr report --title "Deploy preview — ${{ github.sha }}"

- uses: actions/upload-artifact@v4
  with:
    name: snapr-report
    path: screenshots/report.html
```

---

## Other CI systems

snapr is a single binary — it works in any CI environment.

**GitLab CI:**

```yaml
screenshots:
  image: golang:1.22
  before_script:
    - apt-get install -y chromium
    - go install github.com/kdairatchi/snapr@latest
  script:
    - snapr snap --config snapr.toml
  artifacts:
    paths:
      - screenshots/
```

**CircleCI:**

```yaml
- run:
    name: Install snapr
    command: go install github.com/kdairatchi/snapr@latest
- run:
    name: Capture screenshots
    command: snapr snap --config snapr.toml
- store_artifacts:
    path: screenshots
```
