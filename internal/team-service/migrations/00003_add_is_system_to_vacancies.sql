-- +goose Up
-- +goose StatementBegin

ALTER TABLE team.vacancies 
ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_vacancies_is_system ON team.vacancies(is_system);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS team.idx_vacancies_is_system;

ALTER TABLE team.vacancies 
DROP COLUMN IF EXISTS is_system;

-- +goose StatementEnd
