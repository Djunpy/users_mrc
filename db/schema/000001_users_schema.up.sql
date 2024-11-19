CREATE TYPE user_types AS ENUM('company', 'job_seeker');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(120) NOT NULL,
    first_name VARCHAR(120),
    last_name VARCHAR(120),
    password VARCHAR(120) NOT NULL,
    is_deleted BOOL DEFAULT false,
    auth_source VARCHAR(120) NOT NULL,
    updated_at TIMESTAMPTZ,
    last_token_update TIMESTAMPTZ,
    verified_email BOOL DEFAULT false,
    user_type user_types DEFAULT 'job_seeker',
    is_banned BOOL DEFAULT false,               -- Указывает, заблокирован ли пользователь.
    ban_reason TEXT,                            -- Причина блокировки пользователя.
    banned_at TIMESTAMPTZ,                       -- Время, когда пользователь был заблокирован.
    date_joined TIMESTAMPTZ DEFAULT NOW(),
    sexy ENUM('f', 'm') NOT NULL
    UNIQUE (email)
);

CREATE TABLE phones (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    number BIGINT UNIQUE NOT NULL,
    country_code VARCHAR(5) NOT NULL,
    verified BOOL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
