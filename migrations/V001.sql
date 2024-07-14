CREATE TABLE IF NOT EXISTS devices (
    id SERIAL PRIMARY KEY,
    bridgeIdentifier VARCHAR(255) NOT NULL,
    bridgeKey VARCHAR(255) NOT NULL,
    CONSTRAINT unique_identifier_per_bridge UNIQUE (bridgeIdentifier, bridgeKey)
);

CREATE TABLE IF NOT EXISTS deviceAttributes (
    id SERIAL PRIMARY KEY,
    deviceId BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    booleanValue BOOLEAN,
    numericValue NUMERIC,
    textValue VARCHAR(255),
    CONSTRAINT unique_attributes_per_device UNIQUE (deviceId, name),
    FOREIGN KEY (deviceId) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS deviceCapabilities (
    id SERIAL PRIMARY KEY,
    deviceId BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    lastSeen TIMESTAMP NOT NULL,

    CONSTRAINT unique_capabilities_per_device UNIQUE (deviceId, name),
    FOREIGN KEY (deviceId) REFERENCES devices(id) ON DELETE CASCADE
);
