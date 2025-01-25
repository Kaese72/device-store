CREATE TABLE IF NOT EXISTS groupDevices (
    id SERIAL PRIMARY KEY,
    deviceId BIGINT UNSIGNED NOT NULL,
    groupId BIGINT UNSIGNED NOT NULL,
    FOREIGN KEY (deviceId) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (groupId) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT unique_device_in_group UNIQUE (deviceId, groupId)
);
