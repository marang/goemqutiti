FROM golang:1.24-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl git fonts-dejavu-core ffmpeg ttyd && \
    rm -rf /var/lib/apt/lists/*

RUN go install github.com/charmbracelet/vhs@latest

WORKDIR /app
COPY . .
RUN rm -f emqutiti && \
    go build -o emqutiti ./cmd/emqutiti && \
    chmod +x emqutiti

ENV PATH="/go/bin:$PATH"
ENTRYPOINT ["bash"]
