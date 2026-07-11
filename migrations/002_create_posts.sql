CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doc VARCHAR(140) NOT NULL DEFAULT '',
    image_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT posts_content_required
        CHECK (char_length(trim(doc)) > 0 OR image_url IS NOT NULL)
);
