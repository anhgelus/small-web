package backend

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

var trimRefererReg = regexp.MustCompile(`https?://([a-z-0-9.]+(:\d+)?)/.*`)

func getDB(ctx context.Context) *sql.DB {
	return ctx.Value(dbKey).(*sql.DB)
}

func UpdateStats(ctx context.Context, r *http.Request) error {
	target := r.URL.Path
	if strings.HasPrefix(target, "/assets") || strings.HasPrefix(target, "/static") {
		return nil
	}
	ref := r.Header.Get("Referer")
	if ref == "" {
		return nil
	}
	subs := trimRefererReg.FindStringSubmatch(ref)
	if len(subs) < 2 {
		return nil
	}
	ref = subs[1]
	if ref == ctx.Value(configKey).(*Config).Domain || ref == fmt.Sprintf("localhost:%d", 8000) {
		ref = subs[0][strings.Index(subs[0], ref)+len(ref):]
		if ref == target {
			return nil
		}
	}
	db := getDB(ctx)
	rows, err := db.QueryContext(ctx, "SELECT id, visit FROM stats WHERE origin = ? AND target = ?", ref, target)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			slog.Debug("stats updated")
		}
	}()
	if !rows.Next() {
		_, err = db.ExecContext(ctx, "INSERT INTO stats (origin, target, visit) VALUES (?, ?, 1)", ref, target)
		return err
	}
	var id uint
	var nb uint
	err = rows.Scan(&id, &nb)
	if err != nil {
		return err
	}
	err = rows.Close()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "UPDATE stats SET visit = ? WHERE id = ?", nb+1, id)
	return err
}
