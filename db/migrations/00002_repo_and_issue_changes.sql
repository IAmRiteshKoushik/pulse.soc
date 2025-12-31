-- +goose Up

-- +goose StatementBegin
ALTER TABLE repository
  ADD COLUMN on_display BOOLEAN NOT NULL DEFAULT false,
  DROP COLUMN installation_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE issues
  ADD COLUMN bounty_promised INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose StatementBegin
-- Changing an existing column requires specific SET actions
ALTER TABLE issues
  ALTER COLUMN difficulty TYPE TEXT,
  ALTER COLUMN difficulty SET DEFAULT 'EASY',
  ALTER COLUMN difficulty SET NOT NULL;
-- +goose StatementEnd
