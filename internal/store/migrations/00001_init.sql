-- +goose Up
CREATE TABLE case_types (
    id INTEGER PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE motions (
    id INTEGER PRIMARY KEY,
    case_type_id INTEGER NOT NULL REFERENCES case_types(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    UNIQUE(case_type_id, slug)
);

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

-- +goose Down
DROP TABLE issues;
DROP TABLE motions;
DROP TABLE case_types;
