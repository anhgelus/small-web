dev:
    bun run build
    go run . -dev

build:
    rm -fr dist || echo "ok"
    bun run build
    GOAMD64=v3 go build -ldflags "-s" .
