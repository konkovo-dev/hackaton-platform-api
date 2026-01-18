# Hackathon Policy Spec — Matchmaking domain (v1, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Цель домена

Matchmaking сервис строит рекомендации между:
- участниками хакатона, которые ищут команду
- командами и их вакансиями

Сервис не меняет доменные сущности. Он:
- потребляет события из Kafka
- поддерживает read-модель и индекс
- отдаёт ранжированные списки и объяснения (`reasons`)

---

## 1) Контекст и источники данных

### 1.1 Identity (глобально)
- `user.skills[]`: фиксированный список скиллов
- `user.skills_visibility` (для матчмейкинга игнорируется, т.к. MM доступен только внутри хакатона и по правилам хакатона)

### 1.2 Participation&Roles (внутри хакатона)
- `particip.kind ∈ {NONE, LOOKING_FOR_TEAM, SINGLE, TEAM}`
- `particip.team_id` (uuid | null)
- `particip.motivation_text` (string | null)
- `particip.wished_roles[]` (list<enum> + ANY)

### 1.3 Team (внутри хакатона)
- `team.teamname`
- `team.description`
- `team.is_joinable` (bool)
- `vacancy.vacancy_id`
- `vacancy.description`
- `vacancy.wished_roles[]` (list<enum> + ANY)
- `vacancy.wished_skills[]` (list<enum> + ANY)
- `vacancy.slots_total`
- `vacancy.slots_open`

### 1.4 Hackathon (внутри хакатона)
- `stage ∈ {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`
- `policy.allow_team` (bool)
- `policy.allow_individual` (bool)

---

## 2) Предикаты доступа

### 2.1 Окна стадий
- `MatchmakingWindow`: `stage == REGISTRATION`
  - цель: подбор актуален только во время регистрации

### 2.2 Кто может пользоваться матчмейкингом
- `CanBrowseTeams`: `auth && stage != DRAFT`
- `CanMatchTeamsForUser`: `auth && policy.allow_team == true && MatchmakingWindow && particip.kind in {LOOKING_FOR_TEAM, SINGLE, TEAM}`
- `CanMatchCandidatesForVacancy`: `auth && policy.allow_team == true && MatchmakingWindow && is_team_captain == true`

Примечание:
- MM не проверяет `is_team_captain` самостоятельно. Это проверка на write-пути TeamSvc.
- MM отдаёт кандидатов только капитану. Верификация капитанства выполняется через контекст TeamSvc.

---

## 3) Read-модели MatchmakingSvc

### 3.1 UserProfileIndex (глобально)
- `user_id`
- `skills[]`
- `updated_at`
- `version`

### 3.2 HackathonParticipantIndex (на хакатон)
- `hackathon_id`
- `user_id`
- `particip.kind`
- `particip.team_id`
- `wished_roles[]`
- `motivation_text`
- `is_staff` (bool)
- `updated_at`
- `version`

### 3.3 TeamIndex (на хакатон)
- `hackathon_id`
- `team_id`
- `teamname`
- `description`
- `is_joinable`
- `updated_at`
- `version`

### 3.4 VacancyIndex (на хакатон)
- `hackathon_id`
- `team_id`
- `vacancy_id`
- `description`
- `wished_roles[]`
- `wished_skills[]`
- `slots_total`
- `slots_open`
- `updated_at`
- `version`

---

## 4) Алгоритм скоринга (эвристика v1)

### 4.1 Фильтры (hard)
#### Для подбора команд пользователю
Команда/вакансия попадает в кандидаты, если:
- `policy.allow_team == true`
- `team.is_joinable == true`
- `vacancy.slots_open > 0`
- пользователь в состоянии, где он может вступать:
  - `particip.kind in {LOOKING_FOR_TEAM, SINGLE}`

#### Для подбора кандидатов капитану
Кандидат попадает в выборку, если:
- `particip.kind == LOOKING_FOR_TEAM`
- `is_staff == false`

### 4.2 Score (soft)
Состав:
- совпадение ролей: `role_match`
- совпадение скиллов: `skills_match`
- текстовая близость: `text_match` (опционально, v1 можно выключить)

Пример формулы v1:
- `score = 0.5*role_match + 0.5*skills_match + 0.0*text_match`

Где:
- `role_match`:
  - если в ролях есть `ANY` у любой стороны, то `role_match = 1`
  - иначе `|intersection(user.wished_roles, vacancy.wished_roles)| / |vacancy.wished_roles|`
- `skills_match`:
  - если в скиллах есть `ANY` у вакансии, то `skills_match = 1`
  - иначе `|intersection(user.skills, vacancy.wished_skills)| / |vacancy.wished_skills|`

Сервис возвращает `reasons[]`, например:
- `ROLE_MATCH: backend`
- `SKILL_MATCH: go`
- `OPEN_SLOT: 2`

---

## 5) API действий (read-only)

### 5.1 Matchmaking.SearchTeamsForUser
`Matchmaking.SearchTeamsForUser @ auth && policy.allow_team == true && MatchmakingWindow && particip.kind in {LOOKING_FOR_TEAM, SINGLE}`

Ответ:
- `team_id`
- `vacancy_id`
- `score`
- `reasons[]`
- `snippet` (teamname, description, vacancy description)

### 5.2 Matchmaking.SearchCandidatesForVacancy
`Matchmaking.SearchCandidatesForVacancy @ auth && policy.allow_team == true && MatchmakingWindow && is_team_captain == true`

Ответ:
- `user_id`
- `score`
- `reasons[]`
- `snippet` (skills subset, motivation snippet, wished_roles)

Ограничение выдачи:
- в v1 допускается выдавать только `LOOKING_FOR_TEAM`
- доступ только капитану, проверяется через контекст TeamSvc

### 5.3 Matchmaking.SearchTeamsCatalog
`Matchmaking.SearchTeamsCatalog @ auth && stage != DRAFT`

Ответ:
- список команд с фильтрами по ролям/скиллам/наличию слотов

---

## 6) Kafka события (источники → MatchmakingSvc)

Принцип:
- события являются upsert-снимками (`state snapshot`)
- содержат `version` и `occurred_at`
- ключ Kafka: сущность (`user_id`, `hackathon_id:user_id`, `hackathon_id:team_id`, `vacancy_id`)

### 6.1 Identity → Matchmaking
`identity.user_profile.upsert.v1`
- key: `user_id`
- payload:
  - `user_id`
  - `skills[]`
  - `updated_at`
  - `version`

Триггер:
- изменение `skills`

### 6.2 Participation&Roles → Matchmaking
`pr.hackathon_participant.upsert.v1`
- key: `{hackathon_id}:{user_id}`
- payload:
  - `hackathon_id`
  - `user_id`
  - `particip.kind`
  - `particip.team_id`
  - `wished_roles[]`
  - `motivation_text`
  - `is_staff`
  - `updated_at`
  - `version`

Триггеры:
- регистрация/изменение участия
- изменение `wished_roles`, `motivation_text`
- изменения staff-ролей (участник стал staff или наоборот)
- изменения `team_id` у участия

### 6.3 Team → Matchmaking
`team.team.upsert.v1`
- key: `{hackathon_id}:{team_id}`
- payload:
  - `hackathon_id`
  - `team_id`
  - `teamname`
  - `description`
  - `is_joinable`
  - `updated_at`
  - `version`

Триггеры:
- создание/редактирование команды
- изменение `is_joinable`

`team.vacancy.upsert.v1`
- key: `{hackathon_id}:{vacancy_id}`
- payload:
  - `hackathon_id`
  - `team_id`
  - `vacancy_id`
  - `description`
  - `wished_roles[]`
  - `wished_skills[]`
  - `slots_total`
  - `slots_open`
  - `updated_at`
  - `version`

Триггеры:
- создание/редактирование вакансии
- изменение слотов (приняли/вышли/кикнули)
- закрытие/открытие вакансии

`team.vacancy.deleted.v1`
- key: `{hackathon_id}:{vacancy_id}`
- payload:
  - `hackathon_id`
  - `vacancy_id`
  - `deleted_at`
  - `version`

Триггер:
- удаление вакансии

### 6.4 Hackathon → Matchmaking
`hackathon.stage.upsert.v1`
- key: `hackathon_id`
- payload:
  - `hackathon_id`
  - `stage`
  - `policy.allow_team`
  - `policy.allow_individual`
  - `updated_at`
  - `version`

Триггеры:
- публикация хакатона
- изменения времени, влияющие на stage
- изменения allow_team/allow_individual

---

## 7) Идемпотентность и порядок

### 7.1 Version
Каждое событие содержит `version` (монотонно растущий в рамках сущности).
Matchmaking применяет событие только если `version > stored.version`.

### 7.2 Повторная обработка
Повторное получение события с тем же `version` не меняет состояние.

### 7.3 Отставание
Matchmaking может временно отдавать устаревшие рекомендации.
Write-путь (TeamSvc accept/invite) является источником истины и повторно проверяет условия.

---

## 8) validation_errors

Matchmaking — read-only сервис. Ошибки:
- `FORBIDDEN` (нет доступа)
- `STAGE_RULE` (матчмейкинг недоступен на стадии)
- `NOT_READY` (данные ещё не прогреты, опционально)
- `NOT_FOUND` (нет сущности)

Примеры:
- `{code:"STAGE_RULE", field:"stage", message:"матчмейкинг недоступен на этой стадии"}`
- `{code:"FORBIDDEN", field:"auth", message:"требуется авторизация"}`
