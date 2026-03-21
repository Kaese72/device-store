ALTER TABLE groups ADD COLUMN updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

CREATE TRIGGER groupCapabilities_touch_group_insert
AFTER INSERT ON groupCapabilities
FOR EACH ROW
UPDATE groups SET updated = CURRENT_TIMESTAMP WHERE id = NEW.groupId;

CREATE TRIGGER groupCapabilities_touch_group_update
AFTER UPDATE ON groupCapabilities
FOR EACH ROW
UPDATE groups SET updated = CURRENT_TIMESTAMP WHERE id = NEW.groupId;

CREATE TRIGGER groupCapabilities_touch_group_delete
AFTER DELETE ON groupCapabilities
FOR EACH ROW
UPDATE groups SET updated = CURRENT_TIMESTAMP WHERE id = OLD.groupId;

CREATE TRIGGER groupDevices_touch_group_insert
AFTER INSERT ON groupDevices
FOR EACH ROW
UPDATE groups SET updated = CURRENT_TIMESTAMP WHERE id = NEW.groupId;

CREATE TRIGGER groupDevices_touch_group_delete
AFTER DELETE ON groupDevices
FOR EACH ROW
UPDATE groups SET updated = CURRENT_TIMESTAMP WHERE id = OLD.groupId;
