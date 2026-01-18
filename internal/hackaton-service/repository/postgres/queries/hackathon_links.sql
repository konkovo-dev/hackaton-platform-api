-- name: CreateHackathonLink :exec
INSERT INTO hackathon.hackathon_links (id, hackathon_id, title, url)
VALUES ($1, $2, $3, $4);

-- name: GetHackathonLinks :many
SELECT id, hackathon_id, title, url
FROM hackathon.hackathon_links
WHERE hackathon_id = $1
ORDER BY id;

-- name: DeleteHackathonLinks :exec
DELETE FROM hackathon.hackathon_links
WHERE hackathon_id = $1;

