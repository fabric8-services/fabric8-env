UPDATE environments SET name = 'dummy-env' WHERE name IS NULL OR name = '';
ALTER TABLE environments ALTER COLUMN name SET NOT NULL;

UPDATE environments SET type = 'stage' WHERE type IS NULL OR type = '';
ALTER TABLE environments ALTER COLUMN type SET NOT NULL;

UPDATE environments SET cluster_url = 'https://api.starter-us-east-2.openshift.com' WHERE cluster_url IS NULL OR cluster_url = '';
ALTER TABLE environments ALTER COLUMN cluster_url SET NOT NULL;
