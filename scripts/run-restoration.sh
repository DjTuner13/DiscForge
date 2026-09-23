#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "usage: $0 SOURCE_MKV VAPOURSYNTH_SCRIPT OUTPUT_MKV" >&2
  exit 2
fi

source_mkv=$1
vapoursynth_script=$2
output_mkv=$3

if [[ ! -f "$source_mkv" ]]; then
  echo "source does not exist: $source_mkv" >&2
  exit 1
fi
if [[ ! -f "$vapoursynth_script" ]]; then
  echo "VapourSynth script does not exist: $vapoursynth_script" >&2
  exit 1
fi
case "$output_mkv" in
  /mnt/archive|/mnt/archive/*)
    echo "refusing to write inside read-only archive: $output_mkv" >&2
    exit 1
    ;;
esac
if [[ -e "$output_mkv" ]]; then
  echo "output already exists: $output_mkv" >&2
  exit 1
fi

vspipe_bin=${DISCFORGE_VSPIPE:-/home/dj/src/viva-remaster/.venv/bin/vspipe}
ffmpeg_bin=${DISCFORGE_FFMPEG:-/usr/bin/ffmpeg}
if [[ ! -x "$vspipe_bin" ]]; then
  echo "vspipe is not executable: $vspipe_bin" >&2
  exit 1
fi
if [[ ! -x "$ffmpeg_bin" ]]; then
  echo "ffmpeg is not executable: $ffmpeg_bin" >&2
  exit 1
fi

output_dir=$(dirname "$output_mkv")
mkdir -p "$output_dir"
video_mkv="$output_dir/.$(basename "$output_mkv").video.partial.mkv"
log_file="$output_dir/$(basename "$output_mkv").log"
if [[ -e "$video_mkv" ]]; then
  echo "temporary output already exists: $video_mkv" >&2
  exit 1
fi

cleanup() {
  rm -f "$video_mkv"
}
trap cleanup EXIT INT TERM

: > "$log_file"
echo "Encoding video; log: $log_file"
export DISCFORGE_SOURCE="$source_mkv"
"$vspipe_bin" -c y4m "$vapoursynth_script" - 2>>"$log_file" |
  "$ffmpeg_bin" -f yuv4mpegpipe -i - -an \
    -c:v libx265 -preset medium -crf 16 -pix_fmt yuv420p10le \
    -progress pipe:2 -nostats -y "$video_mkv" 2>>"$log_file"

echo "Muxing audio, subtitles, chapters, and metadata"
"$ffmpeg_bin" -i "$video_mkv" -i "$source_mkv" \
  -map 0:v:0 -map 1:a? -map 1:s? \
  -map_metadata 1 -map_chapters 1 -c copy -n "$output_mkv" \
  2>>"$log_file"

test -s "$output_mkv"
trap - EXIT INT TERM
rm -f "$video_mkv"
echo "Created: $output_mkv"
