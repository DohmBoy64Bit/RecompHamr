package tools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type cacheTemporaryFile interface {
	Name() string
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

var (
	cacheLstat      = os.Lstat
	cacheMkdirAll   = os.MkdirAll
	cacheCreateTemp = func(dir, pattern string) (cacheTemporaryFile, error) { return os.CreateTemp(dir, pattern) }
	cacheRename     = os.Rename
	cacheReadFile   = os.ReadFile
	cacheRemoveAll  = os.RemoveAll
)

func preparePrivateDir(path string, restrict func(string, bool) error) error {
	if err := refuseExistingSpecial(path); err != nil {
		return err
	}
	if err := cacheMkdirAll(path, 0o700); err != nil {
		return err
	}
	return restrict(path, true)
}

func refuseExistingSpecial(path string) error {
	info, err := cacheLstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refuses link or reparse point at %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("expected directory at %s", path)
	}
	return nil
}

func atomicPrivateWrite(path string, data []byte, restrict func(string, bool) error) error {
	if info, err := cacheLstat(path); err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular()) {
		return fmt.Errorf("refuses non-regular cache file at %s", path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	tmp, err := cacheCreateTemp(filepath.Dir(path), ".recomphamr-cache-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = restrict(tmpPath, false); err != nil {
		return err
	}
	if err = cacheRename(tmpPath, path); err != nil {
		return err
	}
	return restrict(path, false)
}

func readOptionalRegular(path string, limit int64) ([]byte, error) {
	info, err := cacheLstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("optional file is unsafe or too large")
	}
	return cacheReadFile(path)
}

func humanSize(bytes int64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
