// Package session owns application-session configuration, client, probe, and
// private prompt-history lifecycle outside presentation.
package session

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
)

type historyAppender interface {
	WriteString(string) (int, error)
	Close() error
}

var (
	openHistoryAppend = func(path string) (historyAppender, error) {
		return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	}
	restrictHistoryPath = config.RestrictPrivatePath
	countStoredHistory  = countHistoryLines
	removeHistoryPath   = os.Remove
)

const (
	historyFileName      = "history"
	historyMaxEntries    = 500
	historyMaxEntryBytes = 256 * 1024
	historyScannerMax    = 1024 * 1024
)

// History owns the private, per-project prompt-recall file. Its value is safe
// to copy because it contains only the immutable private-state directory path.
type History struct {
	dir string
}

// NewHistory constructs a prompt-history store rooted in an already secured
// `.rehamr` directory. It performs no I/O until Load, Append, or Clear.
func NewHistory(dir string) History { return History{dir: dir} }

func (h History) path() string { return filepath.Join(h.dir, historyFileName) }

// Load returns every decodable saved prompt oldest-first. Missing or unreadable
// files produce an empty history, and malformed individual lines are skipped.
func (h History) Load() []string {
	f, err := os.Open(h.path())
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), historyScannerMax)
	for sc.Scan() {
		value, err := strconv.Unquote(sc.Text())
		if err != nil {
			continue
		}
		out = append(out, value)
	}
	return out
}

// Append persists one prompt using the retained quoted-line format and lazy
// newest-500 trim. Empty, over-limit, or scanner-incompatible values are
// intentionally omitted; an append or protection failure is returned.
func (h History) Append(value string) error {
	if value == "" || len(value) > historyMaxEntryBytes {
		return nil
	}
	quoted := strconv.Quote(value)
	if len(quoted) >= historyScannerMax {
		return nil
	}
	f, err := openHistoryAppend(h.path())
	if err != nil {
		return err
	}
	if err := restrictHistoryPath(h.path(), false); err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.WriteString(quoted + "\n"); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	count, err := countStoredHistory(h.path())
	if err != nil || count <= historyMaxEntries {
		return nil
	}
	all := h.Load()
	if len(all) > historyMaxEntries {
		all = all[len(all)-historyMaxEntries:]
	}
	var buf []byte
	for _, entry := range all {
		buf = append(buf, strconv.Quote(entry)...)
		buf = append(buf, '\n')
	}
	_ = os.WriteFile(h.path(), buf, 0o600)
	return nil
}

func countHistoryLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), historyScannerMax)
	count := 0
	for sc.Scan() {
		count++
	}
	return count, sc.Err()
}

// Clear removes persisted recall. A missing history file already satisfies the
// empty-history intent and is therefore not an error.
func (h History) Clear() error {
	err := removeHistoryPath(h.path())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
