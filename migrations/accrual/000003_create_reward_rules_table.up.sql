-- Таблица правил вознаграждений
CREATE TABLE IF NOT EXISTS accrual.reward_rules (
    match        TEXT             PRIMARY KEY,
    reward       NUMERIC(10,2)    NOT NULL,
    reward_type  TEXT             NOT NULL CHECK (reward_type IN ('%', 'pt'))
);
