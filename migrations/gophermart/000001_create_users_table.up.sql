CREATE SCHEMA IF NOT EXISTS gophermart;

CREATE TABLE gophermart.users (
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    withdrawn BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_users_login ON gophermart.users(login);

