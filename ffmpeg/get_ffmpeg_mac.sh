#!/bin/bash

set -e

echo "📥 Téléchargement des versions FULL de FFmpeg pour : macOS, Linux, Windows"

FFMPEG_URL="https://evermeet.cx/ffmpeg/ffmpeg-7.1.1.zip"
FFPROBE_URL="https://evermeet.cx/ffmpeg/ffprobe-7.1.1.zip"

curl -L "$FFMPEG_URL" -o "./ffmpeg_mac.zip"
curl -L "$FFPROBE_URL" -o "./ffprobe_mac.zip"

unzip -q "./ffmpeg_mac.zip"
unzip -q "./ffprobe_mac.zip"

rm "./ffmpeg_mac.zip"
rm "./ffprobe_mac.zip"

mv "ffmpeg" "ffmpeg_mac"
mv "ffprobe" "ffprobe_mac"

chmod +x "ffmpeg_mac"
chmod +x "ffprobe_mac"
