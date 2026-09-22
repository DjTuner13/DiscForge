# DiscForge architecture

DiscForge is an orchestrator around existing media tools. It does not rip
discs, restore video, or provide metadata scraping itself.

## Boundaries

- `internal/archive` discovers archive masters. It only reads paths.
- `internal/probe` invokes `ffprobe` and decodes machine-readable media data.
- `internal/jobs` owns job state, queue ordering, and progress values.
- `internal/state` persists queue snapshots atomically and marks active jobs as
  `interrupted` after a restart.
- `internal/profiles` loads editable YAML settings and enforces stream-copy
  defaults for audio, subtitles, and chapters.
- `internal/pipeline` owns process execution, cancellation, mux command
  construction, output safety, and validation.
- `internal/hardware` reads Linux hwmon values without repeatedly spawning
  `sensors`.
- `internal/tui` is the operator interface and should remain independent of
  media implementation details.

The archive root is a logical read-only boundary. Every future output path must
pass through `pipeline.GuardedRunner.ValidateOutput`, and existing outputs must
never be overwritten automatically.

## Execution flow

1. Scan and display archive paths.
2. Probe a selected source with `ffprobe`.
3. Create a queued job and persist the snapshot.
4. Run one restoration pipeline at a time using a cancellable context.
5. Parse FFmpeg `-progress` output and persist updates.
6. Mux restored video with original streams and metadata.
7. Probe and validate the finished output.
8. Mark the job complete only after validation succeeds.

Real command execution is intentionally kept behind these boundaries so local
tests can use synthetic MKVs and fake producer/consumer commands.

