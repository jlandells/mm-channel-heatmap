package progress

import (
	"fmt"
	"os"
	"sync"
)

// Progress displays an in-place progress indicator on stderr.
type Progress struct {
	mu      sync.Mutex
	total   int
	count   int
	verbose bool
	isTTY   bool
	cleared bool
}

// New creates a new progress indicator.
// If verbose is true, progress display is suppressed (verbose logging takes precedence).
func New(verbose bool) *Progress {
	isTTY := false
	if fi, err := os.Stderr.Stat(); err == nil {
		isTTY = fi.Mode()&os.ModeCharDevice != 0
	}
	return &Progress{
		verbose: verbose,
		isTTY:   isTTY,
	}
}

// SetTotal sets the total number of channels to process.
func (p *Progress) SetTotal(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
}

// Update reports progress on a channel being processed.
func (p *Progress) Update(channelName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.verbose {
		return
	}

	p.count++

	if !p.isTTY {
		return
	}

	fmt.Fprintf(os.Stderr, "\r  Processing channel %d/%d: %-60s", p.count, p.total, channelName)
}

// Clear removes the progress line from stderr.
func (p *Progress) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cleared || p.verbose || !p.isTTY {
		return
	}
	p.cleared = true

	// Clear the line
	fmt.Fprintf(os.Stderr, "\r%80s\r", "")
}
