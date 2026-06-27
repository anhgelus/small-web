CREATE TABLE IF NOT EXISTS atproto_documents(
    path TEXT NOT NULL PRIMARY KEY,
    record_key TEXT NOT NULL UNIQUE,
    cid TEXT NOT NULL,
    image_uploaded BOOLEAN DEFAULT FALSE,
    UNIQUE(record_key, cid)
);
