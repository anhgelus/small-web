package atproto

import (
	"context"
	"mime"
	"os"
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
	genTid *atproto.TIDGenerator
}

func LoadSite(ctx context.Context, client xrpc.Client, did *atproto.DID, rkey atproto.RecordKey) (*Site, error) {
	pub, err := xrpc.GetRecord[*site.Publication](
		ctx, client, did, rkey, nil)
	if err != nil {
		return nil, err
	}
	return &Site{
		pub.Value,
		pub.URI,
		rkey,
		atproto.NewTIDGenerator(tidGeneratorClockId),
	}, nil
}

func CreateSite(
	ctx context.Context,
	client xrpc.Client,
	did *atproto.DID,
	rkey atproto.RecordKey,
	pub *site.Publication,
) (*Site, error) {
	res, err := xrpc.CreateRecord(
		ctx, client, pub, rkey, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Site{
		pub,
		*res.URI,
		rkey,
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
) (*xrpc.SendRecordResult, error) {
	var blob *xrpc.Blob
	if imagePath != nil {
		b, err := os.ReadFile(*imagePath)
		if err != nil {
			return nil, err
		}
		typ := mime.TypeByExtension("." + strings.Split(*imagePath, ".")[1])
		blob, err = xrpc.UploadBlob(ctx, client, typ, b)
		if err != nil {
			return nil, err
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
	return xrpc.CreateRecord(
		ctx, client, doc, atproto.RecordKey(s.genTid.Next()), nil, nil)
}
