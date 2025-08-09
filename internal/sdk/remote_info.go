package sdk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jkvn/Switchy/internal/config"
	"github.com/mholt/archiver"
	copier "github.com/otiai10/copy"
)

const sdkJsonUrl = "https://raw.githubusercontent.com/jkvn/Switchy/refs/heads/main/sdk/sdkVersions.json"

var (
	httpClient            = &http.Client{Timeout: 60 * time.Second}
	ProgressWriterFactory func(total int64) io.Writer
)

func DownloadSdk(sdkType, version string) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}
	cacheDir := filepath.Join(cfg.DefaultSdkPath, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}
	vers, err := GetSdks(sdkType)
	if err != nil {
		return "", err
	}
	var v *Version
	for i := range vers {
		if vers[i].Version == version {
			v = &vers[i]
			break
		}
	}
	if v == nil {
		return "", fmt.Errorf("version %q for %q not found", version, sdkType)
	}
	ext := archiveExt(v.Link)
	dst := filepath.Join(cacheDir, fmt.Sprintf("%s-%s%s", sdkType, version, ext))
	if fi, err := os.Stat(dst); err == nil && fi.Size() > 0 {
		if v.Sha256 != "" {
			if err := verifySHA256(dst, v.Sha256); err != nil {
				_ = os.Remove(dst)
			} else {
				return dst, nil
			}
		} else {
			return dst, nil
		}
	}
	tmp, err := os.CreateTemp(cacheDir, "dl_*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, v.Link, nil)
	if err != nil {
		_ = tmp.Close()
		return "", err
	}
	req.Header.Set("User-Agent", "Switchy/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = tmp.Close()
		return "", fmt.Errorf("download: http %d", resp.StatusCode)
	}
	w := io.Writer(tmp)
	if ProgressWriterFactory != nil {
		if p := ProgressWriterFactory(resp.ContentLength); p != nil {
			w = io.MultiWriter(tmp, p)
		}
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp: %w", err)
	}
	if v.Sha256 != "" {
		if err := verifySHA256(tmpName, v.Sha256); err != nil {
			return "", err
		}
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return "", fmt.Errorf("finalize download: %w", err)
	}
	return dst, nil
}

func getSdkList() (*SDKList, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, sdkJsonUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Switchy/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch sdk list: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch sdk list: http %d", resp.StatusCode)
	}
	var list SDKList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode sdk list: %w", err)
	}
	return &list, nil
}

func GetSdkTypes() ([]string, error) {
	l, err := getSdkList()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(l.Sdks))
	for _, s := range l.Sdks {
		out = append(out, s.Name)
	}
	return out, nil
}

func GetSdks(typeName string) ([]Version, error) {
	l, err := getSdkList()
	if err != nil {
		return nil, err
	}
	for _, s := range l.Sdks {
		if s.Name == typeName {
			return s.Versions, nil
		}
	}
	return nil, fmt.Errorf("sdk type %q not found", typeName)
}

func ExtractSdk(fileName, sdkType, version string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	finalDir := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, version)
	if err := os.MkdirAll(finalDir, 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}
	tempDir, err := os.MkdirTemp("", "switchy_extract_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	if err := archiver.Unarchive(fileName, tempDir); err != nil {
		base := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
		tmpDec := filepath.Join(os.TempDir(), base)
		if derr := archiver.DecompressFile(fileName, tmpDec); derr != nil {
			return fmt.Errorf("extract: %w", derr)
		}
		defer os.Remove(tmpDec)
		if uerr := archiver.Unarchive(tmpDec, tempDir); uerr != nil {
			dst := filepath.Join(finalDir, filepath.Base(tmpDec))
			return copier.Copy(tmpDec, dst)
		}
	}
	return copier.Copy(tempDir, finalDir, copier.Options{PreserveTimes: true})
}

func archiveExt(link string) string {
	switch {
	case strings.HasSuffix(link, ".tar.gz"):
		return ".tar.gz"
	case strings.HasSuffix(link, ".tar.xz"):
		return ".tar.xz"
	case strings.HasSuffix(link, ".tar.bz2"):
		return ".tar.bz2"
	default:
		return filepath.Ext(link)
	}
}

func verifySHA256(path, hexSum string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(sum, hexSum) {
		return fmt.Errorf("sha256 mismatch")
	}
	return nil
}
