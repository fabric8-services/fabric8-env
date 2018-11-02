CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE environments (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    name text,
    type text,
    space_id uuid NOT NULL,
    namespace_name text,
    cluster_url text
);

CREATE INDEX environments_space_id_idx ON environments USING BTREE (space_id);
