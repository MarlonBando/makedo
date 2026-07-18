package engine

import (
	"os"
	"sync"
	"syscall"
)

// Status represents command execution outcome.
type Status int

const (
	Completed Status = iota // Command exited
	Ready                   // Directives passed, still running
)

// Result holds command execution result.
type Result struct {
	Status   Status
	Output   []byte      // Captured stdout+stderr (raw)
	CleanOut []byte      // Cleaned output without ANSI and \r
	ExitCode int         // -1 if still running or error
	Process  *os.Process // For cleanup if still running
	Err      error       // Start/pipe error (not exit failure)
	Warnings []error     // Non-fatal warnings encountered during execution
}

// Registry tracks processes for cleanup at document end.
type Registry struct {
	mu        sync.Mutex
	processes []*os.Process
}

// NewRegistry creates a process registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Add registers a process for cleanup.
func (r *Registry) Add(p *os.Process) {
	if p == nil {
		return
	}
	r.mu.Lock()
	r.processes = append(r.processes, p)
	r.mu.Unlock()
}

// KillAll terminates all registered processes and their children.
func (r *Registry) KillAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.processes {
		if p != nil {
			// Kill entire process group (negative PID)
			_ = syscall.Kill(-p.Pid, syscall.SIGKILL)
			_, _ = p.Wait()
		}
	}
	r.processes = nil
}
