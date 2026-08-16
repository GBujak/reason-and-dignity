-- +goose Up

CREATE TABLE users (
    id INTEGER AUTOINCREMENT PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    salt TEXT NOT NULL
) STRICT;