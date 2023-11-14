-- +goose Up
-- +goose StatementBegin
CREATE TYPE list_type AS ENUM ('White','Black');
CREATE TABLE IF NOT EXISTS bw_lists (
    id SERIAL PRIMARY KEY,
    network inet NOT NULL,
    type list_type NOT NULL DEFAULT 'White',
    created_at TIMESTAMP NOT NULL DEFAULT now()
    );
CREATE INDEX IF NOT EXISTS net_idx ON bw_lists USING gist (network inet_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX net_idx;
DROP TABLE bw_lists;
DROP TYPE list_type;
-- +goose StatementEnd

-- insert into bw_lists (network, type)
-- select inet(network), list_type(type) from (
--      select
--      '192.168.4.0/24' as network,
--      'White' as type
-- ) t
-- where not exists (
--      select 1 from bw_lists bw
--      where bw.network && inet(t.network)
-- );

-- select network, type from bw_lists where network && '192.168.1.0/23';
