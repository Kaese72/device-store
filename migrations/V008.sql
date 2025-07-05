CREATE TABLE IF NOT EXISTS deviceAttributeAudit (
    id SERIAL PRIMARY KEY,
    deviceId BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    oldBooleanValue BOOLEAN,
    oldNumericValue DECIMAL(10, 4),
    oldTextValue VARCHAR(255),
    newBooleanValue BOOLEAN,
    newNumericValue DECIMAL(10, 4),
    newTextValue VARCHAR(255),
    FOREIGN KEY (deviceId) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS deviceAttributeAuditTrigger
AFTER UPDATE ON deviceAttributes
FOR EACH ROW
INSERT INTO deviceAttributeAudit (
    deviceId,
    name,
    oldBooleanValue,
    oldNumericValue,
    oldTextValue,
    newBooleanValue,
    newNumericValue,
    newTextValue
) VALUES (
    NEW.deviceId,
    NEW.name,
    OLD.booleanValue,
    OLD.numericValue,
    OLD.textValue,
    NEW.booleanValue,
    NEW.numericValue,
    NEW.textValue
);

CREATE TRIGGER IF NOT EXISTS deviceAttributeAuditTriggerInserts
AFTER INSERT ON deviceAttributes
FOR EACH ROW
INSERT INTO deviceAttributeAudit (
    deviceId,
    name,
    newBooleanValue,
    newNumericValue,
    newTextValue
) VALUES (
    NEW.deviceId,
    NEW.name,
    NEW.booleanValue,
    NEW.numericValue,
    NEW.textValue
);

-- From V006.sql
DELETE TABLE IF EXISTS deviceAudits;

CREATE TABLE IF NOT EXISTS deviceCapabilityTriggerAudit (
    id SERIAL PRIMARY KEY,
    deviceId BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    success BOOLEAN NOT NULL,
    errorMessage TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    arguments TEXT,
    FOREIGN KEY (deviceId) REFERENCES devices(id) ON DELETE CASCADE
);
