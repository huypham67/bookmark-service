ALTER TABLE bookmarks ALTER COLUMN code SET NOT NULL;
CREATE UNIQUE INDEX idx_bookmarks_code ON bookmarks(code);