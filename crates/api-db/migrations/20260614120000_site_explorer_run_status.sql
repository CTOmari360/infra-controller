-- Store the operator-facing result of the most recent Site Explorer iteration.
-- Endpoint rows already hold per-endpoint exploration errors; this singleton
-- captures whole-run failures such as missing global credentials or database
-- setup issues that otherwise only appear in nico-api logs.
CREATE TABLE site_explorer_run_status (
    id                              smallint    PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    started_at                      timestamptz NOT NULL,
    finished_at                     timestamptz NOT NULL,
    success                         boolean     NOT NULL,
    error                           text,
    endpoint_explorations           bigint      NOT NULL,
    endpoint_explorations_success   bigint      NOT NULL,
    endpoint_explorations_failed    bigint      NOT NULL
);
