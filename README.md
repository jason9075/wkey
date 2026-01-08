# Wkey

**Wkey** is a minimalist, Wayland-native voice input utility. It allows you to dictate text anywhere in your Wayland environment (Hyprland, Sway, etc.) using a global hotkey managed by your compositor.

## Features

- **Wayland Native**: Designed for modern Linux compositors.
- **Privacy-Focused**: Only records when triggered.
- **High Quality**: Uses OpenAI Whisper for accurate transcription.
- **Simple Workflow**: Press hotkey -> Speak -> Text appears in clipboard (and auto-pastes).
- **Visual Feedback**: Unobtrusive UI to show recording status.

## Requirements

- **Linux** with a Wayland Compositor (Hyprland, Sway, etc.)
- **PipeWire** (for audio recording)
- **Go** 1.25+ (for building)
- **OpenAI API Key**

## Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/yourusername/wkey.git
    cd wkey
    ```

2.  Build the binary:
    ```bash
    go build -o wkey ./cmd/wkey
    ```

3.  Move the binary to your PATH (optional):
    ```bash
    sudo mv wkey /usr/local/bin/
    ```

## Configuration

Wkey requires an OpenAI API key to function. You must set the `OPENAI_API_KEY` environment variable.

You can set this in your shell configuration or pass it when running the command.

### Flags

- `--keep-temp`: Do not delete the temporary recording file on exit. Useful for debugging audio issues.
- `--mock-response "Your text here"`: Force a specific mock response for STT. Useful for testing without hitting the OpenAI API.

## Usage

Wkey is designed to be triggered by a hotkey. It uses a "toggle" mechanism:
- **First Press**: Starts recording.
- **Second Press**: Stops recording, transcribes, and copies to clipboard.

### Hyprland Configuration

Add the following to your `hyprland.conf`:

```ini
# Bind Super+V to toggle voice input
bind = SUPER, V, exec, OPENAI_API_KEY=sk-your-key-here /path/to/wkey
```

*Note: It is recommended to use a script or a secrets manager to handle your API key securely instead of hardcoding it in the config.*

### Sway Configuration

Add the following to your `config`:

```ini
bindsym Mod4+v exec OPENAI_API_KEY=sk-your-key-here /path/to/wkey
```

## Troubleshooting

- **No Audio**: Ensure PipeWire is running and your default microphone is set correctly in `pavucontrol` or `wpctl`.
- **Stuck in Recording**: Run the command again to toggle it off. Wkey uses a PID file to manage state.
- **Wayland Protocol Errors**: Ensure you are running in a Wayland session.

## License

[MIT](LICENSE)
