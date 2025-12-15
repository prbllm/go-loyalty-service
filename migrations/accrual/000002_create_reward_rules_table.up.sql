-- Таблица правил вознаграждений
CREATE TABLE IF NOT EXISTS reward_rules (
    match        TEXT      PRIMARY KEY,
    reward       BIGINT    NOT NULL,
    reward_type  TEXT      NOT NULL CHECK (reward_type IN ('%', 'pt'))
);
