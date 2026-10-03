package storage

import (
	"context"
	"database/sql"
	"errors"

	"anhgelus.world/xrpc/atproto"
)

type PublishedDocument struct {
	Path          string
	RecordKey     atproto.RecordKey
	CID           *atproto.CIDAsString
	ImageUploaded bool
}

func RecordKey(ctx context.Context, db *sql.DB, path string) (*atproto.RecordKey, error) {
	row := db.QueryRowContext(
		ctx,
		"SELECT record_key FROM atproto_documents WHERE path = ?",
		path)
	var raw string
	err := row.Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return new(atproto.RecordKey(raw)), nil
}

func PublishedDocuments(ctx context.Context, db *sql.DB) (map[string]PublishedDocument, error) {
	rows, err := db.QueryContext(
		ctx,
		"SELECT path, record_key, cid, image_uploaded FROM atproto_documents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mp := make(map[string]PublishedDocument)
	for rows.Next() {
		var path, recordKey, cid string
		var imgUploaded bool
		err = rows.Scan(&path, &recordKey, &cid, &imgUploaded)
		if err != nil {
			return nil, err
		}
		var doc PublishedDocument
		doc.CID, err = atproto.ParseCIDString(cid)
		if err != nil {
			return nil, err
		}
		doc.RecordKey, err = atproto.ParseRecordKey(recordKey)
		if err != nil {
			return nil, err
		}
		doc.Path = path
		doc.ImageUploaded = imgUploaded
		mp[path] = doc
	}
	return mp, nil
}

func SetPublishedDocument(ctx context.Context, db *sql.DB, doc PublishedDocument) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO atproto_documents (path, record_key, cid, image_uploaded) VALUES (?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		record_key = excluded.record_key,
		cid = excluded.cid,
		image_uploaded = MAX(excluded.image_uploaded,image_uploaded)`,
		doc.Path, doc.RecordKey, doc.CID.String(), doc.ImageUploaded)
	return err
}
