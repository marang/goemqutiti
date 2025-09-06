#!/usr/bin/env bash
set -euo pipefail

render_tape() {
    local tape_file=$1
    local gif_file=$2
    echo "Rendering $tape_file to $gif_file..."
    vhs -o "docs/assets/$gif_file" "docs/$tape_file"
    rm -f "docs/${tape_file%.tape}.cast"
}

render_tape create_connection.tape create_connection.gif
render_tape client_view.tape client_view.gif
