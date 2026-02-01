DELETE FROM devices CASCADE;
ALTER TABLE devices DROP INDEX unique_identifier_per_bridge;
ALTER TABLE devices DROP COLUMN bridgeKey;
ALTER TABLE devices ADD COLUMN adapterId BIGINT UNSIGNED NOT NULL;
ALTER TABLE devices ADD CONSTRAINT unique_identifier_per_bridge UNIQUE (bridgeIdentifier, adapterId);

DELETE FROM groups CASCADE;
ALTER TABLE groups DROP INDEX unique_identifier_per_bridge;
ALTER TABLE groups DROP COLUMN bridgeKey;
ALTER TABLE groups ADD COLUMN adapterId BIGINT UNSIGNED NOT NULL;
ALTER TABLE groups ADD CONSTRAINT unique_identifier_per_bridge UNIQUE (bridgeIdentifier, adapterId);