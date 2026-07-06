package backend

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
)

type AssetData struct {
	Src      string
	Checksum string
}

type key uint

const (
	logger key = iota
	assets
	debug
	db
	cfg
	ipAddress
	connected
	assetsFS
)

func ContextLogger(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(logger).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return l
}

func SetContextLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, logger, l)
}

func ContextDB(ctx context.Context) *sql.DB {
	return ctx.Value(db).(*sql.DB)
}

func setContextDB(ctx context.Context, conn *sql.DB) context.Context {
	return context.WithValue(ctx, db, conn)
}

func ContextAssets(ctx context.Context) any {
	return ctx.Value(assets).(map[string]AssetData)
}

func ContextDebug(ctx context.Context) bool {
	return ctx.Value(debug).(bool)
}

func ContextConfig[T any](ctx context.Context) T {
	return ctx.Value(cfg).(T)
}

func ContextAssetsFS(ctx context.Context) fs.FS {
	return ctx.Value(assetsFS).(fs.FS)
}

func SetContextAssetsFS(ctx context.Context, f fs.FS) context.Context {
	return context.WithValue(ctx, assetsFS, f)
}

func ContextIP(ctx context.Context) string {
	return ctx.Value(ipAddress).(string)
}

func SetContextIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipAddress, ip)
}

func ContextConnnected(ctx context.Context) bool {
	return ctx.Value(connected).(bool)
}

func SetContextConnected(ctx context.Context, ok bool) context.Context {
	return context.WithValue(ctx, connected, ok)
}

func SetContext(
	ctx context.Context,
	c any,
	assetsFS map[string]AssetData,
	debugEnabled bool,
	db *sql.DB,
) context.Context {
	ctx = context.WithValue(ctx, assets, assetsFS)
	ctx = context.WithValue(ctx, debug, debugEnabled)
	ctx = SetContextLogger(ctx, slog.Default())
	ctx = context.WithValue(ctx, cfg, c)
	ctx = setContextDB(ctx, db)
	return ctx
}

func CloneContext(parent, source context.Context) context.Context {
	ctx := context.WithValue(parent, assets, ContextAssets(source))
	ctx = context.WithValue(ctx, debug, ContextDebug(source))
	ctx = SetContextLogger(ctx, ContextLogger(source))
	ctx = context.WithValue(ctx, cfg, ContextConfig[any](source))
	ctx = context.WithValue(ctx, ipAddress, ContextIP(source))
	ctx = setContextDB(ctx, ContextDB(source))
	return ctx
}
