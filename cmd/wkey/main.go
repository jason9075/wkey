package main

import (
"flag"
"fmt"
"os"
"os/signal"
"path/filepath"
"strconv"
"syscall"
"time"
"wkey/internal/audio"
"wkey/internal/clipboard"
"wkey/internal/stt"
"wkey/internal/ui"

"github.com/joho/godotenv"
)

const pidFileName = "voice-input.pid"

func getPidFilePath() string {
	// Use XDG_RUNTIME_DIR if available, else /tmp
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = "/tmp"
	}
	return filepath.Join(dir, pidFileName)
}

func main() {
	// Load .env file (ignore error if not found, as env vars might be set otherwise)
	_ = godotenv.Load()

	pidFile := getPidFilePath()

	// Parse flags
	keepTemp := flag.Bool("keep-temp", false, "Do not delete the temporary recording file on exit")
	mockResponse := flag.String("mock-response", "", "Force a specific mock response for STT")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// 1. Check for existing instance (Toggle Logic)
	if content, err := os.ReadFile(pidFile); err == nil {
		pid, err := strconv.Atoi(string(content))
		if err == nil {
			// Process exists, send SIGTERM
			proc, err := os.FindProcess(pid)
			if err == nil {
				// Send signal to stop recording
				err := proc.Signal(syscall.SIGTERM)
				if err == nil {
					// Successfully signaled, exit this instance
					return
				}
			}
		}
		// If we are here, PID file exists but process might be dead. Clean up.
		os.Remove(pidFile)
	}

	// 2. Start New Instance
	// Write PID
	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		fmt.Printf("Failed to write PID file: %v\n", err)
		return
	}
	defer os.Remove(pidFile)

	// Init UI
	u := ui.New()

	// Init Audio
	recorder := audio.NewRecorder()
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("voice-input-%d.wav", pid))
	if !*keepTemp {
		defer os.Remove(tmpFile)
	} else {
		fmt.Printf("Keeping temp file: %s\n", tmpFile)
	}

	if *verbose {
		fmt.Printf("Recording to temp file: %s\n", tmpFile)
	}

	// Init STT
	sttClient, err := stt.NewClient(*mockResponse, *verbose)
	// We check err later in the goroutine to allow UI to show error

	// Logic Goroutine
	go func() {
		if err != nil {
			u.ShowError("API Key Error")
			time.Sleep(3 * time.Second)
			u.Quit()
			return
		}

		// Start Recording
		u.ShowRecording()
		if err := recorder.Start(tmpFile, u.SetAudioLevel); err != nil {
			u.ShowError("Rec Error: " + err.Error())
			time.Sleep(3 * time.Second)
			u.Quit()
			return
		}

		// Wait for Stop Signal or Timeout (60s)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-sigChan:
			// User stopped manually
		case <-time.After(60 * time.Second):
			// Timeout reached
			u.ShowError("Timeout (60s)")
			// Proceed to stop and transcribe
		}

		// Stop Recording
		recorder.Stop()
		u.ShowTranscribing()

		// Transcribe
		text, err := sttClient.Transcribe(tmpFile)
		if err != nil {
			u.ShowError("STT Error: " + err.Error())
			time.Sleep(3 * time.Second)
			u.Quit()
			return
		}

		if text == "" {
			u.ShowError("No speech detected")
			time.Sleep(2 * time.Second)
			u.Quit()
			return
		}

		// Clipboard & Paste
		if err := clipboard.CopyToClipboard(text); err != nil {
			u.ShowError("Copy Failed")
		} else {
			if err := clipboard.Paste(); err != nil {
				u.ShowError("Copied (Paste Failed)")
			} else {
				u.ShowDone()
			}
		}

		time.Sleep(1 * time.Second)
		u.Quit()
	}()

	u.Run()
}
