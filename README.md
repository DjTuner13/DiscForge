# DiscForge

DiscForge is a Go terminal UI for safely coordinating DVD restoration jobs.
It orchestrates external tools such as MakeMKV, VapourSynth, FFmpeg, and x265;
it does not replace them.

The default launch is the real profile-driven VapourSynth/FFmpeg TUI. The
untouched archive is always treated as read-only. Use `--demo` for a simulated
runner that makes no media-processing changes.

The implementation includes durable queue state, interrupted-job recovery,
per-job logs, FFmpeg progress parsing, YAML profile validation, and guarded
process execution. Live mode runs one job at a time and records failures in
the durable queue state.

## Run

```sh
go run ./cmd/discforge
```

The default archive root is `/mnt/archive`. Override it with:

```sh
DISCFORGE_ARCHIVE_ROOT=/path/to/archive go run ./cmd/discforge
```

Override the local state file during development with
`DISCFORGE_STATE_FILE=/tmp/discforge-state.json`.

Equivalent flags are available:

```sh
go run ./cmd/discforge --archive-root /mnt/archive --state-file /tmp/discforge-state.json

# Demo mode; no media-processing commands are connected.
go run ./cmd/discforge --demo --archive-root /mnt/archive
```

For a noninteractive, read-only archive check:

```sh
discforge --archive-root /mnt/archive --scan
```

For a full restoration outside the archive, use the safety-checked helper
script from the processing server:

```sh
./scripts/run-restoration.sh \
  "/mnt/archive/Viva La Bam/Season 01/Disc 01/A1_t05.mkv" \
  scripts/qtgmc-source.vpy \
  /mnt/work/A1_t05.restored.mkv
```

Use `j`/`k` to move, `h`/`l` to switch views, `Enter` to open, `Space` to
select, `r` to queue a restoration, `?` for help, and `q` to quit. By default,
`r` starts the real executor and writes process output to the per-job log
directory. Use `--demo` to simulate jobs.

On the processing server, set the external tool paths before launching live
mode:

```sh
export DISCFORGE_VSPIPE=/home/dj/src/viva-remaster/.venv/bin/vspipe
export DISCFORGE_FFMPEG=/usr/bin/ffmpeg
```

The repository also includes `scripts/launch-live.sh`, which sets these
defaults automatically:

```sh
./scripts/launch-live.sh
```

## Project plan

See [PLAN.md](PLAN.md) for the architecture, safety requirements, and phased
roadmap.
