-- name: CreateHackathon :exec
INSERT INTO hackathon.hackathons (
    id,
    name,
    short_description,
    description,
    location_online,
    location_city,
    location_country,
    location_venue,
    starts_at,
    ends_at,
    registration_opens_at,
    registration_closes_at,
    submissions_opens_at,
    submissions_closes_at,
    judging_ends_at,
    stage,
    team_size_max,
    allow_individual,
    allow_team,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
);

-- name: GetHackathonByID :one
SELECT 
    id,
    name,
    short_description,
    description,
    location_online,
    location_city,
    location_country,
    location_venue,
    starts_at,
    ends_at,
    registration_opens_at,
    registration_closes_at,
    submissions_opens_at,
    submissions_closes_at,
    judging_ends_at,
    stage,
    team_size_max,
    allow_individual,
    allow_team,
    created_at,
    updated_at
FROM hackathon.hackathons
WHERE id = $1;

-- name: UpdateHackathon :exec
UPDATE hackathon.hackathons
SET 
    name = $2,
    short_description = $3,
    description = $4,
    location_online = $5,
    location_city = $6,
    location_country = $7,
    location_venue = $8,
    starts_at = $9,
    ends_at = $10,
    registration_opens_at = $11,
    registration_closes_at = $12,
    submissions_opens_at = $13,
    submissions_closes_at = $14,
    judging_ends_at = $15,
    stage = $16,
    team_size_max = $17,
    allow_individual = $18,
    allow_team = $19,
    updated_at = $20
WHERE id = $1;

