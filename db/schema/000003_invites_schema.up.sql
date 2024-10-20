CREATE TABLE invites (
    id BIGSERIAL PRIMARY KEY,
    invite_code VARCHAR(255) UNIQUE,
    is_used BOOL DEFAULT FALSE,
    group_id INTEGER,
    used_by_user_id BIGINT UNIQUE,
    created_by_user_id BIGINT,
    expiration_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (used_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
);