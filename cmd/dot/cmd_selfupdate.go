package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func cmdSelfUpdate() error {
	fmt.Println(mutedStyle.Render("Checking for updates..."))

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("fetch latest version: %w", err)
	}

	current := strings.TrimPrefix(buildVersion, "v")
	latestClean := strings.TrimPrefix(latest, "v")

	if current == latestClean {
		fmt.Printf("%s  Already on %s — nothing to do.\n", successStyle.Render("✓"), buildVersion)
		return nil
	}

	if buildVersion == "dev" {
		fmt.Println(mutedStyle.Render("dev build — skipping version check, downloading latest anyway"))
	} else {
		fmt.Printf("Updating %s → %s\n", mutedStyle.Render(buildVersion), successStyle.Render("v"+latestClean))
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate current executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolve symlink: %w", err)
	}

	url := archiveURL(latestClean)
	fmt.Println(mutedStyle.Render("Downloading " + url))

	if err := downloadAndReplace(url, exe); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}

	fmt.Printf("%s  dot updated to v%s\n", successStyle.Render("✓"), latestClean)
	return nil
}

// fetchLatestVersion queries the GitHub Releases API and returns the tag name
// of the latest release (e.g. "v0.2.0").
func fetchLatestVersion() (string, error) {
	req, err := http.NewRequest(http.MethodGet,
		"https://api.github.com/repos/version14/dot/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}
	return release.TagName, nil
}

// archiveURL builds the download URL for the current OS/arch.
// Matches the name_template in .goreleaser.yaml: dot_VERSION_OS_ARCH.tar.gz
func archiveURL(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf(
		"https://github.com/version14/dot/releases/download/v%s/dot_%s_%s_%s.%s",
		version, version, goos, goarch, ext,
	)
}

// downloadAndReplace downloads the archive at url, extracts the "dot" binary,
// and atomically replaces the file at dest.
func downloadAndReplace(url, dest string) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Write to a temp file in the same directory so os.Rename is atomic.
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".dot-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // clean up on any error path

	// Extract the binary from the archive.
	var extractErr error
	if strings.HasSuffix(url, ".zip") {
		extractErr = extractFromZip(resp.Body, tmp)
	} else {
		extractErr = extractFromTarGz(resp.Body, tmp)
	}
	if extractErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("extract binary: %w", extractErr)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Match the permissions of the current executable.
	info, err := os.Stat(dest)
	if err != nil {
		return fmt.Errorf("stat current binary: %w", err)
	}
	if err := os.Chmod(tmpName, info.Mode()); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}

	// Atomic replace.
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

// extractFromTarGz finds and copies the "dot" (or "dot.exe") entry from a
// .tar.gz archive into dst.
func extractFromTarGz(r io.Reader, dst io.Writer) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		base := filepath.Base(hdr.Name)
		if base == "dot" || base == "dot.exe" {
			_, err = io.Copy(dst, tr) //nolint:gosec
			return err
		}
	}
	return fmt.Errorf("binary not found in archive")
}

// extractFromZip finds and copies the "dot.exe" entry from a .zip archive
// into dst. The zip must be fully buffered — we write it to a temp file first.
func extractFromZip(r io.Reader, dst io.Writer) error {
	// zip.NewReader requires io.ReaderAt + size, so buffer to a temp file.
	tmp, err := os.CreateTemp("", ".dot-zip-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	size, err := io.Copy(tmp, r) //nolint:gosec
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(tmp, size)
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		base := filepath.Base(f.Name)
		if base == "dot" || base == "dot.exe" {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(dst, rc) //nolint:gosec
			_ = rc.Close()
			return err
		}
	}
	return fmt.Errorf("binary not found in zip archive")
}
