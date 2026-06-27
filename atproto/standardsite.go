package atproto

import (
	"context"
	"io"
	"io/fs"
	"mime"
	"strings"
	"time"

	site "anhgelus.world/goat-site"
	"anhgelus.world/xrpc"
	"anhgelus.world/xrpc/atproto"
)

const tidGeneratorClockId uint = 0

type Site struct {
	*site.Publication
	URL    atproto.RawURI
	RKey   atproto.RecordKey
	Files  fs.FS
	genTid *atproto.TIDGenerator
}

func LoadSite(
	ctx context.Context,
	client xrpc.Client,
	files fs.FS,
	did *atproto.DID,
	rkey atproto.RecordKey,
) (*Site, error) {
	pub, err := xrpc.GetRecord[*site.Publication](
		ctx, client, did, rkey, nil)
	if err != nil {
		return nil, err
	}
	return &Site{
		pub.Value,
		pub.URI,
		rkey,
		files,
		atproto.NewTIDGenerator(tidGeneratorClockId),
	}, nil
}

func CreateSite(
	ctx context.Context,
	client xrpc.Client,
	files fs.FS,
	did *atproto.DID,
	rkey atproto.RecordKey,
	pub *site.Publication,
) (*Site, error) {
	res, err := xrpc.PutRecord(
		ctx, client, pub, rkey, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Site{
		pub,
		*res.URI,
		rkey,
		files,
		atproto.NewTIDGenerator(tidGeneratorClockId),
	}, nil
}

func (s *Site) PublishDoc(
	ctx context.Context,
	client xrpc.Client,
	title string,
	path string,
	publishedAt time.Time,
	description string,
	imagePath *string,
	tags []string,
	contributors []*site.Contributor,
) (*xrpc.SendRecordResult, atproto.RecordKey, error) {
	var blob *xrpc.Blob
	if imagePath != nil {
		f, err := s.Files.Open(*imagePath)
		if err != nil {
			return nil, "", err
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, "", err
		}
		typ := mime.TypeByExtension("." + strings.Split(*imagePath, ".")[1])
		blob, err = xrpc.UploadBlob(ctx, client, typ, b)
		if err != nil {
			return nil, "", err
		}
	}
	doc := &site.Document{
		Site:         site.FromRawAT(s.URL),
		Title:        title,
		PublishedAt:  publishedAt,
		Path:         &path,
		Description:  &description,
		Tags:         tags,
		Contributors: contributors,
		CoverImage:   blob,
	}
	tid := s.genTid.Next().RecordKey()
	res, err := xrpc.PutRecord(
		ctx, client, doc, tid, nil, nil, nil)
	return res, tid, err
}
