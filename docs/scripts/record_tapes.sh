#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

if ! command -v vhs >/dev/null; then
    echo "vhs is not installed. Run 'make tape' to use the helper container or" >&2
    echo "install it from https://github.com/charmbracelet/vhs" >&2
    exit 1
fi

render_tape() {
    local tape_file=$1
    local gif_file=$2
    echo "Rendering $tape_file to $gif_file..."
    vhs -o "docs/assets/$gif_file" "docs/$tape_file"
    rm -f "docs/${tape_file%.tape}.cast"
}

render_tape create_connection.tape create_connection.gif
render_tape client_view.tape client_view.gif
