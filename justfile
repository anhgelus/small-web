dev:
    deno run build
    go run . -dev

build:
    rm -fr dist || echo "ok"
    deno run build
    go build -ldflags "-w" .

install: build
    install -m 755 small-web /usr/local/bin/
