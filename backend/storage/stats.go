package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

type loaded struct {
	data map[string]struct{}
	mu   *sync.RWMutex
}

func (l *loaded) Has(k string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.data[k]
	return ok
}

func (l *loaded) Add(k string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.data[k] = struct{}{}
}

func (l *loaded) Remove(k string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.data, k)
}

func newLoaded() *loaded {
	return &loaded{
		data: make(map[string]struct{}),
		mu:   new(sync.RWMutex),
	}
}

var load = newLoaded()

func UpdateStats(ctx context.Context, r *http.Request, domain string) error {
	target := r.URL.Path
	if strings.HasPrefix(target, "/static") || strings.HasPrefix(target, "/admin") {
		return nil
	}
	ref := r.Header.Get("Referer")
	if ref == "" {
		return nil
	}
	refUrl, err := url.Parse(ref)
	if err != nil {
		return nil
	}
	ref = refUrl.Host
	if ref == domain || ref == fmt.Sprintf("localhost:%d", 8000) {
		ref = refUrl.Path
		if ref == target || strings.HasPrefix(ref, "/admin") || ref == "/favicon.ico" {
			return nil
		}
	}
	// using /assets/styles.css to detect if a page is loaded → majority of bots will not load this
	if target == "/assets/styles.css" {
		target = ref
		ref = "?"
	}
	if load.Has(target) {
		return nil
	}
	db := getDB(ctx)
	rows, err := db.QueryContext(ctx, "SELECT id, visit FROM stats WHERE origin = ? AND target = ?", ref, target)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			slog.Debug("stats updated")
			load.Add(target)
			go func(target string) {
				time.Sleep(5 * time.Second)
				load.Remove(target)
			}(target)
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

type StatsRow struct {
	Origin string
	Target string
	Visit  uint
}

const StatsPerPage = 25

func GetStatsRows(ctx context.Context, page uint) ([]StatsRow, error) {
	rows, err := getDB(ctx).QueryContext(
		ctx,
		"SELECT origin, target, visit FROM stats ORDER BY visit DESC LIMIT ? OFFSET ?",
		StatsPerPage, (page-1)*StatsPerPage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	statRows := make([]StatsRow, StatsPerPage)
	var i uint8
	for i = 0; rows.Next(); i++ {
		var stat StatsRow
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

func GetUnionStatsRows(ctx context.Context) ([]StatsRow, error) {
	rows, err := getDB(ctx).QueryContext(ctx, "SELECT target, visit FROM stats ORDER BY visit DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := make(map[string]uint)
	for rows.Next() {
		var stat StatsRow
		err = rows.Scan(&stat.Target, &stat.Visit)
		if err != nil {
			return nil, err
		}
		if _, ok := data[stat.Target]; !ok {
			data[stat.Target] = stat.Visit
		} else {
			data[stat.Target] += stat.Visit
		}
	}
	var statRows []StatsRow
	for k, v := range data {
		statRows = append(statRows, StatsRow{
			Target: k,
			Visit:  v,
		})
	}
	slices.SortFunc(statRows, func(a, b StatsRow) int {
		return int(b.Visit) - int(a.Visit)
	})
	return statRows, nil
}
