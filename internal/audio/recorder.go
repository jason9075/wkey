package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
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
// onLevel is called with normalized audio level (0.0-1.0) periodically.
func (r *Recorder) Start(filename string, onLevel func(float64)) error {
	// Check if pw-record is available
	_, err := exec.LookPath("pw-record")
	if err != nil {
		return fmt.Errorf("pw-record not found: %w", err)
	}

	// pw-record --format=s16 --rate=16000 --channels=1 -
	// We output to stdout (-) to capture data
	r.cmd = exec.Command("pw-record", "--format=s16", "--rate=16000", "--channels=1", "-")
	
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	r.cmd.Stderr = os.Stderr
	
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	if err := r.cmd.Start(); err != nil {
		outFile.Close()
		return fmt.Errorf("failed to start recording: %w", err)
	}

	// Process audio in background
	go func() {
		defer outFile.Close()
		
		// 16kHz, 16-bit mono = 32000 bytes/sec
		// Process chunks of ~50ms = 1600 bytes
		buf := make([]byte, 1600)
		
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				// Write to file
				if _, wErr := outFile.Write(buf[:n]); wErr != nil {
					fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", wErr)
				}

				// Calculate RMS
				if onLevel != nil {
					var sumSquares float64
					numSamples := n / 2
					for i := 0; i < numSamples; i++ {
						// Little endian 16-bit
						sample := int16(binary.LittleEndian.Uint16(buf[i*2 : i*2+2]))
						normalized := float64(sample) / 32768.0
						sumSquares += normalized * normalized
					}
					rms := math.Sqrt(sumSquares / float64(numSamples))
					
					// Boost level slightly for better visual
					displayLevel := rms * 5.0
					if displayLevel > 1.0 { displayLevel = 1.0 }
					
					onLevel(displayLevel)
				}
			}
			if err != nil {
				if err != io.EOF && err != os.ErrClosed {
					// fmt.Fprintf(os.Stderr, "Error reading stdout: %v\n", err)
				}
				break
			}
		}
	}()
	
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
