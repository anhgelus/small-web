package storage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"git.anhgelus.world/anhgelus/small-web/backend/log"
)

const IPAddressKey = "ip_address"

type loadRequest struct {
	target  string
	referer string
}

type loaded struct {
	data map[string]loadRequest
	mu   *sync.RWMutex
}

func (l *loaded) Get(ip, target string) (loadRequest, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	val, ok := l.data[ip]
	if !ok {
		return loadRequest{}, false
	}
	return val, val.target == target
}

func (l *loaded) Add(ip, referer, target string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.data[ip] = loadRequest{referer: referer, target: target}
}

func (l *loaded) Remove(ip, target string) {
	if _, ok := l.Get(ip, target); ok {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.data, ip)
	}
}

var load = loaded{
	data: make(map[string]loadRequest),
	mu:   new(sync.RWMutex),
}

// using /assets/styles.css to detect if a page is loaded → majority of bots will not load this
const HumanPageLoad = "/assets/styles.css"

func getReferer(r *http.Request, domain string) (string, bool) {
	ref := r.Header.Get("Referer")
	if len(ref) == 0 {
		return "", false
	}
	refUrl, err := url.Parse(ref)
	if err != nil {
		return "", false
	}
	ref = refUrl.Host
	if len(ref) == 0 {
		return "", false
	}
	if ref == domain || ref == fmt.Sprintf("localhost:%d", 8000) {
		ref = refUrl.Path
		if !strings.HasPrefix(ref, "/") {
			ref = "/" + ref
		}
		if ref == r.URL.Path || strings.HasPrefix(ref, "/admin") || ref == "/favicon.ico" {
			return ref, false
		}
	}
	return ref, true
}

func UpdateStats(ctx context.Context, r *http.Request, domain string) error {
	target := r.URL.Path
	if target == HumanPageLoad {
		return humanLoad(ctx, r, domain)
	}
	if strings.HasPrefix(target, "/admin") {
		return nil
	}
	ref, ok := getReferer(r, domain)
	if !ok {
		return nil
	}
	ip := ctx.Value(IPAddressKey).(string)
	load.Add(ip, ref, target)
	go func(ip, target string) {
		time.Sleep(5 * time.Second)
		load.Remove(ip, target)
	}(ip, target)
	return nil
}

func humanLoad(ctx context.Context, r *http.Request, domain string) error {
	ip := ctx.Value(IPAddressKey).(string)
	ref, ok := getReferer(r, domain)
	if !ok {
		return nil
	}
	lr, ok := load.Get(ip, ref)
	if !ok {
		lr.referer = "?"
		lr.target = ref
	}
	db := getDB(ctx)
	rows, err := db.QueryContext(ctx, "SELECT id, visit FROM stats WHERE origin = ? AND target = ?", lr.referer, lr.target)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			log.GetLogger(ctx).Debug("stats updated")
			load.Remove(ip, lr.target)
		}
	}()
	if !rows.Next() {
		_, err = db.ExecContext(ctx, "INSERT INTO stats (origin, target, visit) VALUES (?, ?, 1)", lr.referer, lr.target)
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
