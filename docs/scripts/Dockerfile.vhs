FROM golang:1.24-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl git fonts-dejavu-core ffmpeg chromium && \
    curl -L https://github.com/tsl0922/ttyd/releases/latest/download/ttyd.x86_64 \
        -o /usr/local/bin/ttyd && \
    chmod +x /usr/local/bin/ttyd && \
    rm -rf /var/lib/apt/lists/*

RUN go install github.com/charmbracelet/vhs@latest

ENV PATH="/go/bin:$PATH"
ENV ROD_BROWSER_PATH=/usr/bin/chromium

WORKDIR /work
ENTRYPOINT ["bash"]
