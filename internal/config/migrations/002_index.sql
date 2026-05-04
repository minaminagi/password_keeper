CREATE INDEX IF NOT EXISTS idx_items_title
ON items(title);
CREATE INDEX IF NOT EXISTS idx_items_updated_at
ON items(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_item_tags_tag_id
ON item_tags(tag_id);
