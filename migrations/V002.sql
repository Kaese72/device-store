CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    bridgeKey VARCHAR(255) NOT NULL,
    bridgeIdentifier VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    CONSTRAINT unique_identifier_per_bridge UNIQUE (bridgeIdentifier, bridgeKey)
);

CREATE TABLE IF NOT EXISTS groupCapabilities (
    id SERIAL PRIMARY KEY,
    groupId BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,

    CONSTRAINT unique_capabilities_per_group UNIQUE (groupId, name),
    FOREIGN KEY (groupId) REFERENCES groups(id) ON DELETE CASCADE
);