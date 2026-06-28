# osx-cd-tools

A macOS command-line utility for working with audio CDs.

## Requirements

### Homebrew dependencies

Install [Homebrew](https://brew.sh) if you haven't already, then:

```bash
brew install cdrdao
```

`cdrdao` handles reading and writing CDs with raw sector accuracy, which is
the highest-quality approach available for audio discs.

## Installation

```bash
git clone https://github.com/scarroll32/osx-cd-tools.git
cd osx-cd-tools
go build -o cd-tools .
```

Optionally move the binary onto your PATH:

```bash
mv cd-tools /usr/local/bin/
```

## Usage

### `copy` — duplicate an audio CD

```
cd-tools copy [flags]
```

The command walks you through a two-step process:

1. **Read** — insert the source CD and press Enter; the disc is imaged in raw
   mode at the chosen speed.
2. **Write** — eject the source, insert a blank CD-R, press Enter; the image
   is burned to the blank disc.

**Flags**

| Flag | Default | Description |
|------|---------|-------------|
| `-d`, `--device` | auto | CD drive device path (e.g. `/dev/disk2`). Auto-detected if omitted. |
| `-s`, `--speed` | `4` | Read/write speed multiplier. Lower values reduce read errors and improve burn quality. Values 1–8 are recommended for audio CDs. |
| `--keep-image` | false | Keep the disc image (`.toc` + `.bin`) after the copy finishes. |
| `--image-dir` | temp dir | Directory to store the disc image. Implies `--keep-image`. |

**Examples**

```bash
# Basic copy — auto-detect drive, 4x speed
cd-tools copy

# Specify drive and use 2x speed for a difficult disc
cd-tools copy --device /dev/disk2 --speed 2

# Keep the disc image for later use
cd-tools copy --keep-image --image-dir ~/disc-images/album
```

### Quality notes

- `--read-raw` is always used, which reads 2352-byte raw sectors and preserves
  all audio data exactly as it sits on the disc.
- Lower speeds (1x–4x) give the drive more time to re-read marginal sectors,
  which reduces audio dropouts on worn or scratched discs.
- `cdrdao` writes a `.toc` file (track layout / CD-Text) alongside the binary
  `.bin` data file, so the copy is structurally identical to the original.

### Finding your drive device path

If auto-detection fails, list available drives:

```bash
cdrdao scan-bus
diskutil list
```

The optical drive typically appears as `/dev/disk2` or `/dev/disk3`.

## Commands summary

```
cd-tools
└── copy    Copy an audio CD (read source, then write to blank)
```
