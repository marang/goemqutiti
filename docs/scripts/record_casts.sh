#!/usr/bin/env bash
# Record cast + convert to GIF using asciinema and expect (inside Docker)

set -euo pipefail

restore_config() {
    mkdir -p /root/.emqutiti
    cat > /root/.emqutiti/config.toml <<EOF
default_profile = "HiveMQ"

[[profiles]]
  name = "HiveMQ"
  schema = "tcp"
  host = "broker.hivemq.com"
  port = 1883
  client_id = "goemqutiti-client"
  ssl_tls = true
  mqtt_version = "3"
  clean_start = false
EOF
}

run_cast() {
    local exp_script=$1
    local cast_file=$2
    local gif_file=$3

    echo "Recording $cast_file..."
    asciinema rec -q --overwrite --cols 80 --rows 36 \
        -c "expect docs/scripts/$exp_script" "docs/$cast_file"

    echo "Converting $cast_file to $gif_file..."
    agg "docs/$cast_file" "docs/$gif_file"
}

# First recording
restore_config
run_cast create_connection.exp create_connection.cast create_connection.gif

# Restore config to ensure HiveMQ still exists
restore_config
run_cast client_view.exp client_view.cast client_view.gif
