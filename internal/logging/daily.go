package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

// dailyRotator writes log lines to a per-day file:
//   logs/app-2006-01-02.log
//
// It checks the date on every Write and opens a new file when the day
// rolls over, closing the previous one.
type dailyRotator struct {
	mu       sync.Mutex
	dir      string
	prefix   string
	file     *os.File
	today    string
}

func newDailyRotator(cfg config.Config) *dailyRotator {
	dir := filepath.Dir(cfg.LogFilePath)
	base := filepath.Base(cfg.LogFilePath)
	ext := filepath.Ext(base)
	prefix := base[:len(base)-len(ext)]

	return &dailyRotator{
		dir:    dir,
		prefix: prefix,
	}
}

func (r *dailyRotator) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	today := now.Format("2006-01-02")

	if r.file == nil || r.today != today {
		if err := r.rotate(today); err != nil {
			return 0, err
		}
	}

	return r.file.Write(p)
}

func (r *dailyRotator) rotate(today string) error {
	if r.file != nil {
		_ = r.file.Close()
		r.file = nil
	}

	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	name := filepath.Join(r.dir, fmt.Sprintf("%s-%s.log", r.prefix, today))
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	r.file = f
	r.today = today
	return nil
}

func (r *dailyRotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		err := r.file.Close()
		r.file = nil
		return err
	}
	return nil
}