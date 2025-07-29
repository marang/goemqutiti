#!/usr/bin/env bash
# Record GoEmqutiti demos using asciinema. Requires asciinema, agg,
# and the util-linux `script` command.
set -euo pipefail

run_cast() {
    local cmd=$1
    local outfile=$2
    local input=$3
    asciinema rec -q --overwrite --cols 80 --rows 24 \
        -c "script -q /dev/null -c '$cmd' < $input" "$outfile"
    stty sane
    rm -f "$input"
}

DIR=$(dirname "$0")/..
cd "$DIR"

# create connection profile
input=$(mktemp)
printf '\x02\r' >"$input"
printf 'local\r' >>"$input"
printf 'tcp\r' >>"$input"
printf 'localhost\r' >>"$input"
printf '1883\r' >>"$input"
printf '\x04' >>"$input"
run_cast "HOME=/root ./goemqutiti -p HiveMQ" docs/create_connection_profile.cast "$input"

# client view demo
input=$(mktemp)
printf '\n' >"$input"
printf 'test/topic\n' >>"$input"
printf '\t' >>"$input"
printf 'Hello' >>"$input"
printf '\x13' >>"$input"  # Ctrl+S
printf '\x04' >>"$input"  # Ctrl+D
run_cast "HOME=/root ./goemqutiti -p HiveMQ" docs/client_view.cast "$input"

# import wizard
input=$(mktemp)
printf '\x04' >"$input"
run_cast "HOME=/root ./goemqutiti -p HiveMQ --import data.csv" docs/create_import.cast "$input"

# create headless trace
input=$(mktemp)
printf '\x03' >"$input"
run_cast "HOME=/root ./goemqutiti --trace demo --topics 'sensors/#' -p HiveMQ" \
    docs/create_headless_trace.cast "$input"

# view headless trace
input=$(mktemp)
printf '\x04' >"$input"
run_cast "HOME=/root ./goemqutiti --trace demo --topics 'sensors/#' -p HiveMQ" \
    docs/view_headless_trace.cast "$input"

agg docs/create_connection_profile.cast docs/create_connection_profile.gif
agg docs/client_view.cast docs/client_view.gif
agg docs/create_import.cast docs/create_import.gif
agg docs/create_headless_trace.cast docs/create_headless_trace.gif
agg docs/view_headless_trace.cast docs/view_headless_trace.gif
