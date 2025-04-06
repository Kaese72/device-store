CREATE TABLE IF NOT EXISTS deviceTriggers (
    id SERIAL PRIMARY KEY,
    deviceId BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT unique_triggers_per_device UNIQUE (deviceId, name),
    FOREIGN KEY (deviceId) REFERENCES devices(id) ON DELETE CASCADE
);