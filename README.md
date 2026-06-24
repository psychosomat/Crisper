<p align="center">
  <img src="frontend/src/assets/logo.svg" alt="Crisper" width="120" />
</p>

<h1 align="center">Crisper</h1>

<p align="center">
  <strong>Crisper than Whisper</strong><br/>
  <em>Local video transcription to Markdown with speaker diarization.</em>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-a0ec06?style=for-the-badge" alt="License" /></a>
  <a href="https://github.com/psychosomat/Crisper/releases"><img src="https://img.shields.io/github/v/release/psychosomat/Crisper?style=for-the-badge&color=a0ec06" alt="Release" /></a>
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-333?style=for-the-badge&labelColor=333" alt="Platforms" />
</p>

---

## Features

| | Feature |
|---|---------|
| **Batch** | Drag-and-drop multiple video files at once |
| **Speakers** | Automatic speaker diarization |
| **Markdown** | Clean transcripts with optional timestamps |
| **Models** | Choose from tiny to large-v3 Whisper models |
| **Cross-platform** | Linux, macOS, Windows |

## Requirements

- [ffmpeg](https://ffmpeg.org)
- [whisper-cli](https://github.com/ggerganov/whisper.cpp) (in PATH or at `~/.local/bin/whisper-cli`)

## Quick Start

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
wails build
./build/bin/Crisper
```

On first launch, select a Whisper model — it downloads automatically.

## Usage

1. Drop video files (or click "Select Files")
2. Press **Start**
3. Transcripts appear as `<name>.md` next to each video

```markdown
# Transcript: interview.mp4

## Speaker 1 (00:01:23)
Hello and welcome to our podcast.

## Speaker 2 (00:01:27)
Thanks for having me.
```

## Settings

| Setting | Description |
|---------|-------------|
| Model | tiny / base / small / medium / large-v3 |
| Language | en, ru, auto-detect |
| Timestamps | toggle HH:MM:SS in output |
| Threads | CPU threads for inference |
| Output | where .md files are saved (default: next to video) |

## Development

```bash
wails dev       # hot reload
go vet ./...    # lint
```

## License

MIT
