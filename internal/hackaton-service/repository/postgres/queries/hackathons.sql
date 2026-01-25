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
    state,
    published_at,
    result_published_at,
    task,
    result,
    team_size_max,
    allow_individual,
    allow_team,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
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
    state,
    published_at,
    result_published_at,
    task,
    result,
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
    state = $17,
    published_at = $18,
    result_published_at = $19,
    task = $20,
    result = $21,
    team_size_max = $22,
    allow_individual = $23,
    allow_team = $24,
    updated_at = $25
WHERE id = $1;

-- name: ListHackathons :many
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
    state,
    published_at,
    result_published_at,
    task,
    result,
    team_size_max,
    allow_individual,
    allow_team,
    created_at,
    updated_at
FROM hackathon.hackathons
WHERE state = 'published'
ORDER BY 
    CASE 
        WHEN stage IN ('registration', 'prestart', 'running') THEN 1
        WHEN stage = 'upcoming' THEN 2
        WHEN stage = 'judging' THEN 3
        WHEN stage = 'finished' THEN 4
        ELSE 5
    END,
    created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPublishedHackathons :one
SELECT COUNT(*) FROM hackathon.hackathons WHERE state = 'published';

