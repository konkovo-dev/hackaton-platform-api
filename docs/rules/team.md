# Hackathon Policy Spec — Team domain (v1.2, updated)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Role/Participation service)
- `roles[]`: `{OWNER, ORGANIZER, MENTOR, JURY}`
- `particip.status ∈ {PART_NONE, PART_INDIVIDUAL, PART_LOOKING_FOR_TEAM, PART_TEAM_MEMBER, PART_TEAM_CAPTAIN}`
- `particip.team_id`: `uuid | null`

Derived:
- `is_in_team := particip.status in {PART_TEAM_MEMBER, PART_TEAM_CAPTAIN}`
- `is_participant := particip.status != PART_NONE`
- `can_seek_team := particip.status in {PART_LOOKING_FOR_TEAM, PART_INDIVIDUAL}`

### 0.2 Hackathon context (Hackathon service)
- `published_at`: `timestamp | null`
- `stage ∈ {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`
- `policy.allow_team` (bool)
- `policy.team_size_max` (int | null)

### 0.3 Team context (Team service)
`Team`:
- `team_id`
- `hackathon_id`
- `name`
- `description`
- `is_joinable` (bool)

`Vacancy`:
- `vacancy_id`
- `team_id`
- `description`
- `desired_roles[]`
- `desired_skills[]`
- `slots_total` (int)
- `slots_open` (int)

`Membership`:
- `team_id`
- `user_id`
- `is_captain` (bool)
- `assigned_vacancy_id` (uuid | null)

`TeamInvite`:
- `invite_id`
- `hackathon_id`
- `team_id`
- `vacancy_id` (uuid)                 // always set
- `target_user_id`
- `status ∈ {PENDING, ACCEPTED, DECLINED, CANCELED, EXPIRED}`
- `message` (string | null)
- `created_at`
- `expires_at` (timestamp | null)

`JoinRequest`:
- `request_id`
- `hackathon_id`
- `team_id`
- `vacancy_id` (uuid)                 // always set
- `requester_user_id`
- `status ∈ {PENDING, ACCEPTED, DECLINED, CANCELED, EXPIRED}`
- `message` (string | null)
- `created_at`
- `expires_at` (timestamp | null)

---

## 1) Инварианты домена

### 1.1 Взаимоисключение staff и participation
В одном хакатоне запрещено одновременно:
- иметь любую роль из `{OWNER, ORGANIZER, MENTOR, JURY}`
- и иметь участие `particip.status != PART_NONE`

### 1.2 Один пользователь может состоять только в одной команде хакатона
- `particip.team_id` может указывать только на одну команду в рамках одного `hackathon_id`
- `particip.status in {PART_TEAM_MEMBER, PART_TEAM_CAPTAIN} -> particip.team_id != null`

### 1.3 Ограничение на размер команды
Если `policy.team_size_max != null`, то всегда должно выполняться:
- `members_count(team) <= policy.team_size_max`
- `members_count(team) + total_open_slots(team) <= policy.team_size_max`

### 1.4 Вакансии и слоты
- `0 <= slots_open <= slots_total`
- при принятии в команду слот закрывается, если у membership задан `assigned_vacancy_id`
- при выходе или исключении участника слот по `membership.assigned_vacancy_id` переоткрывается

### 1.5 Запрет на уменьшение слотов ниже занятых
Для любой вакансии:
- `slots_total_new >= slots_occupied_old`
где `slots_occupied_old = slots_total_old - slots_open_old`

### 1.6 Инвайты и заявки всегда привязаны к вакансии
- `TeamInvite.vacancy_id != null`
- `JoinRequest.vacancy_id != null`
- `vacancy.team_id == team.team_id`

---

## 2) Предикаты

### 2.1 Роли и участие
- `role in {…}`: actor имеет хотя бы одну роль из множества
- `is_staff := role in {OWNER, ORGANIZER, MENTOR, JURY}`
- `is_participant := particip.status != PART_NONE`
- `is_in_team := particip.status in {PART_TEAM_MEMBER, PART_TEAM_CAPTAIN}`
- `can_seek_team := particip.status in {PART_LOOKING_FOR_TEAM, PART_INDIVIDUAL}`

### 2.2 Окна стадий
- `TeamReadWindow`: `stage in {REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`
- `TeamWriteWindow`: `stage == REGISTRATION`
- `TeamLeaveWindow`: `stage in {REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`

### 2.3 Права в команде
- `is_team_member(team)`: существует `Membership(team_id, actor.user_id)`
- `is_team_captain(team)`: существует `Membership(team_id, actor.user_id, is_captain=true)`

---

## 3) Базовые условия доступа

### 3.1 CanJoinTeam(actor)
`CanJoinTeam(actor)` выполняется, если:
- `is_staff` не выполняется
- `particip.status NOT IN {PART_TEAM_MEMBER, PART_TEAM_CAPTAIN}`

---

## 4) Политика чтения

### 4.1 Team.ReadCatalog
`Team.ReadCatalog @ auth && TeamReadWindow && (role in {OWNER, ORGANIZER, MENTOR} || is_participant)`

Описание:
- читать список команд и вакансий можно начиная с `REGISTRATION`
- доступно staff (`OWNER/ORGANIZER/MENTOR`) и пользователям, которые участвуют в хакатоне (`particip.status != PART_NONE`)

### 4.2 Team.ReadById
`Team.ReadById @ auth && TeamReadWindow && (role in {OWNER, ORGANIZER, MENTOR} || is_participant)`

Описание:
- читать команду и вакансии можно начиная с `REGISTRATION`
- доступно staff (`OWNER/ORGANIZER/MENTOR`) и пользователям, которые участвуют в хакатоне (`particip.status != PART_NONE`)

### 4.3 Team.ReadRoster
`Team.ReadRoster @ auth && TeamReadWindow && (role in {OWNER, ORGANIZER, MENTOR} || is_participant)`

Описание:
- состав команд виден `OWNER/ORGANIZER/MENTOR`
- состав команд виден любому пользователю с `particip.status != PART_NONE`

### 4.4 Team.ReadMyInbox
`Team.ReadMyInbox @ auth && TeamReadWindow && is_participant`

Описание:
- пользователь читает свои командные инвайты и свои заявки
- доступно только участникам хакатона (не staff)

### 4.5 Team.ReadTeamInbox
`Team.ReadTeamInbox @ auth && TeamReadWindow && is_team_captain(team)`

Описание:
- капитан читает входящие заявки в свою команду и исходящие приглашения

---

## 5) Политика записи

Общее условие для любых действий записи (кроме выхода):
- `auth && TeamWriteWindow && policy.allow_team == true`

### 5.1 Team.Create
`Team.Create @ auth && TeamWriteWindow && policy.allow_team == true && CanJoinTeam(actor)`

Описание:
- пользователь создаёт команду в хакатоне
- пользователь становится капитаном
- если у пользователя нет участия, оно создаётся
- если у пользователя участие `PART_LOOKING_FOR_TEAM` или `PART_INDIVIDUAL`, оно конвертируется в `PART_TEAM_CAPTAIN`

Строгие требования:
- `team.name` задан
- `team.name` уникален в пределах `hackathon_id`

### 5.2 Team.Update
`Team.Update @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team)`

Описание:
- капитан обновляет `team.name`, `team.description`, `team.is_joinable`

Ограничения:
- `team.name` уникален в пределах `hackathon_id`
- инвариант `team_size_max` сохраняется

### 5.3 Vacancy.Upsert
`Vacancy.Upsert @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team)`

Описание:
- капитан создаёт или обновляет вакансии

Ограничения:
- `slots_total >= 0`
- запрещено установить `slots_total_new < slots_occupied_old` для вакансии
- инвариант `team_size_max` сохраняется

### 5.4 TeamInvite.Create
`TeamInvite.Create @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team) && vacancy_id задан && slots_open(vacancy) > 0`

Описание:
- капитан приглашает пользователя в команду
- приглашение всегда привязано к вакансии (включая "Any/Generalist" vacancy)

Ограничения:
- `slots_open(vacancy) > 0`
- target не является staff в этом хакатоне
- target не имеет `particip.status in {PART_TEAM_MEMBER, PART_TEAM_CAPTAIN}` в этом хакатоне

### 5.5 TeamInvite.Cancel
`TeamInvite.Cancel @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team) && invite.status == PENDING`

Описание:
- капитан отменяет приглашение

### 5.6 TeamInvite.Accept
`TeamInvite.Accept @ auth && TeamWriteWindow && policy.allow_team == true && invite.status == PENDING && invite.target_user_id == actor.user_id && CanJoinTeam(actor)`

Описание:
- пользователь принимает приглашение
- пользователь становится участником команды

Ограничения:
- `slots_open(vacancy) > 0`
- пользователь не является staff в этом хакатоне
- пользователь не имеет `particip.status in {PART_TEAM_MEMBER, PART_TEAM_CAPTAIN}` в этом хакатоне

### 5.7 TeamInvite.Reject
`TeamInvite.Reject @ auth && TeamWriteWindow && policy.allow_team == true && invite.status == PENDING && invite.target_user_id == actor.user_id`

Описание:
- пользователь отклоняет приглашение

### 5.8 JoinRequest.Create
`JoinRequest.Create @ auth && TeamWriteWindow && policy.allow_team == true && team.is_joinable == true && CanJoinTeam(actor) && can_seek_team && slots_open(vacancy) > 0`

Описание:
- пользователь отправляет заявку на вступление в команду на конкретную вакансию

### 5.9 JoinRequest.Cancel
`JoinRequest.Cancel @ auth && TeamWriteWindow && policy.allow_team == true && request.status == PENDING && request.requester_user_id == actor.user_id`

Описание:
- пользователь отменяет свою заявку

### 5.10 JoinRequest.Accept
`JoinRequest.Accept @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team) && request.status == PENDING && slots_open(vacancy) > 0`

Описание:
- капитан принимает заявку
- пользователь становится участником команды

Ограничения:
- пользователь не является staff в этом хакатоне
- пользователь не имеет `particip.status in {PART_TEAM_MEMBER, PART_TEAM_CAPTAIN}` в этом хакатоне

### 5.11 JoinRequest.Reject
`JoinRequest.Reject @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team) && request.status == PENDING`

Описание:
- капитан отклоняет заявку

### 5.12 TeamMember.Kick
`TeamMember.Kick @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team) && target is_team_member(team) && target.user_id != actor.user_id`

Описание:
- капитан исключает участника

Ограничения:
- исключение капитана запрещено

### 5.13 TeamMember.Leave (non-captain)
`TeamMember.Leave @ auth && TeamLeaveWindow && policy.allow_team == true && is_team_member(team) && actor is NOT team captain`

Описание:
- участник (не капитан) выходит из команды
- доступно и вне `REGISTRATION`

Ограничения:
- капитан не может выйти без передачи капитанства (и transfer доступен только в REGISTRATION)

### 5.14 TeamCaptain.Transfer
`TeamCaptain.Transfer @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team) && target is_team_member(team) && target.user_id != actor.user_id`

Описание:
- капитан передаёт капитанство

### 5.15 Team.Delete
`Team.Delete @ auth && TeamWriteWindow && policy.allow_team == true && is_team_captain(team) && members_count(team) == 1`

Описание:
- капитан удаляет команду, если он единственный участник

---

## 6) Побочные эффекты и интеграция

### 6.1 Конверсия участия в TEAM
При успешном выполнении действий:
- `Team.Create`               -> actor becomes captain
- `TeamInvite.Accept`         -> target becomes member
- `JoinRequest.Accept`        -> requester becomes member

Team service инициирует обновление участия в Role/Participation service:
- для create:
  - `particip.status = PART_TEAM_CAPTAIN`
  - `particip.team_id = team_id`
- для accept:
  - `particip.status = PART_TEAM_MEMBER`
  - `particip.team_id = team_id`

### 6.2 Выход из TEAM
При успешном выполнении действий:
- `TeamMember.Leave` (non-captain)
- `TeamMember.Kick`
- `Team.Delete` (captain alone)

Team service инициирует обновление участия в Role/Participation service:
- `particip.status = PART_LOOKING_FOR_TEAM`
- `particip.team_id = null`

### 6.3 Отмена конкурирующих заявок и инвайтов
При успешном выполнении действий:
- `TeamInvite.Accept`
- `JoinRequest.Accept`

Team service:
- отменяет все `PENDING` инвайты пользователя в другие команды этого хакатона
- отменяет все `PENDING` заявки пользователя в другие команды этого хакатона

---

## 7) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `Team.ReadCatalog` | teams + vacancies | Читать список команд | `auth && TeamReadWindow && (role in {OWNER, ORGANIZER, MENTOR} || is_participant)` | Доступно начиная с REGISTRATION, только staff/участникам |
| `Team.ReadById` | team + vacancies | Читать команду | `auth && TeamReadWindow && (role in {OWNER, ORGANIZER, MENTOR} || is_participant)` | Доступно начиная с REGISTRATION, только staff/участникам |
| `Team.ReadRoster` | team members | Читать состав команды | `auth && TeamReadWindow && (role in {OWNER, ORGANIZER, MENTOR} || is_participant)` | Доступно staff и участникам |
| `Team.ReadMyInbox` | invites + requests | Читать свои приглашения и заявки | `auth && TeamReadWindow && is_participant` | Доступно только участникам хакатона |
| `Team.ReadTeamInbox` | invites + requests | Читать входящие в команду | `auth && TeamReadWindow && is_team_captain(team)` | Доступно капитану |
| `Team.Create` | team fields | Создать команду | `auth && TeamWriteWindow && policy.allow_team && CanJoinTeam(actor)` | Только REGISTRATION |
| `Team.Update` | name, description, is_joinable | Обновить команду | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team)` | Только REGISTRATION |
| `Vacancy.Upsert` | vacancy fields | Создать/обновить вакансию | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team)` | Только REGISTRATION |
| `TeamInvite.Create` | invite fields | Пригласить в команду | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team) && slots_open(vacancy)>0` | Только REGISTRATION |
| `TeamInvite.Cancel` | invite id | Отменить приглашение | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team) && invite.status==PENDING` | Только REGISTRATION |
| `TeamInvite.Accept` | invite id | Принять приглашение | `auth && TeamWriteWindow && policy.allow_team && invite.status==PENDING && invite.target_user_id==actor.user_id && CanJoinTeam(actor)` | Только REGISTRATION |
| `TeamInvite.Reject` | invite id | Отклонить приглашение | `auth && TeamWriteWindow && policy.allow_team && invite.status==PENDING && invite.target_user_id==actor.user_id` | Только REGISTRATION |
| `JoinRequest.Create` | request fields | Отправить заявку | `auth && TeamWriteWindow && policy.allow_team && team.is_joinable && CanJoinTeam(actor) && can_seek_team && slots_open(vacancy)>0` | Только REGISTRATION |
| `JoinRequest.Cancel` | request id | Отменить заявку | `auth && TeamWriteWindow && policy.allow_team && request.status==PENDING && request.requester_user_id==actor.user_id` | Только REGISTRATION |
| `JoinRequest.Accept` | request id | Принять заявку | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team) && request.status==PENDING && slots_open(vacancy)>0` | Только REGISTRATION |
| `JoinRequest.Reject` | request id | Отклонить заявку | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team) && request.status==PENDING` | Только REGISTRATION |
| `TeamMember.Kick` | user_id | Исключить участника | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team) && target is_team_member(team) && target!=actor` | Только REGISTRATION |
| `TeamMember.Leave` | — | Выйти из команды (не капитан) | `auth && TeamLeaveWindow && policy.allow_team && is_team_member(team) && actor is NOT team captain` | Доступно и вне REGISTRATION |
| `TeamCaptain.Transfer` | user_id | Передать капитанство | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team) && target is_team_member(team) && target!=actor` | Только REGISTRATION |
| `Team.Delete` | — | Удалить команду | `auth && TeamWriteWindow && policy.allow_team && is_team_captain(team) && members_count(team)==1` | Только REGISTRATION |

---

## 8) validation_errors

### 8.1 Минимальная структура
- `code`: `REQUIRED | CONFLICT | FORBIDDEN | STAGE_RULE | POLICY_RULE | LIMIT_RULE`
- `field`: имя поля или группы
- `message`: строка
- `meta`: map для детализации (например: max_team_size, occupied_slots, etc.)