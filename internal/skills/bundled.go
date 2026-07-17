package skills

import (
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
)

//go:embed all:builtin
var bundledFiles embed.FS

var (
	bundledManifestFn = bundledManifest
	bundledWriteFn    = writeBundledFile
	bundledParentsFn  = ensureBundledParents
	bundledReadFile   = fs.ReadFile
	bundledLstat      = os.Lstat
	bundledMkdir      = os.Mkdir
	bundledTemp       = func(dir, pattern string) (bundledTemporaryFile, error) { return os.CreateTemp(dir, pattern) }
	bundledRename     = os.Rename
	bundledRestrict   = config.RestrictPrivatePath
)

type bundledTemporaryFile interface {
	Name() string
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

// InstallBundled materializes the immutable embedded skill set into a
// content-addressed private directory and returns that discovery root. Files
// are rewritten atomically from the binary on every start, so local tampering
// cannot alter application-authored instructions.
func InstallBundled(privateRoot string) (string, error) {
	files, digest, err := bundledManifestFn()
	if err != nil {
		return "", err
	}
	base := filepath.Join(privateRoot, "bundled-skills")
	root := filepath.Join(base, digest)
	for _, directory := range []string{base, root} {
		if err := ensureBundledDirectory(directory); err != nil {
			return "", err
		}
	}
	for _, name := range files {
		data, err := bundledReadFile(bundledFiles, "builtin/"+name)
		if err != nil {
			return "", fmt.Errorf("bundled skill read: %w", err)
		}
		destination := filepath.Join(root, filepath.FromSlash(name))
		if err := bundledParentsFn(root, filepath.Dir(destination)); err != nil {
			return "", err
		}
		if err := bundledWriteFn(destination, data); err != nil {
			return "", err
		}
	}
	return root, nil
}

func bundledManifest() ([]string, string, error) {
	return bundledManifestFrom(bundledFiles, bundledReadFile)
}

func bundledManifestFrom(source fs.FS, read func(fs.FS, string) ([]byte, error)) ([]string, string, error) {
	files := make([]string, 0)
	err := fs.WalkDir(source, "builtin", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return errors.New("bundled skill contains a non-regular entry")
		}
		files = append(files, strings.TrimPrefix(filepath.ToSlash(path), "builtin/"))
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("bundled skill manifest: %w", err)
	}
	sort.Strings(files)
	hash := sha256.New()
	for _, name := range files {
		data, err := read(source, "builtin/"+name)
		if err != nil {
			return nil, "", err
		}
		_, _ = hash.Write([]byte(name + "\x00"))
		_, _ = hash.Write(data)
	}
	return files, fmt.Sprintf("%x", hash.Sum(nil))[:16], nil
}

func ensureBundledDirectory(path string) error {
	info, err := bundledLstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := bundledMkdir(path, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("bundled skill directory: %w", err)
		}
		info, err = bundledLstat(path)
	}
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("bundled skill directory is unsafe")
	}
	return bundledRestrict(path, true)
}

func ensureBundledParents(root, parent string) error {
	relative, err := filepath.Rel(root, parent)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("bundled skill path escapes its root")
	}
	current := root
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		if err := ensureBundledDirectory(current); err != nil {
			return err
		}
	}
	return nil
}

func writeBundledFile(path string, data []byte) error {
	if info, err := bundledLstat(path); err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular()) {
		return errors.New("bundled skill destination is unsafe")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("bundled skill destination: %w", err)
	}
	temporary, err := bundledTemp(filepath.Dir(path), ".bundled-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := bundledRename(temporaryPath, path); err != nil {
		return err
	}
	return bundledRestrict(path, false)
}
