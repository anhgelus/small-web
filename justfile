dev:
    bun run build
    go run . -dev

build:
    bun run build
    GOAMD64=v3 go build -ldflags "-s" .
