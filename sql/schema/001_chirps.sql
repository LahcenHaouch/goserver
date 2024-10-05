-- +goose Up
CREATE TABLE chirps (id UUID, created_at TIMESTAMP, updated_at TIMESTAMP, body TEXT, user_id UUID);

-- +goose Down
DROP TABLE chirps;
