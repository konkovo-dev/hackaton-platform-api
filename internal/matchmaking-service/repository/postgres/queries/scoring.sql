-- name: ScoreTeamsForUser :many
WITH user_data AS (
    SELECT
        p.wished_role_ids,
        p.motivation_text,
        p.motivation_tsv,
        u.catalog_skill_ids AS user_catalog_skills,
        u.custom_skill_names AS user_custom_skills
    FROM matchmaking.participations p
    JOIN matchmaking.users u ON u.user_id = p.user_id
    WHERE p.user_id = $1 AND p.hackathon_id = $2
),
team_vacancies AS (
    SELECT
        t.team_id,
        t.name AS team_name,
        t.description AS team_description,
        t.description_tsv AS team_description_tsv,
        v.vacancy_id,
        v.desired_role_ids,
        v.desired_skill_ids,
        v.description AS vacancy_description,
        v.description_tsv AS vacancy_description_tsv
    FROM matchmaking.teams t
    JOIN matchmaking.vacancies v ON v.team_id = t.team_id
    WHERE t.hackathon_id = $2
        AND t.is_joinable = true
        AND v.slots_open > 0
),
scored AS (
    SELECT
        tv.team_id,
        tv.vacancy_id,
        tv.team_name,
        CASE
            WHEN cardinality(tv.desired_skill_ids) = 0 THEN 1.0
            ELSE (
                SELECT COUNT(*)::float / cardinality(tv.desired_skill_ids)
                FROM unnest(tv.desired_skill_ids) AS ds
                WHERE ds = ANY(ud.user_catalog_skills)
            )
        END AS skills_score,
        CASE
            WHEN cardinality(tv.desired_role_ids) = 0 THEN 1.0
            ELSE (
                SELECT COUNT(*)::float / cardinality(tv.desired_role_ids)
                FROM unnest(tv.desired_role_ids) AS dr
                WHERE dr = ANY(ud.wished_role_ids)
            )
        END AS roles_score,
        COALESCE(
            ts_rank(
                tv.team_description_tsv || tv.vacancy_description_tsv,
                to_tsquery('english', plainto_tsquery('english', ud.motivation_text)::text) ||
                to_tsquery('russian', plainto_tsquery('russian', ud.motivation_text)::text)
            ),
            0.0
        ) AS text_score
    FROM team_vacancies tv
    CROSS JOIN user_data ud
)
SELECT
    team_id,
    vacancy_id,
    (skills_score * 0.63 + roles_score * 0.27 + text_score * 0.10) AS total_score,
    skills_score,
    roles_score,
    text_score
FROM scored
WHERE (skills_score * 0.63 + roles_score * 0.27 + text_score * 0.10) > 0
ORDER BY total_score DESC
LIMIT $3;

-- name: ScoreCandidatesForVacancy :many
WITH vacancy_data AS (
    SELECT
        v.desired_role_ids,
        v.desired_skill_ids,
        v.description AS vacancy_description,
        v.description_tsv AS vacancy_description_tsv,
        t.description AS team_description,
        t.description_tsv AS team_description_tsv
    FROM matchmaking.vacancies v
    JOIN matchmaking.teams t ON t.team_id = v.team_id
    WHERE v.vacancy_id = $1
),
candidates AS (
    SELECT
        p.user_id,
        p.wished_role_ids,
        p.motivation_text,
        p.motivation_tsv,
        u.catalog_skill_ids AS user_catalog_skills,
        u.custom_skill_names AS user_custom_skills
    FROM matchmaking.participations p
    JOIN matchmaking.users u ON u.user_id = p.user_id
    WHERE p.hackathon_id = $2
        AND p.status = 'looking_for_team'
        AND p.team_id IS NULL
),
scored AS (
    SELECT
        c.user_id,
        CASE
            WHEN cardinality(vd.desired_skill_ids) = 0 THEN 1.0
            ELSE (
                SELECT COUNT(*)::float / cardinality(vd.desired_skill_ids)
                FROM unnest(vd.desired_skill_ids) AS ds
                WHERE ds = ANY(c.user_catalog_skills)
            )
        END AS skills_score,
        CASE
            WHEN cardinality(vd.desired_role_ids) = 0 THEN 1.0
            ELSE (
                SELECT COUNT(*)::float / cardinality(vd.desired_role_ids)
                FROM unnest(vd.desired_role_ids) AS dr
                WHERE dr = ANY(c.wished_role_ids)
            )
        END AS roles_score,
        COALESCE(
            ts_rank(
                vd.team_description_tsv || vd.vacancy_description_tsv,
                to_tsquery('english', plainto_tsquery('english', c.motivation_text)::text) ||
                to_tsquery('russian', plainto_tsquery('russian', c.motivation_text)::text)
            ),
            0.0
        ) AS text_score
    FROM candidates c
    CROSS JOIN vacancy_data vd
)
SELECT
    user_id,
    (skills_score * 0.63 + roles_score * 0.27 + text_score * 0.10) AS total_score,
    skills_score,
    roles_score,
    text_score
FROM scored
WHERE (skills_score * 0.63 + roles_score * 0.27 + text_score * 0.10) > 0
ORDER BY total_score DESC
LIMIT $3;
