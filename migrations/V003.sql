UPDATE groups SET name="" where name IS NULL;
ALTER TABLE groups MODIFY name VARCHAR(255) NOT NULL;
