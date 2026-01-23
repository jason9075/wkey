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

### Nix (Recommended)

If you are using Nix, you can run `wkey` directly or install it via Flakes.

**Run directly:**
```bash
nix run github:jason9075/wkey
```

**Install to profile:**
```bash
nix profile install github:jason9075/wkey
```

**Add to NixOS configuration:**
```nix
{
  inputs.wkey.url = "github:jason9075/wkey";
  # ...
  outputs = { self, nixpkgs, wkey }: {
    nixosConfigurations.my-pc = nixpkgs.lib.nixosSystem {
      modules = [
        ({ pkgs, ... }: {
          environment.systemPackages = [ wkey.packages.${pkgs.system}.default ];
        })
      ];
    };
  };
}
```

### Manual Installation

## Configuration

Wkey requires an OpenAI API key to function. You can provide this in two ways:

1.  **Environment Variable**: Set `OPENAI_API_KEY`.
2.  **Config File**: Create `~/.config/wkey/config.json`.

### Config File Example

Create the file `~/.config/wkey/config.json`:

```json
{
  "openai_api_key": "sk-...",
  "language": "zh",
  "visual": {
    "bar_count": 32,
    "bar_color_start": "#00FFFF",
    "bar_color_end": "#8A2BE2",
    "animation_speed": 1.0
  }
}
```

### Configuration Options

- **openai_api_key**: Your OpenAI API key.
- **language**: The language for transcription (e.g., `zh`, `zh-TW`, `en`). Defaults to `zh`.
- **visual**:
  - **bar_count**: Number of bars in the visualizer (default: 32).
  - **bar_color_start**: Start color gradient in hex (default: "#00FFFF").
  - **bar_color_end**: End color gradient in hex (default: "#8A2BE2").
  - **animation_speed**: Animation speed multiplier (default: 1.0).

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

## Troubleshooting

- **No Audio**: Ensure PipeWire is running and your default microphone is set correctly in `pavucontrol` or `wpctl`.
- **Stuck in Recording**: Run the command again to toggle it off. Wkey uses a PID file to manage state.
- **Wayland Protocol Errors**: Ensure you are running in a Wayland session.

## License

[MIT](LICENSE)
