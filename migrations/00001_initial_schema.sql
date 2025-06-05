--  Copyright 2025 Canonical Ltd.
--  SPDX-License-Identifier: AGPL-3.0

-- +goose Up
-- +goose StatementBegin

CREATE SEQUENCE identity_id_seq;
CREATE SEQUENCE application_id_seq;
CREATE SEQUENCE group_id_seq;
CREATE SEQUENCE role_id_seq;
CREATE SEQUENCE policy_id_seq;
CREATE SEQUENCE permission_id_seq;

CREATE TABLE identity
(
    id          BIGINT PRIMARY KEY DEFAULT nextval('identity_id_seq'),
    identity_id TEXT NOT NULL UNIQUE,
    username    TEXT NOT NULL UNIQUE
);

CREATE TABLE application
(
    id        BIGINT PRIMARY KEY DEFAULT nextval('application_id_seq'),
    name      TEXT NOT NULL,
    client_id TEXT NOT NULL,
    owner     TEXT NOT NULL,
    owner_id  BIGINT REFERENCES identity (id) ON DELETE CASCADE -- not used now, will become NOT NULL during the epic
);

CREATE TABLE "group"
(
    id       BIGINT PRIMARY KEY DEFAULT nextval('group_id_seq'),
    name     TEXT NOT NULL,
    owner    TEXT NOT NULL,
    owner_id BIGINT REFERENCES identity (id) ON DELETE CASCADE -- not used now, will become NOT NULL during the epic
);

CREATE TABLE role
(
    id       BIGINT PRIMARY KEY DEFAULT nextval('role_id_seq'),
    name     TEXT NOT NULL,
    owner    TEXT NOT NULL,
    owner_id BIGINT REFERENCES identity (id) ON DELETE CASCADE -- not used now, will become NOT NULL during the epic
);

CREATE TABLE policy
(
    id       BIGINT PRIMARY KEY DEFAULT nextval('policy_id_seq'),
    version  TEXT   NOT NULL,
    name     TEXT   NOT NULL,
    owner    TEXT   NOT NULL,
    owner_id BIGINT REFERENCES identity (id) ON DELETE CASCADE, -- not used now, will become NOT NULL during the epic
    role_id  BIGINT NOT NULL REFERENCES role (id) ON DELETE CASCADE
);

CREATE TABLE permission
(
    id        BIGINT PRIMARY KEY DEFAULT nextval('permission_id_seq'),
    resource  TEXT   NOT NULL,
    action    TEXT   NOT NULL,
    policy_id BIGINT NOT NULL REFERENCES policy (id) ON DELETE CASCADE
);


-- group ↔ identity
CREATE TABLE group_identity
(
    group_id    BIGINT NOT NULL REFERENCES "group" (id) ON DELETE CASCADE,
    identity_id BIGINT NOT NULL REFERENCES identity (id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, identity_id)
);

-- group ↔ role
CREATE TABLE group_role
(
    group_id BIGINT NOT NULL REFERENCES "group" (id) ON DELETE CASCADE,
    role_id  BIGINT NOT NULL REFERENCES role (id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, role_id)
);

-- identity ↔ role
CREATE TABLE identity_role
(
    identity_id BIGINT NOT NULL REFERENCES identity (id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES role (id) ON DELETE CASCADE,
    PRIMARY KEY (identity_id, role_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "role_identity";
DROP TABLE IF EXISTS "group_role";
DROP TABLE IF EXISTS "group_identity";
DROP TABLE IF EXISTS "permission";
DROP TABLE IF EXISTS "policy";
DROP TABLE IF EXISTS "role";
DROP TABLE IF EXISTS "group";
DROP TABLE IF EXISTS "identity";
DROP TABLE IF EXISTS "application";

DROP SEQUENCE IF EXISTS "application_id_seq";
DROP SEQUENCE IF EXISTS "identity_id_seq";
DROP SEQUENCE IF EXISTS "group_id_seq";
DROP SEQUENCE IF EXISTS "role_id_seq";
DROP SEQUENCE IF EXISTS "policy_id_seq";
DROP SEQUENCE IF EXISTS "permission_id_seq";

-- +goose StatementEnd
