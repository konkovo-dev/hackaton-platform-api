-- ============================================
-- Тестовые данные для Participation Service
-- ============================================
-- Использование: 
-- docker-compose -f deployments/docker-compose.yml exec -T postgres \
--   psql -U hackathon -d hackathon < docs/participation-and-roles/test-data-participation.sql
--
-- ВАЖНО: Пользователи создаются без credentials (без паролей).
-- Для тестирования нужно сначала зарегистрировать пользователей через API
-- или использовать автоматические тестовые скрипты, которые создают пользователей.
-- ============================================

BEGIN;

-- ============================================
-- 1. Пользователи для тестирования
-- ============================================

-- Alice: Owner хакатона (staff member)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'alice_staff', 'Alice', 'Staff', 'https://example.com/alice.jpg', 'UTC', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- Bob: Participant (individual active)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'bob_participant', 'Bob', 'Participant', 'https://example.com/bob.jpg', 'UTC', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- Charlie: Participant (looking for team)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'charlie_participant', 'Charlie', 'Seeker', '', 'Europe/London', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- Diana: New participant
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'diana_new', 'Diana', 'New', '', 'America/New_York', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- ============================================
-- 2. Хакатон для тестирования
-- ============================================

INSERT INTO hackathon.hackathons (
  id, 
  name, 
  short_description, 
  description, 
  stage,
  state,
  location_online, 
  location_country, 
  location_city, 
  location_venue,
  registration_opens_at,
  registration_closes_at,
  starts_at,
  ends_at,
  submissions_opens_at,
  submissions_closes_at,
  judging_ends_at,
  allow_individual,
  allow_team,
  team_size_max,
  task,
  result,
  created_at, 
  updated_at,
  published_at
)
VALUES (
  '55555555-5555-5555-5555-555555555555',
  'Spring Hackathon 2026',
  'Code, Create, Innovate',
  'Join us for an amazing hackathon experience with team building and innovation.',
  'registration',
  'active',
  false,
  'Russia',
  'Moscow',
  'Skolkovo Innovation Center',
  '2026-03-01T00:00:00Z',
  '2026-03-20T23:59:59Z',
  '2026-03-25T10:00:00Z',
  '2026-03-27T18:00:00Z',
  '2026-03-25T10:00:00Z',
  '2026-03-27T18:00:00Z',
  '2026-03-30T18:00:00Z',
  true,
  true,
  5,
  '',
  '',
  NOW(),
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- 3. Alice = OWNER этого хакатона
-- ============================================

INSERT INTO participation_and_roles.staff_roles (hackathon_id, user_id, role, created_at)
VALUES 
  ('55555555-5555-5555-5555-555555555555', 'a11ce000-0000-0000-0000-000000000001', 'owner', NOW())
ON CONFLICT (hackathon_id, user_id, role) DO NOTHING;

-- ============================================
-- 4. Team Roles (если еще не созданы миграцией)
-- ============================================

INSERT INTO participation_and_roles.team_role_catalog (id, name) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Any'),
  ('00000000-0000-0000-0000-000000000002', 'Backend'),
  ('00000000-0000-0000-0000-000000000003', 'Frontend'),
  ('00000000-0000-0000-0000-000000000004', 'Fullstack'),
  ('00000000-0000-0000-0000-000000000005', 'Mobile'),
  ('00000000-0000-0000-0000-000000000006', 'Designer'),
  ('00000000-0000-0000-0000-000000000007', 'Product Manager'),
  ('00000000-0000-0000-0000-000000000008', 'Data Scientist'),
  ('00000000-0000-0000-0000-000000000009', 'DevOps'),
  ('00000000-0000-0000-0000-00000000000a', 'QA')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- 5. Existing participations для Bob и Charlie
-- ============================================

-- Bob: Individual active participant
INSERT INTO participation_and_roles.participations (
  hackathon_id, user_id, status, team_id, motivation_text, registered_at, created_at, updated_at
)
VALUES 
  ('55555555-5555-5555-5555-555555555555', 'b0b00000-0000-0000-0000-000000000002', 'individual', NULL, 'I love to code and build!', NOW(), NOW(), NOW())
ON CONFLICT (hackathon_id, user_id) DO NOTHING;

-- Charlie: Looking for team participant
INSERT INTO participation_and_roles.participations (
  hackathon_id, user_id, status, team_id, motivation_text, registered_at, created_at, updated_at
)
VALUES 
  ('55555555-5555-5555-5555-555555555555', 'c4a41e00-0000-0000-0000-000000000003', 'looking_for_team', NULL, 'Want to find a great team to collaborate with', NOW(), NOW(), NOW())
ON CONFLICT (hackathon_id, user_id) DO NOTHING;

-- ============================================
-- 6. Wished roles для Charlie
-- ============================================

INSERT INTO participation_and_roles.participation_wished_roles (hackathon_id, user_id, team_role_id)
VALUES 
  ('55555555-5555-5555-5555-555555555555', 'c4a41e00-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002'),
  ('55555555-5555-5555-5555-555555555555', 'c4a41e00-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000004')
ON CONFLICT (hackathon_id, user_id, team_role_id) DO NOTHING;

COMMIT;

-- ============================================
-- Проверка данных
-- ============================================

\echo '============================================'
\echo 'Verification:'
\echo '============================================'

\echo ''
\echo 'Users:'
SELECT id, username, first_name, last_name FROM identity.users 
WHERE id IN (
  'a11ce000-0000-0000-0000-000000000001',
  'b0b00000-0000-0000-0000-000000000002',
  'c4a41e00-0000-0000-0000-000000000003',
  'd1a4a000-0000-0000-0000-000000000004'
);

\echo ''
\echo 'Hackathon:'
SELECT id, name, stage, allow_individual, allow_team FROM hackathon.hackathons 
WHERE id = '55555555-5555-5555-5555-555555555555';

\echo ''
\echo 'Staff Roles:'
SELECT hackathon_id, user_id, role FROM participation_and_roles.staff_roles
WHERE hackathon_id = '55555555-5555-5555-5555-555555555555';

\echo ''
\echo 'Participations:'
SELECT hackathon_id, user_id, status, motivation_text FROM participation_and_roles.participations
WHERE hackathon_id = '55555555-5555-5555-5555-555555555555';

\echo ''
\echo 'Team Roles Catalog:'
SELECT id, name FROM participation_and_roles.team_role_catalog LIMIT 5;

\echo ''
\echo 'Charlie Wished Roles:'
SELECT pwr.hackathon_id, pwr.user_id, tr.name as role_name
FROM participation_and_roles.participation_wished_roles pwr
JOIN participation_and_roles.team_role_catalog tr ON pwr.team_role_id = tr.id
WHERE pwr.hackathon_id = '55555555-5555-5555-5555-555555555555' 
  AND pwr.user_id = 'c4a41e00-0000-0000-0000-000000000003';

\echo ''
\echo '============================================'
\echo 'Test data loaded successfully!'
\echo 'Expected:'
\echo '  - 4 users'
\echo '  - 1 hackathon (registration stage)'
\echo '  - 1 staff role (Alice = OWNER)'
\echo '  - 2 participations (Bob, Charlie)'
\echo '  - 10 team roles'
\echo '  - 2 wished roles for Charlie'
\echo '============================================'
