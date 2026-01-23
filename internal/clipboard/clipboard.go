package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CopyToClipboard writes the text to the Wayland clipboard using wl-copy.
// It runs in the background to avoid blocking the main process.
func CopyToClipboard(text string) error {
	cmd := exec.Command("wl-copy")
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(text)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Use Start instead of Run to avoid hanging.
	// wl-copy stays alive to serve the clipboard content.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wl-copy: %w, stderr: %s", err, stderr.String())
	}

	// We don't Wait() here because wl-copy needs to remain running.
	return nil
}

// Paste triggers the most common paste shortcuts for both GUI and Terminals.
// We use Ctrl+Shift+V as it's widely supported in modern Linux apps and terminals.
func Paste() error {
	if _, err := exec.LookPath("wtype"); err != nil {
		return fmt.Errorf("wtype not found: %w", err)
	}

	// Use Ctrl+Shift+V. This works in Kitty and most modern GUI apps (as paste-plain-text).
	// Avoiding sending both Ctrl+V and Ctrl+Shift+V to prevent double-pasting in browsers.
	cmd := exec.Command("wtype", "-M", "ctrl", "-M", "shift", "-k", "v", "-m", "shift", "-m", "ctrl")
	cmd.Env = os.Environ()
	return cmd.Run()
}
