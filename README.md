# DiscForge

DiscForge is a Go terminal UI for safely coordinating DVD restoration jobs.
It orchestrates external tools such as MakeMKV, VapourSynth, FFmpeg, and x265;
it does not replace them.

The first milestone is a safe, fake-runner TUI that establishes the navigation,
queue, and job-state architecture before any media-processing commands are
connected. The untouched archive is always treated as read-only.

The current implementation also includes durable queue state, interrupted-job
recovery, FFmpeg progress parsing, YAML profile validation, and guarded process
execution. Real VapourSynth/FFmpeg execution remains intentionally gated until
the pipeline is tested against the processing server.

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

Use `j`/`k` to move, `h`/`l` to switch views, `Enter` to open, `Space` to
select, `r` to queue a fake restoration, `?` for help, and `q` to quit.

## Project plan

See [PLAN.md](PLAN.md) for the architecture, safety requirements, and phased
roadmap.
