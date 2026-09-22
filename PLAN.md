# DiscForge — Go/Charm TUI for DVD Restoration

## Project Goal

Build a terminal UI called **DiscForge** that manages long-running DVD restoration jobs.

DiscForge should orchestrate the existing working toolchain:

- MakeMKV for archival DVD ripping
- VapourSynth for restoration/deinterlacing
- QTGMC/QTempGaussMC for 480i → 480p59.94
- FFmpeg/libx265 for HEVC encoding
- Sonarr for final episode identification/naming/import
- Jellyfin for playback and metadata

DiscForge should **not replace MakeMKV, VapourSynth, FFmpeg, Sonarr, or Jellyfin**.

Its primary job is to make the restoration stage easy to operate:

```text
Untouched MKV archive
        ↓
     DiscForge
        ↓
VapourSynth/QTGMC
        ↓
      x265
        ↓
mux original audio/subtitles/chapters
        ↓
finished staging file
        ↓
Sonarr Manual Import
        ↓
Jellyfin
```

The untouched MakeMKV archive must never be modified.

---

## Repository / Git Expectations

This project will be maintained as a **Git repository**.

Recommended initial setup:

```bash
mkdir discforge
cd discforge
git init
go mod init github.com/<owner>/discforge
```

Use normal Go project conventions and keep commits small and reviewable.

Recommended branch model:

```text
main
feature/<short-description>
fix/<short-description>
```

Use conventional-style commit messages where practical:

```text
feat: add archive browser
feat: add ffmpeg progress parser
fix: preserve queue state after interruption
refactor: split pipeline runner from tui model
docs: document restoration profiles
```

Do not commit:

- generated binaries
- runtime state
- logs
- temporary encodes
- benchmark outputs unless intentionally stored as fixtures
- secrets/API keys

Suggested `.gitignore`:

```gitignore
discforge
bin/
dist/
tmp/
*.log
*.partial
.env
.env.*
!.env.example
.coverage/
coverage.out
```

The repository should contain:

- source code
- example configuration
- restoration profile examples
- test fixtures where appropriate
- README
- architecture notes
- migration notes if persistent state format changes

---

# Current Infrastructure

## Hypervisor

Proxmox host:

```text
hostname: homelab
```

Current CPU:

```text
Intel Xeon W-2123
4 cores / 8 threads
```

Planned upgrade:

```text
Intel Xeon W-2245
8 cores / 16 threads
```

The program must not hard-code CPU assumptions.

---

## Ubuntu Processing VM

Main processing VM:

```text
hostname: ubuntu-server
IP: 192.168.50.158
user: dj
```

This VM handles:

- MakeMKV
- VapourSynth
- FFmpeg
- BluForge
- Sonarr
- DiscForge
- DVD drive access
- restoration workloads

---

# Storage Layout

Proxmox storage is exposed to Ubuntu through VirtioFS.

## Archive

Ubuntu:

```text
/mnt/archive
```

Backing Proxmox path:

```text
/tank/archive/dvd
```

Example untouched masters:

```text
/mnt/archive/Viva La Bam/Season 01/Disc 01/
```

Example files:

```text
A1_t01.mkv
A1_t02.mkv
A1_t05.mkv
...
```

These files are permanent archival masters.

**DiscForge must NEVER alter, rename, overwrite, move, or delete them.**

Allowed operations:

```text
read
probe
decode
```

---

## Restoration Work Area

```text
/mnt/work/video/viva-la-bam
```

Suggested finished staging area:

```text
/mnt/work/video/viva-la-bam/finished
```

Temporary intermediate files may live under:

```text
/mnt/work/video/viva-la-bam/tmp
```

---

## Jellyfin Remaster Library

```text
/mnt/media/Viva La Bam (2003) - Remaster
```

Desired structure:

```text
Viva La Bam (2003) - Remaster/
├── Season 00/
├── Season 01/
├── Season 02/
...
```

Desired filename format:

```text
Viva La Bam - S01E05 - Paint Phil Blue & Other Stories.mkv
```

Sonarr currently handles final episode naming.

---

# Optical Drive

Physical DVD drive passed through to Ubuntu:

```text
/dev/sr0
/dev/sg1
```

Drive:

```text
PLDS DVD-RW DH16AFSH
```

DiscForge v0.1 does **not** need to perform ripping.

BluForge/MakeMKV currently handle that.

DiscForge starts from already-ripped MKV files.

---

# Existing Restoration Toolchain

Current restoration project:

```text
~/src/viva-remaster
```

Python environment is managed by:

```text
uv
```

Python:

```text
3.12
```

Relevant packages include:

```text
vapoursynth
vsjetpack[deinterlace]
vapoursynth-lsmas
vapoursynth-ffms2
vapoursynth-znedi3
```

The source loader should use **L-SMASH Works** rather than BestSource.

BestSource failed reproducibly on one valid DVD source frame.

Working source loader:

```python
src = core.lsmas.LWLibavSource(
    source="/path/to/source.mkv"
)
```

---

# Proven Deinterlacing Pipeline

Viva La Bam DVD analysis showed:

```text
codec: MPEG-2
resolution: 720x480
sample aspect ratio: 8:9
display aspect ratio: 4:3
field order: top-field-first
frame rate: 30000/1001
```

`idet` showed the footage is predominantly true interlaced TFF material.

Therefore:

```text
480i29.97 TFF
    ↓
QTGMC-style bob deinterlace
    ↓
480p59.94
```

No IVTC should be applied to this source.

Current implementation:

```python
import vapoursynth as vs
from vsdeinterlace import QTempGaussMC

core = vs.core
core.max_cache_size = 2048

src = core.lsmas.LWLibavSource(
    source="/path/to/input.mkv"
)

out = QTempGaussMC().deinterlace(
    src,
    tff=True
)

out.set_output()
```

DiscForge should treat the VapourSynth script as configurable rather than embedding all restoration logic directly in Go.

---

# Current Video Decision

A test compared:

```text
480p59 QTGMC
```

against:

```text
1440x1080p59 QTGMC + Spline36 upscale
```

on a 75-inch 4K LG television through Jellyfin Direct Play.

The visual difference was negligible.

Current direction:

```text
720x480
59.94 progressive
10-bit HEVC
correct 4:3 display aspect ratio
```

Do **not** assume 1080p is the final output format.

Resolution, CRF, preset, etc. must be profile-driven.

---

# Existing Benchmark

1080p test:

```text
1440x1080
59.94 fps
HEVC
10-bit
x265 preset slow
CRF 14
```

Current Xeon W-2123 performance:

```text
7192 frames
2191.83 seconds
3.28 fps
```

QTGMC-only 480p processing previously measured approximately:

```text
32.23 fps
```

Current CPU temperature under restoration workload:

```text
~47°C package
```

A W-2245 benchmark will be performed later.

DiscForge should include a benchmark mode so these comparisons become first-class functionality.

---

# Working Name

Use:

```text
discforge
```

Avoid naming it specifically for Viva La Bam.

The program should eventually support:

- Viva La Bam
- Home Improvement
- other NTSC DVDs
- potentially other restoration profiles later

---

# Technology Stack

Use Go.

Use the Charm ecosystem:

```text
Bubble Tea
Lip Gloss
Bubbles
```

Recommended packages:

```text
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss
github.com/charmbracelet/bubbles
```

The UI should feel like a modern native terminal application rather than formatted shell output.

---

# Vim-Style Navigation

Vim-style navigation is a **first-class requirement**, not an afterthought.

The entire TUI should be operable without arrow keys.

Global movement:

```text
j       move down
k       move up
h       move left / parent / previous pane
l       move right / open / next pane
g g     jump to top
G       jump to bottom
Ctrl-d  half-page down
Ctrl-u  half-page up
Ctrl-f  page down
Ctrl-b  page up
```

Selection/action conventions:

```text
Enter   open / confirm
Space   toggle selection
Esc     back / close modal
/       filter/search
n       next search result
N       previous search result
?       help
q       back or quit depending context
```

Tab switching should support both numbers and Vim-friendly keys.

Suggested:

```text
1 / H   Library
2 / J   Queue
3 / K   Current Job
4 / L   Logs
```

Or another collision-free mapping chosen during implementation.

Important UX rule:

- `j/k` must always mean vertical navigation when a list/table is focused.
- `h/l` should move between logical panes or navigate parent/child views.
- arrow keys may also work, but Vim bindings are mandatory.
- key hints should show the Vim keys, not hide them behind arrow-key-only documentation.

For text fields, ensure navigation keys are not intercepted while the user is actively typing.

---

# High-Level Architecture

Suggested package layout:

```text
discforge/
├── cmd/
│   └── discforge/
│       └── main.go
│
├── internal/
│   ├── tui/
│   │   ├── model.go
│   │   ├── library.go
│   │   ├── queue.go
│   │   ├── job.go
│   │   ├── logs.go
│   │   └── keymap.go
│   │
│   ├── archive/
│   │   └── scanner.go
│   │
│   ├── jobs/
│   │   ├── queue.go
│   │   ├── runner.go
│   │   ├── state.go
│   │   └── progress.go
│   │
│   ├── pipeline/
│   │   ├── vapoursynth.go
│   │   ├── ffmpeg.go
│   │   └── mux.go
│   │
│   ├── probe/
│   │   └── ffprobe.go
│   │
│   ├── hardware/
│   │   └── temperature.go
│   │
│   ├── profiles/
│   │   └── profiles.go
│   │
│   └── state/
│       └── store.go
│
├── profiles/
│   └── dvd-ntsc-qtgmc.yaml
│
├── .gitignore
├── go.mod
├── README.md
└── PLAN.md
```

Exact structure can change if there is a better Go design.

---

# Main TUI Screens

## 1. Library

Browse existing archive masters.

Example:

```text
╭──────────────────────── DiscForge ────────────────────────╮
│ Library                                                   │
│                                                           │
│ Viva La Bam                                               │
│ ├── Season 01                                             │
│ │   ├── Disc 01                       8 titles            │
│ │   └── Disc 02                      15 titles            │
│ │                                                         │
│ Home Improvement                                          │
│ └── ...                                                   │
│                                                           │
│ j/k move   h/l nav   Enter open   / filter   q quit       │
╰───────────────────────────────────────────────────────────╯
```

Scanner should discover MKVs recursively beneath configured archive roots.

---

## 2. Disc / Title Selection

Example:

```text
╭──────── Viva La Bam / Season 01 / Disc 01 ───────────────╮
│                                                           │
│ [x] A1_t01.mkv    19:04    720x480 MPEG-2                │
│ [x] A1_t02.mkv    18:35    720x480 MPEG-2                │
│ [ ] A1_t03.mkv    19:08    720x480 MPEG-2                │
│ [x] A1_t04.mkv    19:08    720x480 MPEG-2                │
│                                                           │
│ Profile: DVD NTSC QTGMC / HEVC                            │
│                                                           │
│ j/k move   Space select   a all   r restore   b bench     │
╰───────────────────────────────────────────────────────────╯
```

Use `ffprobe` to populate:

- duration
- codec
- width
- height
- frame rate
- SAR
- DAR
- field order
- audio tracks
- subtitle tracks

---

## 3. Queue

Example:

```text
╭──────────────────── Restoration Queue ────────────────────╮
│                                                           │
│ ✓ A1_t01.mkv     19:04     Complete                       │
│ ✓ A1_t02.mkv     18:35     Complete                       │
│ ▶ A1_t03.mkv     19:08      47%       ETA 1h 32m         │
│ ○ A1_t04.mkv     19:08     Queued                         │
│ ○ A1_t05.mkv     18:16     Queued                         │
│                                                           │
│ Overall                                                   │
│ ███████████░░░░░░░░░░░░░░░░░ 38%                        │
│                                                           │
│ 12.8 fps   CPU 67°C   elapsed 03:18:42   ETA 08:31:09    │
│                                                           │
│ j/k move   Enter detail   s skip   l logs   q back        │
╰───────────────────────────────────────────────────────────╯
```

Version 1 should process only **one encode at a time**.

Do not run multiple x265/QTGMC jobs concurrently by default.

---

## 4. Job Detail

Example:

```text
A1_t03.mkv

Source:
720x480 MPEG-2
29.970i TFF
DAR 4:3

Output:
720x480 HEVC Main10
59.940p
CRF 16
preset medium

Video:
████████████████░░░░░░░░ 67%

frame       48291 / 71920
fps         13.1
speed       0.219x
elapsed     01:02:17
ETA         00:30:41

CPU package 68°C
output size 1.8 GB
```

---

## 5. Logs

Show raw process output in a scrollable viewport.

Support logs for:

```text
vspipe
ffmpeg
ffprobe
mux
DiscForge itself
```

Log navigation must support:

```text
j/k
gg
G
Ctrl-d
Ctrl-u
/
n
N
```

---

# Job State Model

Each job should have a persistent state:

```text
queued
preparing
running
muxing
validating
completed
failed
cancelled
interrupted
```

Potential Go structure:

```go
type Job struct {
    ID          string
    InputPath   string
    OutputPath  string
    Profile     string
    Status      JobStatus

    TotalFrames int64
    Frame       int64
    FPS         float64

    StartedAt   time.Time
    FinishedAt  time.Time

    Error       string
}
```

---

# Persistent State

Long encodes may run for many hours.

The queue must survive:

- SSH disconnect
- TUI restart
- accidental UI exit
- application crash

Suggested location:

```text
~/.local/state/discforge/
```

Example:

```text
~/.local/state/discforge/
├── state.json
├── queue.json
└── logs/
    ├── <job-id>.log
    └── ...
```

SQLite is also acceptable and may be preferable if it simplifies persistence.

Do NOT mark a job complete solely because the process exited successfully.

Validate its output first.

A job that was `running` when DiscForge crashed should become:

```text
interrupted
```

rather than silently restarting.

---

# FFmpeg Progress Parsing

Do not scrape normal FFmpeg terminal formatting.

Use FFmpeg's machine-readable progress mode:

```text
-progress pipe:2
-nostats
```

Parse keys such as:

```text
frame=
fps=
out_time=
speed=
progress=
```

Use total output frame count to calculate:

```text
percent complete
job ETA
disc ETA
queue ETA
```

Use moving averages rather than instantaneous FPS for stable ETA.

---

# VapourSynth Execution

Current working invocation style:

```bash
uv run vspipe -c y4m script.vpy - |
ffmpeg ...
```

DiscForge should invoke external commands rather than embedding VapourSynth.

Prefer explicit `exec.CommandContext()` process management.

Cancellation should cleanly terminate the entire pipeline.

Be careful with pipe children so killing DiscForge does not leave orphaned:

```text
vspipe
ffmpeg
x265
```

processes.

---

# Restoration Profiles

Encode settings must be configurable.

Do NOT hard-code the final production settings.

Suggested profile:

```yaml
name: dvd-ntsc-qtgmc-hevc

video:
  width: 720
  height: 480
  fps: 60000/1001
  pixel_format: yuv420p10le

source:
  field_order: tff

vapoursynth:
  script: qtgmc.vpy

encoder:
  codec: libx265
  preset: medium
  crf: 16

audio:
  mode: copy

subtitles:
  mode: copy

chapters:
  mode: copy
```

The final CRF/preset may change after W-2245 benchmarking.

Profiles should be editable without recompiling DiscForge.

---

# Benchmark Mode

Benchmarking should be a first-class feature.

Example key:

```text
b = benchmark
```

A benchmark should:

1. Use a configurable short frame/time range.
2. Run the selected restoration profile.
3. Record:
   - CPU model
   - wall-clock time
   - FPS
   - realtime multiplier
   - average/maximum CPU temperature if available
   - output bitrate
4. Save result.

Example output:

```text
Benchmark Complete

CPU:
Intel Xeon W-2245

Profile:
dvd-ntsc-qtgmc-hevc

Frames:
7192

Time:
543.2 sec

Performance:
13.24 fps
0.221x realtime

Estimated:
19 minute episode: 1h 26m
Disc 1: 11h 34m
```

This will be used to compare:

```text
Xeon W-2123
vs
Xeon W-2245
```

and potentially other machines later.

---

# CPU Temperature Monitoring

`lm-sensors` is installed.

Linux currently exposes Intel `coretemp`.

Observed output:

```text
Package id 0
Core 0
Core 1
...
```

Prefer reading Linux hwmon directly rather than repeatedly spawning:

```bash
sensors
```

Look beneath:

```text
/sys/class/hwmon/
```

Find the sensor exposing:

```text
Package id 0
```

Show CPU package temperature in the TUI.

Do not perform automatic thermal throttling in version 1.

Display only.

---

# Full MKV Output Requirements

The final restored MKV must contain:

```text
restored video
original DVD audio
original subtitles
original chapters
```

Video is the stream being transformed.

Audio/subtitles should normally be stream-copied.

Conceptually:

```text
SOURCE MKV
│
├── video ──→ VapourSynth ──→ x265 ──┐
│                                    │
├── audio ───────────────────────────┤
├── subtitles ───────────────────────┤
├── chapters ────────────────────────┤
└── metadata ────────────────────────┤
                                     ↓
                                FINAL MKV
```

Potential mux strategy:

1. Encode restored video to a temporary file.
2. Combine it with source streams using FFmpeg.

Conceptual command:

```bash
ffmpeg \
  -i restored-video.mkv \
  -i source.mkv \
  -map 0:v:0 \
  -map 1:a? \
  -map 1:s? \
  -map_metadata 1 \
  -map_chapters 1 \
  -c copy \
  final.mkv
```

Carefully verify stream disposition/language metadata.

Do not overwrite the source.

---

# Output Validation

Before marking a job complete, run `ffprobe`.

Validate at minimum:

```text
file exists
file size > 0
expected video codec exists
expected width/height
expected progressive frame rate
audio stream count preserved
subtitle stream count preserved
duration approximately matches source
chapters present when source had chapters
```

If validation fails:

```text
Status = failed
```

Do not delete intermediates automatically after a validation failure.

---

# Sonarr Integration

Sonarr is currently running in Docker:

```text
http://192.168.50.158:8989
```

It currently has access to:

```text
/work
/tv
```

Sonarr naming configuration:

```text
Standard Episode Format:
{Series Title} - S{season:00}E{episode:00} - {Episode Title}
```

Season Folder Format:

```text
Season {season:00}
```

Desired result:

```text
Season 01/
Viva La Bam - S01E05 - Paint Phil Blue & Other Stories.mkv
```

Version 1 of DiscForge should NOT automatically import into Sonarr.

Instead mark successfully completed outputs:

```text
Ready for Sonarr
```

Later version:

- connect to Sonarr API
- retrieve series
- retrieve seasons/episodes
- allow episode selection inside DiscForge
- optionally trigger/import/rename

Possible future UI:

```text
A1_t05.mkv

Series:
Viva La Bam

Season:
01

Episode:
05 — Paint Phil Blue & Other Stories

[Confirm Mapping]
```

---

# Jellyfin

Jellyfin is responsible for:

- posters
- descriptions
- actors
- episode metadata
- playback

DiscForge should not try to become a Jellyfin metadata scraper.

Correct filenames are sufficient.

---

# BluForge

BluForge is installed and operational.

It is useful as a visual MakeMKV frontend.

TheDiscDB currently does NOT contain Viva La Bam.

Therefore BluForge should NOT be a dependency of DiscForge.

DiscForge should operate entirely on existing MKV archive files.

---

# Safety Requirements

These are non-negotiable.

## Archive is read-only logically

Never modify files underneath:

```text
/mnt/archive
```

Operations allowed:

```text
read
probe
decode
```

Operations NOT allowed:

```text
delete
rename
overwrite
move
modify metadata in-place
```

---

## No overwrite by default

If output already exists:

```text
STOP
```

Prompt user or mark conflict.

Do not overwrite automatically.

---

## Graceful cancellation

If user cancels a job:

- stop FFmpeg
- stop vspipe
- update persistent job state
- preserve logs
- preserve partial output with identifiable `.partial` extension or delete only after explicit policy

---

## Original streams

Audio/subtitles/chapters should be copied rather than re-encoded unless a future profile explicitly requests otherwise.

---

# MVP Scope

Version 0.1 should do ONLY:

1. Scan archive directories.
2. Display MKV files in Charm TUI.
3. Support complete Vim navigation.
4. Select files.
5. Add them to a queue.
6. Run one job at a time.
7. Execute configurable VapourSynth → FFmpeg pipeline.
8. Show:
   - frame
   - percentage
   - FPS
   - elapsed time
   - ETA
9. Show CPU package temperature.
10. Save logs.
11. Persist queue state.
12. Produce finished output under `/mnt/work`.
13. Validate the output.

Do NOT implement Sonarr API integration in the MVP.

Do NOT implement DVD ripping in the MVP.

Do NOT implement metadata scraping in the MVP.

---

# Suggested Development Phases

## Phase 1 — Git + TUI Skeleton

Build:

- initialize Git repo
- initialize Go module
- add `.gitignore`
- Bubble Tea app
- Library view
- Queue view
- Job detail view
- Logs view
- Vim keymap
- keyboard navigation
- help overlay

Use fake/mock jobs initially.

Acceptance:

```text
discforge
```

opens a polished functional TUI and can be fully navigated with:

```text
h j k l
gg
G
Ctrl-d
Ctrl-u
/
n
N
```

---

## Phase 2 — Archive Scanner

Implement recursive MKV discovery.

Read configured root:

```text
/mnt/archive
```

Use `ffprobe` for media information.

Cache probe results if useful.

---

## Phase 3 — Job Runner

Implement external command pipeline.

Start with a fake/test FFmpeg pipeline if necessary.

Add:

- process management
- stdout/stderr handling
- cancellation
- logs

---

## Phase 4 — Real Progress

Use:

```text
ffmpeg -progress
```

Parse progress messages.

Calculate:

- current percentage
- FPS
- elapsed
- ETA

---

## Phase 5 — Hardware Metrics

Read:

```text
/sys/class/hwmon
```

Find CPU package sensor.

Display temperature.

---

## Phase 6 — Persistent Queue

Persist:

```text
queued
running
completed
failed
cancelled
interrupted
```

Recover gracefully after restart.

---

## Phase 7 — Actual Restoration Profile

Integrate current working VapourSynth script.

Use:

```text
QTempGaussMC
TFF
L-SMASH Works source
```

Keep settings profile-driven.

---

## Phase 8 — MKV Muxing

Preserve:

- audio
- subtitles
- chapters
- useful source metadata

Validate resulting MKV.

---

## Phase 9 — Benchmark Mode

Add repeatable benchmark runs.

Store history so CPU/profile comparisons are possible.

---

## Phase 10 — Sonarr Integration

Only after the restoration engine is stable.

Use Sonarr API to expose:

```text
series
season
episode
episode title
```

Allow mapping finished DVD titles to episodes.

---

# Initial Agent Tasks

Start with these tasks in order:

1. Create the Git repo and Go module.

```bash
mkdir discforge
cd discforge
git init
go mod init github.com/<owner>/discforge
```

2. Add Charm dependencies.

3. Add a clean `.gitignore`.

4. Create a `README.md` and keep this plan as `PLAN.md`.

5. Build a Bubble Tea application with:
   - Library tab
   - Queue tab
   - Job tab
   - Logs tab

6. Implement a centralized keymap with Vim bindings from day one.

7. Make the archive root configurable.

Default development root:

```text
/mnt/archive
```

8. Scan recursively for `.mkv`.

9. Display discovered files in a table.

10. Implement job model/state machine.

11. Implement a fake runner that simulates:
    - frame progress
    - FPS
    - ETA
    - completion

12. Make the TUI polished before connecting real FFmpeg.

13. Once fake queue behavior is solid, connect `ffprobe`.

Do NOT immediately wire production restoration commands into the first UI commit.

Build the application architecture first.

---

# UX Expectations

The TUI should feel polished.

Desired qualities:

- Vim-first navigation
- keyboard-first
- no mouse dependency
- arrow keys optional but supported if easy
- responsive to terminal resizing
- consistent key hints
- modal confirmation for cancel/destructive operations
- scrollable logs
- visible job status icons
- useful error messages
- no raw FFmpeg spam on normal screens
- stable focus model
- searchable/filterable lists
- `Esc` consistently backs out of modals/views
- `q` should not unexpectedly kill an active job

Suggested global keys:

```text
j / k       move down/up
h / l       parent/open or previous/next pane
gg / G      first/last
Ctrl-d/u    half-page down/up
Ctrl-f/b    page down/up
/           filter/search
n / N       next/previous result
?           help
Esc         back
q           back/quit depending context
```

Contextual keys:

```text
Space       select
Enter       open
a           select all
r           restore
b           benchmark
p           pause if supported
s           skip/cancel
l           open logs where it does not conflict with pane navigation
```

Resolve key conflicts intentionally and document them in the help overlay.

---

# Important Design Principle

DiscForge should orchestrate proven external tools rather than reimplement them.

Use:

```text
ffprobe       media inspection
VapourSynth   restoration
FFmpeg        encoding/muxing
x265          HEVC
Sonarr        episode metadata/naming
Jellyfin      library metadata/playback
```

DiscForge's value is:

```text
queue management
visibility
progress
ETA
repeatability
safety
workflow orchestration
```

not reinventing the underlying media stack.

---

# Definition of Success for v0.1

From `ubuntu-server`, the user should be able to run:

```bash
discforge
```

navigate entirely with Vim keys, select:

```text
Viva La Bam
→ Season 01
→ Disc 01
→ A1_t05.mkv
```

press:

```text
Restore
```

and watch:

```text
QTGMC processing
x265 encoding
frame progress
FPS
CPU temperature
ETA
muxing
validation
```

until DiscForge reports:

```text
✓ A1_t05.mkv
Restoration complete
Ready for Sonarr
```

while the original:

```text
/mnt/archive/Viva La Bam/Season 01/Disc 01/A1_t05.mkv
```

remains completely untouched.

---

# First Deliverable Expected From the Coding Agent

The first useful milestone should be a Git repository containing:

```text
README.md
PLAN.md
.gitignore
go.mod
cmd/discforge/main.go
internal/tui/...
```

with a working Bubble Tea interface that:

- launches successfully
- has Library / Queue / Job / Logs views
- uses centralized Vim keybindings
- supports terminal resize
- displays fake queue progress
- has no real media-processing side effects yet
- includes basic tests for non-UI state transitions where practical

The initial code should establish clean architecture rather than rushing directly into FFmpeg integration.
