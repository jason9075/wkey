package audio

import (
	"fmt"
	"os"
	"os/exec"
)

// Recorder handles audio recording using external tools (pw-record)
type Recorder struct {
	cmd *exec.Cmd
}

// NewRecorder creates a new Recorder instance
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Start begins recording to the specified filename.
// It uses pw-record with 16kHz, mono, 16-bit PCM settings.
func (r *Recorder) Start(filename string) error {
	// Check if pw-record is available
	_, err := exec.LookPath("pw-record")
	if err != nil {
		return fmt.Errorf("pw-record not found: %w", err)
	}

	// pw-record --format=s16 --rate=16000 --channels=1 <filename>
	r.cmd = exec.Command("pw-record", "--format=s16", "--rate=16000", "--channels=1", filename)
	
	// Connect stdout/stderr to parent for debugging if needed, or ignore
	r.cmd.Stderr = os.Stderr
	
	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}
	
	return nil
}

// Stop stops the recording gracefully by sending SIGINT.
func (r *Recorder) Stop() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	// Send SIGINT to allow pw-record to finalize the file header
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		// If process is already dead, ignore
		return fmt.Errorf("failed to send interrupt signal: %w", err)
	}

	// Wait for the process to exit
	err := r.cmd.Wait()
	if err != nil {
		// It's common for command to return error on signal kill, check if it's just that
		// But usually Wait() returns exit code. pw-record might return 0 on SIGINT.
		// Let's return the error for now to be safe, but we might need to ignore "exit status X"
		return err
	}
	
	return nil
}
