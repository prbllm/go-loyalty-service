-- Таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    number      TEXT      PRIMARY KEY,
    status      TEXT      NOT NULL,
    accrual     BIGINT,   -- NULL = нет начисления
    goods       JSONB     -- состав заказа
);
