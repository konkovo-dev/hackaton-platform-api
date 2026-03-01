-- name: UpsertVacancy :exec
INSERT INTO matchmaking.vacancies (
    vacancy_id,
    team_id,
    hackathon_id,
    description,
    desired_role_ids,
    desired_skill_ids,
    slots_open,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (vacancy_id) DO UPDATE SET
    team_id = EXCLUDED.team_id,
    hackathon_id = EXCLUDED.hackathon_id,
    description = EXCLUDED.description,
    desired_role_ids = EXCLUDED.desired_role_ids,
    desired_skill_ids = EXCLUDED.desired_skill_ids,
    slots_open = EXCLUDED.slots_open,
    updated_at = EXCLUDED.updated_at;

-- name: GetVacancyByID :one
SELECT vacancy_id, team_id, hackathon_id, description, desired_role_ids, desired_skill_ids, slots_open, updated_at
FROM matchmaking.vacancies
WHERE vacancy_id = $1;

-- name: GetVacanciesByTeamID :many
SELECT vacancy_id, team_id, hackathon_id, description, desired_role_ids, desired_skill_ids, slots_open, updated_at
FROM matchmaking.vacancies
WHERE team_id = $1;

-- name: ListVacanciesByHackathon :many
SELECT vacancy_id, team_id, hackathon_id, description, desired_role_ids, desired_skill_ids, slots_open, updated_at
FROM matchmaking.vacancies
WHERE hackathon_id = $1
  AND slots_open > 0;

-- name: DeleteVacancy :exec
DELETE FROM matchmaking.vacancies
WHERE vacancy_id = $1;

-- name: DeleteVacanciesByTeamID :exec
DELETE FROM matchmaking.vacancies
WHERE team_id = $1;
