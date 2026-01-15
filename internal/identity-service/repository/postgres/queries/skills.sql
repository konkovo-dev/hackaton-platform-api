-- name: SkillCatalogListByIDs :many
SELECT id, name
FROM identity.skill_catalog
WHERE id = ANY($1::uuid[]);

-- name: SkillCatalogGetByUserID :many
SELECT cs.id, cs.name
FROM identity.user_catalog_skills ucs
JOIN identity.skill_catalog cs ON ucs.catalog_skill_id = cs.id
WHERE ucs.user_id = $1;

-- name: SkillCustomGetByUserID :many
SELECT id, user_id, name
FROM identity.user_custom_skills
WHERE user_id = $1;

-- name: SkillCatalogDeleteByUserID :exec
DELETE FROM identity.user_catalog_skills
WHERE user_id = $1;

-- name: SkillCustomDeleteByUserID :exec
DELETE FROM identity.user_custom_skills
WHERE user_id = $1;

-- name: SkillCatalogCreate :exec
INSERT INTO identity.user_catalog_skills (user_id, catalog_skill_id)
VALUES ($1, $2);

-- name: SkillCustomCreate :exec
INSERT INTO identity.user_custom_skills (id, user_id, name)
VALUES ($1, $2, $3);

-- name: SkillCatalogGetByUserIDs :many
SELECT ucs.user_id, cs.id, cs.name
FROM identity.user_catalog_skills ucs
JOIN identity.skill_catalog cs ON ucs.catalog_skill_id = cs.id
WHERE ucs.user_id = ANY($1::uuid[]);

-- name: SkillCustomGetByUserIDs :many
SELECT id, user_id, name
FROM identity.user_custom_skills
WHERE user_id = ANY($1::uuid[]);

-- name: SkillCatalogListAll :many
SELECT id, name
FROM identity.skill_catalog
ORDER BY name ASC;

