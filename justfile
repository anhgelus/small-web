dev:
    deno run build
    go run . -dev

build:
    rm -fr dist || echo "ok"
    deno run build
    go build -ldflags "-w" .
