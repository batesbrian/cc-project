-- +goose Up
CREATE TABLE groups (
    id INTEGER PRIMARY KEY,
    motion_id INTEGER NOT NULL REFERENCES motions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    UNIQUE(motion_id, slug)
);

DROP TABLE issues;
CREATE TABLE issues (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    template_path TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    last_seen TEXT NOT NULL DEFAULT '',
    UNIQUE(group_id, slug)
);


-- +goose Down
DROP TABLE issues;
DROP TABLE groups;

CREATE TABLE issues (
    id INTEGER PRIMARY KEY,
    motion_id INTEGER NOT NULL REFERENCES motions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    template_path TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    last_seen TEXT NOT NULL DEFAULT '',
    UNIQUE(motion_id, slug)
);
