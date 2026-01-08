package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"wkey/internal/audio"
	"wkey/internal/clipboard"
	"wkey/internal/config"
	"wkey/internal/stt"
	"wkey/internal/ui"
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

func getFileSize(path string) int64 {
	stat, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return stat.Size()
}

func main() {
	pidFile := getPidFilePath()

	// Parse flags
	keepTemp := flag.Bool("keep-temp", false, "Do not delete the temporary recording file on exit")
	mockResponse := flag.String("mock-response", "", "Force a specific mock response for STT")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Failed to load config: %v\n", err)
	}

	defer fmt.Println("[Main] Exiting main function")

	// Setup Signal Handling EARLY to prevent race condition
	stopChan := make(chan struct{})
	doneChan := make(chan struct{}) // New: to ensure main waits for logic
	sigChan := make(chan os.Signal, 1)
	// Listen for INT, TERM (for actual kills) and USR1 (for our toggle)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	go func() {
		first := true
		for sig := range sigChan {
			fmt.Printf("\n[Main] Received signal: %v\n", sig)
			if sig == syscall.SIGUSR1 || sig == syscall.SIGTERM || sig == syscall.SIGINT {
				if first {
					fmt.Printf("[Main] Signaling stopChan...\n")
					close(stopChan)
					first = false
				} else {
					fmt.Printf("[Main] Subsequent signal ignored to allow transcription to finish\n")
				}
			}
		}
	}()

	// 1. Check for existing instance (Toggle Logic)
	if content, err := os.ReadFile(pidFile); err == nil {
		pidStr := strings.TrimSpace(string(content))
		pid, err := strconv.Atoi(pidStr)
		if err == nil {
			// Process exists, send SIGUSR1 (less likely to be intercepted as "kill")
			proc, err := os.FindProcess(pid)
			if err == nil {
				fmt.Printf("[Main] Signaling existing process %d with SIGUSR1\n", pid)
				err := proc.Signal(syscall.SIGUSR1)
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
	fmt.Printf("[Main] Starting new instance (PID: %d)\n", pid)
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		fmt.Printf("Failed to write PID file: %v\n", err)
		return
	}
	defer os.Remove(pidFile)

	// Init UI
	u := ui.New(cfg)

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
	sttClient, err := stt.NewClient(cfg.OpenAIAPIKey, cfg.Language, *mockResponse, *verbose)
	// We check err later in the goroutine to allow UI to show error

	// Logic Goroutine
	go func() {
		if err != nil {
			fmt.Printf("[Logic] STT Client Init Error: %v\n", err)
			u.ShowError("Missing API Key")
			time.Sleep(3 * time.Second)
			u.Quit()
			return
		}

		// Start Recording
		fmt.Printf("[Logic] Starting recording...\n")
		u.ShowRecording()
		if err := recorder.Start(tmpFile, u.SetAudioLevel); err != nil {
			fmt.Printf("[Logic] Recorder Start Error: %v\n", err)
			u.ShowError("Rec Error: " + err.Error())
			time.Sleep(3 * time.Second)
			u.Quit()
			return
		}

		// Wait for Stop Signal or Timeout (60s)
		select {
		case <-stopChan:
			fmt.Printf("[Logic] Stop signal received via stopChan\n")
		case <-time.After(60 * time.Second):
			fmt.Printf("[Logic] Recording timeout (60s) reached\n")
			u.ShowError("Timeout (60s)")
			// Proceed to stop and transcribe
		}

		// Stop Recording
		fmt.Printf("[Logic] Stopping recorder...\n")
		recorder.Stop()
		fmt.Printf("[Logic] Recorder stopped. File size: %d bytes\n", getFileSize(tmpFile))
		u.ShowTranscribing()

		// Transcribe
		fmt.Printf("[Logic] Starting transcription...\n")
		text, err := sttClient.Transcribe(tmpFile)
		if err != nil {
			fmt.Printf("[Logic] Transcription Error: %v\n", err)
			u.ShowError(err.Error())
			time.Sleep(3 * time.Second)
			u.Quit()
			return
		}
		fmt.Printf("[Logic] Transcription finished. Result: %q\n", text)

		if text == "" {
			u.ShowError("No speech detected")
			time.Sleep(2 * time.Second)
			u.Quit()
			return
		}

		// Clipboard & Paste
		fmt.Printf("[Logic] Hiding UI to restore focus...\n")
		u.Hide()
		time.Sleep(500 * time.Millisecond) // Increased delay for focus restoration

		fmt.Printf("[Logic] Typing text directly to focus using wtype...\n")
		if err := clipboard.Type(text); err != nil {
			fmt.Printf("[Logic] Type Failed: %v\n", err)
			u.ShowError("Type Failed")
			time.Sleep(2 * time.Second)
		} else {
			fmt.Printf("[Logic] Type successful\n")
		}

		fmt.Printf("[Logic] Done. Quitting UI...\n")
		u.Quit()
		close(doneChan)
	}()

	u.Run()
	fmt.Println("[Main] u.Run() returned")
	<-doneChan
	fmt.Println("[Main] doneChan closed, exiting")
}
