CREATE INDEX IF NOT EXISTS recipes_title_idx ON recipes USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS recipess_categories_idx ON recipes USING GIN (categories);
