-- Отключаем вывод уведомлений для чистоты логов
SET client_min_messages TO WARNING;

-- Таблица для хранения сырых данных страниц, загруженных краулером
CREATE TABLE IF NOT EXISTS pages (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    title TEXT,
    html_content TEXT,
    last_crawled_at TIMESTAMPTZ,
    is_indexed BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для ускорения поиска по URL
CREATE INDEX IF NOT EXISTS idx_pages_url ON pages(url);
-- Индекс для быстрой выборки неиндексированных страниц
CREATE INDEX IF NOT EXISTS idx_pages_not_indexed ON pages(id) WHERE is_indexed = FALSE;

-- Таблица для хранения графа ссылок (какая страница на какую ссылается)
CREATE TABLE IF NOT EXISTS links (
    from_page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    to_page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    PRIMARY KEY (from_page_id, to_page_id)
);

-- Таблицы для инвертированного индекса
-- Хранит уникальные термины (токены)
CREATE TABLE IF NOT EXISTS terms (
    id SERIAL PRIMARY KEY,
    term TEXT NOT NULL UNIQUE
);

-- Связующая таблица (postings list)
-- Хранит информацию о том, в каких документах и как часто встречается термин
CREATE TABLE IF NOT EXISTS postings (
    term_id INTEGER NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    page_id BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    frequency INTEGER NOT NULL, -- Количество вхождений термина на странице
    PRIMARY KEY (term_id, page_id)
);