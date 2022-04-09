BEGIN;

CREATE TABLE IF NOT EXISTS tasks (
      id varchar(64) primary key not null,
      created_at timestamp without time zone default now() not null,
      updated_at timestamp without time zone default now() not null,

      from_page varchar(512) not null,
      to_page varchar(512) not null,

      status integer default 0 not null,
      result jsonb
);

CREATE INDEX IF NOT EXISTS tasks_created_at_idx ON tasks USING btree(created_at);
CREATE INDEX IF NOT EXISTS tasks_updated_at_idx ON tasks USING btree(updated_at);
CREATE INDEX IF NOT EXISTS tasks_status_idx ON tasks USING btree(status);

COMMIT;