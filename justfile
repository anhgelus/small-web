dev:
    bun run build
    go run . -dev

build:
    bun run build
    go build -ldflags "-s" .