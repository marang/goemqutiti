#!/usr/bin/env bash
set -euo pipefail

render_tape() {
    local tape_file=$1
    local gif_file=$2
    echo "Rendering $tape_file to $gif_file..."
    vhs "docs/$tape_file" > "docs/assets/$gif_file"
}

render_tape create_connection.tape create_connection.gif
render_tape client_view.tape client_view.gif
