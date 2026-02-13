package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nerifect/nerifect-cli/internal/config"
	"github.com/nerifect/nerifect-cli/internal/store"
)

// Status holds a snapshot of the daemon's running state.
type Status struct {
	Running      bool       `json:"running"`
	PID          int        `json:"pid,omitempty"`
	Interval     int        `json:"interval_hours"`
	SourceCount  int        `json:"source_count"`
	ErrorCount   int        `json:"error_count"`
	LastCheckAt  *time.Time `json:"last_check_at,omitempty"`
	NextCheckAt  *time.Time `json:"next_check_at,omitempty"`
}

// PidPath returns the path to the daemon PID file.
func PidPath(cfg *config.Config) string {
	return filepath.Join(cfg.DataDir, "agent.pid")
}

// LogPath returns the path to the daemon log file.
func LogPath(cfg *config.Config) string {
	return filepath.Join(cfg.DataDir, "agent.log")
}

// IsRunning checks if a daemon is running by reading the PID file and probing the process.
func IsRunning(pidPath string) (int, bool) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}

	// Signal 0 checks if process exists without affecting it.
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return pid, false
	}

	return pid, true
}

// StartDaemon launches the daemon as a detached background process.
func StartDaemon(cfg *config.Config) error {
	pidPath := PidPath(cfg)
	if pid, running := IsRunning(pidPath); running {
		return fmt.Errorf("agent is already running (PID %d)", pid)
	}

	// Find our own binary path for re-exec.
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable: %w", err)
	}

	logFile, err := os.OpenFile(LogPath(cfg), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}

	cmd := exec.Command(binPath, "_run-daemon")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("starting daemon: %w", err)
	}
	logFile.Close()

	// Write PID file.
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0600); err != nil {
		return fmt.Errorf("writing PID file: %w", err)
	}

	return nil
}

// StopDaemon sends SIGTERM to the running daemon, waits, then SIGKILL if necessary.
func StopDaemon(cfg *config.Config) error {
	pidPath := PidPath(cfg)
	pid, running := IsRunning(pidPath)
	if !running {
		// Clean up stale PID file if it exists.
		os.Remove(pidPath)
		return fmt.Errorf("agent is not running")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process: %w", err)
	}

	// Send SIGTERM.
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		os.Remove(pidPath)
		return fmt.Errorf("sending SIGTERM: %w", err)
	}

	// Poll for exit (up to 3 seconds).
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			os.Remove(pidPath)
			return nil
		}
	}

	// Force kill.
	_ = proc.Signal(syscall.SIGKILL)
	time.Sleep(200 * time.Millisecond)
	os.Remove(pidPath)
	return nil
}

// GetStatus returns the current daemon status and source statistics.
func GetStatus(cfg *config.Config) (*Status, error) {
	pidPath := PidPath(cfg)
	pid, running := IsRunning(pidPath)

	st := &Status{
		Running:  running,
		PID:      pid,
		Interval: cfg.AgentCheckInterval,
	}

	// Try to get source stats from the database.
	if _, err := store.Open(cfg.DatabasePath); err != nil {
		return st, nil
	}

	sources, err := store.ListAgentSources()
	if err != nil {
		return st, nil
	}

	st.SourceCount = len(sources)
	var latestCheck *time.Time
	for _, s := range sources {
		if s.LastError != "" {
			st.ErrorCount++
		}
		if s.LastCheckAt != nil {
			if latestCheck == nil || s.LastCheckAt.After(*latestCheck) {
				latestCheck = s.LastCheckAt
			}
		}
	}

	if latestCheck != nil {
		st.LastCheckAt = latestCheck
		next := latestCheck.Add(time.Duration(cfg.AgentCheckInterval) * time.Hour)
		st.NextCheckAt = &next
	}

	return st, nil
}
