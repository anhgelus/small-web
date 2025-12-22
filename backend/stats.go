package backend

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

var trimRefererReg = regexp.MustCompile(`https?://([a-z-0-9.]+(:\d+)?)/.*`)

func getDB(ctx context.Context) *sql.DB {
	return ctx.Value(dbKey).(*sql.DB)
}

func UpdateStats(ctx context.Context, r *http.Request) error {
	target := r.URL.Path
	if strings.HasPrefix(target, "/assets") ||
		strings.HasPrefix(target, "/static") ||
		strings.HasPrefix(target, "/admin") {
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
		if ref == target || strings.HasPrefix(ref, "/admin") || ref == "/favicon.ico" {
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

type statRow struct {
	Origin string
	Target string
	Visit  uint
}

const statPerPage = 25

func GetStatRows(ctx context.Context, page uint) ([]statRow, error) {
	rows, err := getDB(ctx).QueryContext(
		ctx,
		"SELECT origin, target, visit FROM stats ORDER BY visit DESC LIMIT ? OFFSET ?",
		statPerPage, (page-1)*statPerPage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	statRows := make([]statRow, statPerPage)
	var i uint8
	for i = 0; rows.Next(); i++ {
		var stat statRow
		err = rows.Scan(&stat.Origin, &stat.Target, &stat.Visit)
		if err != nil {
			return nil, err
		}
		statRows[i] = stat
	}
	if i == 0 {
		return nil, nil
	}
	return statRows[:i], nil
}

type adminData struct {
	*data
	Rows        []statRow
	PagesNumber int
	CurrentPage int
}

func HandleAdmin(r *chi.Mux) {
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !ctx.Value(loginKey).(bool) {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		d := new(adminData)
		d.data = new(data)
		rawPage := r.URL.Query().Get("page")
		page := 1
		var err error
		if rawPage != "" {
			page, err = strconv.Atoi(rawPage)
			if err != nil || page < 1 {
				slog.Warn("invalid page number", "rawPage", rawPage)
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}
		d.Rows, err = GetStatRows(ctx, uint(page))
		if err != nil {
			panic(err)
		}
		d.PagesNumber = page + max(len(d.Rows)-statPerPage+1, 0)
		d.CurrentPage = page
		d.handleGeneric(w, r, "admin", d)
	})
}
