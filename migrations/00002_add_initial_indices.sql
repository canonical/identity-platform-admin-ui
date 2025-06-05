--  Copyright 2025 Canonical Ltd.
--  SPDX-License-Identifier: AGPL-3.0

-- +goose Up
CREATE INDEX idx_group_name_owner ON "group" (name, owner);
CREATE INDEX idx_role_name_owner ON role (name, owner);
CREATE INDEX idx_application_name_owner ON application (name, owner);
CREATE UNIQUE INDEX idx_application_client_id ON application (client_id);

-- +goose Down
DROP INDEX idx_group_name_owner;
DROP INDEX idx_role_name_owner;
DROP INDEX idx_application_name_owner;
DROP INDEX idx_application_client_id;
