-- Отключаем вывод уведомлений для чистоты логов
SET client_min_messages TO WARNING;

-- Таблица для хранения данных страниц, загруженных краулером
CREATE TABLE IF NOT EXISTS pages (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    title TEXT,
    html_content TEXT,
    content_tsvector tsvector,
    last_crawled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для ускорения поиска по URL
CREATE INDEX IF NOT EXISTS idx_pages_url ON pages(url);
CREATE INDEX IF NOT EXISTS idx_pages_tsvector ON pages USING GIN (content_tsvector);

-- Таблица для хранения графа ссылок (остается без изменений)
CREATE TABLE IF NOT EXISTS links (
    from_page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    to_page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    PRIMARY KEY (from_page_id, to_page_id)
);