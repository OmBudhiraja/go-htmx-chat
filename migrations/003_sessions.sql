-- +goose Up
CREATE TABLE sessions (
    token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE sessions;