package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Recorder handles audio recording using external tools (pw-record)
type Recorder struct {
	cmd        *exec.Cmd
	wg         sync.WaitGroup
	outFile    *os.File
	totalBytes uint32
}

// NewRecorder creates a new Recorder instance
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Start begins recording to the specified filename.
// It uses pw-record with 16kHz, mono, 16-bit PCM settings.
// onLevel is called with normalized audio level (0.0-1.0) periodically.
func (r *Recorder) Start(filename string, onLevel func(float64)) error {
	fmt.Printf("[Recorder] Starting pw-record to %s\n", filename)
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

	r.outFile, err = os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	// Write initial empty header (44 bytes)
	header := make([]byte, 44)
	if _, err := r.outFile.Write(header); err != nil {
		r.outFile.Close()
		return fmt.Errorf("failed to write initial header: %w", err)
	}
	r.totalBytes = 0

	if err := r.cmd.Start(); err != nil {
		r.outFile.Close()
		return fmt.Errorf("failed to start recording: %w", err)
	}

	r.wg.Add(1)
	// Process audio in background
	go func() {
		defer r.wg.Done()
		fmt.Printf("[Recorder] Data processing goroutine started\n")

		// 16kHz, 16-bit mono = 32000 bytes/sec
		// Process chunks of ~50ms = 1600 bytes
		buf := make([]byte, 1600)
		
		lastPrint := time.Now()

		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				// Write to file
				if _, wErr := r.outFile.Write(buf[:n]); wErr != nil {
					fmt.Fprintf(os.Stderr, "[Recorder] Error writing to file: %v\n", wErr)
				}
				r.totalBytes += uint32(n)
				
				if time.Since(lastPrint) > 2*time.Second {
					fmt.Printf("[Recorder] Total bytes written: %d\n", r.totalBytes)
					lastPrint = time.Now()
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
					if displayLevel > 1.0 {
						displayLevel = 1.0
					}

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
		fmt.Printf("[Recorder] Stop called but no process running\n")
		return nil
	}

	fmt.Printf("[Recorder] Sending SIGINT to pw-record (PID: %d)...\n", r.cmd.Process.Pid)
	// Send SIGINT to allow pw-record to finalize the file header
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		// If process is already dead, ignore
		return fmt.Errorf("failed to send interrupt signal: %w", err)
	}

	// Wait for the process to exit
	fmt.Printf("[Recorder] Waiting for pw-record to exit...\n")
	err := r.cmd.Wait()

	// Wait for file writing to finish
	fmt.Printf("[Recorder] Waiting for data processing goroutine to finish...\n")
	r.wg.Wait()

	// Finalize WAV header
	if r.outFile != nil {
		fmt.Printf("[Recorder] Finalizing WAV header (total bytes: %d)...\n", r.totalBytes)
		if _, sErr := r.outFile.Seek(0, 0); sErr == nil {
			r.writeWavHeader(r.outFile, r.totalBytes)
		}
		r.outFile.Close()
		r.outFile = nil
	}
	fmt.Printf("[Recorder] Stop finished\n")

	if err != nil {
		// It's common for command to return error on signal kill, check if it's just that
		// But usually Wait() returns exit code. pw-record might return 0 on SIGINT.
		// Let's return the error for now to be safe, but we might need to ignore "exit status X"
		return err
	}

	return nil
}

func (r *Recorder) writeWavHeader(w io.Writer, dataSize uint32) {
	// Standard 44-byte WAV header
	binary.Write(w, binary.BigEndian, []byte("RIFF"))
	binary.Write(w, binary.LittleEndian, dataSize+36)
	binary.Write(w, binary.BigEndian, []byte("WAVE"))
	binary.Write(w, binary.BigEndian, []byte("fmt "))
	binary.Write(w, binary.LittleEndian, uint32(16))
	binary.Write(w, binary.LittleEndian, uint16(1))     // PCM
	binary.Write(w, binary.LittleEndian, uint16(1))     // Mono
	binary.Write(w, binary.LittleEndian, uint32(16000)) // Sample Rate
	binary.Write(w, binary.LittleEndian, uint32(32000)) // Byte Rate
	binary.Write(w, binary.LittleEndian, uint16(2))     // Block Align
	binary.Write(w, binary.LittleEndian, uint16(16))    // Bits Per Sample
	binary.Write(w, binary.BigEndian, []byte("data"))
	binary.Write(w, binary.LittleEndian, dataSize)
}
