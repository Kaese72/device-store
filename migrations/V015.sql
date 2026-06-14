CREATE TABLE IF NOT EXISTS groupCapabilityTriggerAudit (
    id SERIAL PRIMARY KEY,
    groupId BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    success BOOLEAN NOT NULL,
    errorMessage TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    arguments TEXT,
    FOREIGN KEY (groupId) REFERENCES groups(id) ON DELETE CASCADE
);
