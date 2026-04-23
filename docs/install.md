# Installation

## Requirements

- Chrome or Chromium must be installed on the machine running snapr.
- On CI (GitHub Actions, etc.) install with `apt-get install -y chromium-browser` or use a runner image that includes Chrome.

## Go install

```bash
go install github.com/kdairatchi/snapr@latest
```

Adds `snapr` to `$(go env GOPATH)/bin`. Make sure that's in your `$PATH`.

## Homebrew (macOS / Linux)

```bash
brew install kdairatchi/tap/snapr
```

## Prebuilt binary

Download the archive for your platform from [Releases](https://github.com/kdairatchi/snapr/releases):

| Platform | Archive |
|---|---|
| Linux x86_64 | `snapr_linux_amd64.tar.gz` |
| Linux arm64 | `snapr_linux_arm64.tar.gz` |
| macOS x86_64 | `snapr_darwin_amd64.tar.gz` |
| macOS arm64 (M-series) | `snapr_darwin_arm64.tar.gz` |
| Windows x86_64 | `snapr_windows_amd64.zip` |

Verify the checksum before extracting:

```bash
sha256sum -c checksums.txt
```

Extract and move to your PATH:

```bash
tar -xz snapr < snapr_linux_amd64.tar.gz
sudo mv snapr /usr/local/bin/
```

## GitHub Actions

The action installs snapr automatically — no manual install needed:

```yaml
- uses: kdairatchi/snapr@v1
```

It downloads the binary, verifies the sha256 checksum against the release `checksums.txt`, and adds it to `$PATH`.

## Build from source

```bash
git clone https://github.com/kdairatchi/snapr
cd snapr
go build -o snapr .
```

Requires Go 1.22+.
