-- +goose Up
CREATE TABLE chirps (id UUID PRIMARY KEY, created_at TIMESTAMP, updated_at TIMESTAMP, body TEXT,
    user_id UUID REFERENCES users ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;
