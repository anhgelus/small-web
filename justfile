dev:
    bun run build
    go run . -dev -config config.toml

build:
    bun run build
    go build -ldflags "-s" .