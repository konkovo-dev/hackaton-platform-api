-- ============================================
-- Test Data for Participation and Roles Service
-- ============================================

-- ============================================
-- Пользователи для тестирования staff management
-- ============================================

-- Alice: Owner хакатона
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'alice_owner', 'Alice', 'Owner', 'https://example.com/alice.jpg', 'UTC', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('a11ce000-0000-0000-0000-000000000001', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- Bob: будет приглашен как ORGANIZER (и уже имеет эту роль для тестирования RemoveHackathonRole)
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'bob_organizer', 'Bob', 'Organizer', 'https://example.com/bob.jpg', 'UTC', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('b0b00000-0000-0000-0000-000000000002', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- Charlie: будет приглашен как MENTOR
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'charlie_mentor', 'Charlie', 'Mentor', '', 'Europe/London', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('c4a41e00-0000-0000-0000-000000000003', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- Diana: будет приглашена как JUDGE
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'diana_judge', 'Diana', 'Judge', '', 'America/New_York', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('d1a4a000-0000-0000-0000-000000000004', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- Eve: простой пользователь без staff ролей
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES 
  ('e5e00000-0000-0000-0000-000000000005', 'eve_user', 'Eve', 'User', '', 'UTC', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES 
  ('e5e00000-0000-0000-0000-000000000005', 'public', 'public')
ON CONFLICT (user_id) DO NOTHING;

-- ============================================
-- Хакатон для тестирования
-- ============================================

INSERT INTO hackaton.hackathons (
  id, 
  name, 
  short_description, 
  description, 
  stage,
  online, 
  country, 
  city, 
  venue,
  registration_opens_at,
  registration_closes_at,
  starts_at,
  ends_at,
  judging_ends_at,
  allow_individual,
  allow_team,
  team_size_max,
  created_at, 
  updated_at,
  published_at
)
VALUES (
  '44444444-4444-4444-4444-444444444444',
  'AI Innovation Hackathon 2026',
  'Build the future with AI',
  'Join us for an exciting hackathon focused on AI and ML innovations. Teams will compete to create innovative solutions using cutting-edge AI technologies.',
  'published',
  false,
  'Russia',
  'Moscow',
  'Skolkovo Innovation Center',
  '2026-03-01T00:00:00Z',
  '2026-03-20T23:59:59Z',
  '2026-03-25T10:00:00Z',
  '2026-03-27T18:00:00Z',
  '2026-03-30T18:00:00Z',
  true,
  true,
  5,
  NOW(),
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- Alice = OWNER этого хакатона
-- ============================================

INSERT INTO participation_and_roles.staff_roles (hackathon_id, user_id, role, created_at)
VALUES 
  ('44444444-4444-4444-4444-444444444444', 'a11ce000-0000-0000-0000-000000000001', 'owner', NOW())
ON CONFLICT (hackathon_id, user_id, role) DO NOTHING;

-- ============================================
-- Bob уже является ORGANIZER (для тестирования RemoveHackathonRole)
-- ============================================

INSERT INTO participation_and_roles.staff_roles (hackathon_id, user_id, role, created_at)
VALUES 
  ('44444444-4444-4444-4444-444444444444', 'b0b00000-0000-0000-0000-000000000002', 'organizer', NOW())
ON CONFLICT (hackathon_id, user_id, role) DO NOTHING;

-- ============================================
-- Verification queries
-- ============================================

-- Проверить пользователей
SELECT 'Users:' AS section;
SELECT id, username, first_name, last_name FROM identity.users 
WHERE id IN (
  'a11ce000-0000-0000-0000-000000000001',
  'b0b00000-0000-0000-0000-000000000002',
  'c4a41e00-0000-0000-0000-000000000003',
  'd1a4a000-0000-0000-0000-000000000004',
  'e5e00000-0000-0000-0000-000000000005'
);

-- Проверить хакатон
SELECT 'Hackathon:' AS section;
SELECT id, name, stage, published_at IS NOT NULL AS is_published FROM hackaton.hackathons 
WHERE id = '44444444-4444-4444-4444-444444444444';

-- Проверить staff роли
SELECT 'Staff Roles:' AS section;
SELECT hackathon_id, user_id, role, created_at FROM participation_and_roles.staff_roles
WHERE hackathon_id = '44444444-4444-4444-4444-444444444444'
ORDER BY role, created_at;

-- ============================================
-- Expected Results Summary
-- ============================================
-- Users: 5 (alice_owner, bob_organizer, charlie_mentor, diana_judge, eve_user)
-- Hackathon: 1 (AI Innovation Hackathon 2026, stage=published)
-- Staff Roles: 2 (Alice=owner, Bob=organizer)
-- ============================================

