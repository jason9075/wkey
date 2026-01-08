# Wkey - Agent Context & Guidelines

This document serves as the primary context and rulebook for AI agents working on the **Wkey** project.

## 1. Project Overview
**Wkey** is a lightweight voice input utility designed specifically for **Wayland** compositors (e.g., Hyprland, Sway).
- **Trigger**: Activated via a compositor-level hotkey (not a global listener).
- **Input**: Records audio from the default **PipeWire** source.
- **Processing**: Transcribes audio using **OpenAI Whisper** (non-streaming).
- **Output**: Copies text to the clipboard and attempts to paste it.
- **UX**: Minimalist UI (Fyne) to show state (Recording -> Transcribing -> Done/Error).

## 2. Core Engineering Guidelines (Non-Negotiable)
These rules are derived from `guidelines` and must be strictly followed.

### Scope & Constraints
- **Not an Input Method**: Do not implement IM protocols.
- **Security**: Do not bypass Wayland security. No global key interception.
- **Single Instance**: Use a PID file (`XDG_RUNTIME_DIR/wkey.pid`) to enforce single instance. Re-launching toggles the existing instance (stops recording).

### Audio (PipeWire)
- **Format**: Fixed to **Mono, 16kHz, PCM WAV**.
- **Control**: Simple Start/Stop. No VAD (Voice Activity Detection).

### Speech-to-Text (STT)
- **Provider**: OpenAI Whisper API.
- **Mode**: Non-streaming (send file after recording stops).
- **Handling**: Async, cancellable, with timeouts.

### UI (Fyne)
- **Role**: Informational only (State/Progress/Error).
- **Focus**: Must NOT steal keyboard focus.
- **Interaction**: No user interaction (buttons/text input) required for core flow.

### Error Handling
- **Fail Gracefully**: UI should briefly show errors before exiting.
- **Clipboard**: Writing to clipboard is mandatory. Pasting is best-effort.

## 3. Project Structure
```
wkey/
├── cmd/
│   └── wkey/           # Main entry point
├── internal/
│   ├── audio/          # PipeWire recording logic
│   ├── clipboard/      # Clipboard & Paste logic (wl-clipboard/xclip fallback)
│   ├── stt/            # OpenAI API client
│   └── ui/             # Fyne UI implementation
├── guidelines          # Original engineering guidelines
├── go.mod              # Dependencies
└── README.md           # User documentation
```

## 4. Development Workflow
- **Build**: `go build ./cmd/wkey`
- **Run**: `OPENAI_API_KEY=sk-... ./wkey`
- **Testing**:
    - Run the binary.
    - It should start recording (UI shows "Recording").
    - Send `SIGTERM` or run the binary again to stop recording.
    - Verify text is in clipboard.

## 5. Tech Stack
- **Language**: Go 1.25+
- **UI**: Fyne (v2)
- **Audio**: PipeWire (via `parec` or native bindings if implemented)
- **STT**: OpenAI API
