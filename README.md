# DiscForge

DiscForge is a Go terminal UI for safely coordinating DVD restoration jobs.
It orchestrates external tools such as MakeMKV, VapourSynth, FFmpeg, and x265;
it does not replace them.

The default launch is a safe simulated-runner TUI. The untouched archive is
always treated as read-only. After the processing server has been validated,
`--live` enables the real profile-driven VapourSynth/FFmpeg executor.

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

# Explicit live mode; this starts real media processing only after you queue a job.
go run ./cmd/discforge --live --archive-root /mnt/archive \
  --profile profiles/dvd-ntsc-qtgmc.yaml \
  --script /path/to/qtgmc.vpy
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
select, `r` to queue a restoration, `?` for help, and `q` to quit. Without
`--live`, queued jobs are simulated. With `--live`, `r` starts the real
executor and writes process output to the per-job log directory.

## Project plan

See [PLAN.md](PLAN.md) for the architecture, safety requirements, and phased
roadmap.
