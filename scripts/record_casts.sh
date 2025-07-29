#!/usr/bin/env bash
# Record GoEmqutiti demos using asciinema and expect.
# Requires asciinema, agg, and expect.
set -euo pipefail

run_cast() {
    local script=$1
    local outfile=$2
    asciinema rec -q --overwrite --cols 80 --rows 24 -c "expect $script" "$outfile"
}

DIR=$(dirname "$0")/..
cd "$DIR"

run_cast "$DIR/scripts/create_connection.exp" docs/create_connection_profile.cast
run_cast "$DIR/scripts/client_view.exp" docs/client_view.cast
run_cast "$DIR/scripts/create_import.exp" docs/create_import.cast
run_cast "$DIR/scripts/create_headless_trace.exp" docs/create_headless_trace.cast
run_cast "$DIR/scripts/view_headless_trace.exp" docs/view_headless_trace.cast

agg docs/create_connection_profile.cast docs/create_connection_profile.gif
agg docs/client_view.cast docs/client_view.gif
agg docs/create_import.cast docs/create_import.gif
agg docs/create_headless_trace.cast docs/create_headless_trace.gif
agg docs/view_headless_trace.cast docs/view_headless_trace.gif
