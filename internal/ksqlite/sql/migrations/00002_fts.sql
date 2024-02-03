-- FTS FOOD LOCALES
CREATE VIRTUAL TABLE fts_foods_locales USING fts5(
	food_id,
	lang_id,
	value,
	content=foods_locales,
	content_rowid=rowid
);
CREATE TRIGGER foods_locales_ai AFTER INSERT ON foods_locales BEGIN
  INSERT INTO fts_foods_locales(rowid, food_id, lang_id, value)
	VALUES (new.rowid, new.food_id, new.lang_id, new.value);
END;
CREATE TRIGGER foods_locales_ad AFTER DELETE ON foods_locales BEGIN
  INSERT INTO fts_foods_locales(fts_foods_locales, rowid, food_id, lang_id, value)
	VALUES('delete', old.rowid, old.food_id, old.lang_id, old.value);
END;
CREATE TRIGGER foods_locales_au AFTER UPDATE ON foods_locales BEGIN
  INSERT INTO fts_foods_locales(fts_foods_locales, rowid, food_id, lang_id, value)
	VALUES('delete', old.rowid, old.food_id, old.lang_id, old.value);
  INSERT INTO fts_foods_locales(rowid, food_id, lang_id, value)
	VALUES (new.rowid, new.food_id, new.lang_id, new.value);
END;
