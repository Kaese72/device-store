--- Since we do not implement backwards compatability for this change, we will just have the system reset itself
DELETE FROM devices CASCADE;
ALTER TABLE devices DROP COLUMN bridgeKey;
ALTER TABLE devices ADD COLUMN adapterId BIGINT UNSIGNED NOT NULL;

DELETE FROM groups CASCADE;
ALTER TABLE groups DROP COLUMN bridgeKey;
ALTER TABLE groups ADD COLUMN adapterId BIGINT UNSIGNED NOT NULL;