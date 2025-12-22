package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations
var migrations embed.FS

var nameReg = regexp.MustCompile(`(\d{3})_[a-zA-Z_-]+.sql`)

func ConnectDatabase(file string) *sql.DB {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared", file))
	if err != nil {
		panic(err)
	}
	return db
}

func RunMigration(ctx context.Context, db *sql.DB) error {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	type runMig struct {
		val string
		n   int
	}
	var toRun []runMig
	for _, e := range entries {
		rawId := nameReg.FindStringSubmatch(e.Name())
		id, err := strconv.Atoi(rawId[1])
		if err != nil {
			return err
		}
		b, err := migrations.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		slog.Debug("loading migration", "n", id, "file", e.Name(), "content", string(b))
		toRun = append(toRun, runMig{
			val: string(b), n: id,
		})
	}
	if len(toRun) == 0 {
		return nil
	}
	slices.SortFunc(toRun, func(a, b runMig) int {
		return a.n - b.n
	})
	for _, m := range toRun {
		slog.Info("migrating", "n", m.n)
		_, err := db.ExecContext(ctx, m.val)
		if err != nil {
			return err
		}
	}
	return nil
}
