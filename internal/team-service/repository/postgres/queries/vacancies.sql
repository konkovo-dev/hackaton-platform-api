-- name: GetVacanciesByTeamIDs :many
SELECT 
    id,
    team_id,
    description,
    desired_role_ids,
    desired_skill_ids,
    slots_total,
    slots_open,
    is_system,
    created_at,
    updated_at
FROM team.vacancies
WHERE team_id = ANY($1::uuid[])
ORDER BY created_at ASC;

-- name: GetVacanciesByTeamID :many
SELECT 
    id,
    team_id,
    description,
    desired_role_ids,
    desired_skill_ids,
    slots_total,
    slots_open,
    is_system,
    created_at,
    updated_at
FROM team.vacancies
WHERE team_id = $1
ORDER BY created_at ASC;

-- name: GetVacancyByID :one
SELECT 
    id,
    team_id,
    description,
    desired_role_ids,
    desired_skill_ids,
    slots_total,
    slots_open,
    is_system,
    created_at,
    updated_at
FROM team.vacancies
WHERE id = $1;

-- name: CreateVacancy :one
INSERT INTO team.vacancies (
    id,
    team_id,
    description,
    desired_role_ids,
    desired_skill_ids,
    slots_total,
    slots_open,
    is_system,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING
    id,
    team_id,
    description,
    desired_role_ids,
    desired_skill_ids,
    slots_total,
    slots_open,
    is_system,
    created_at,
    updated_at;

-- name: UpdateVacancy :one
UPDATE team.vacancies
SET
    description = $2,
    desired_role_ids = $3,
    desired_skill_ids = $4,
    slots_total = $5,
    slots_open = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    team_id,
    description,
    desired_role_ids,
    desired_skill_ids,
    slots_total,
    slots_open,
    is_system,
    created_at,
    updated_at;

-- name: CountOccupiedSlots :one
SELECT COUNT(*)
FROM team.memberships
WHERE assigned_vacancy_id = $1::uuid;

-- name: CountTotalOpenSlots :one
SELECT COALESCE(SUM(slots_open), 0)::bigint
FROM team.vacancies
WHERE team_id = $1;

-- name: DecrementSlotsOpen :exec
UPDATE team.vacancies
SET
    slots_open = slots_open - 1,
    updated_at = NOW()
WHERE id = $1;

-- name: IncrementSlotsOpen :exec
UPDATE team.vacancies
SET
    slots_open = LEAST(slots_open + 1, slots_total),
    updated_at = NOW()
WHERE id = $1;
