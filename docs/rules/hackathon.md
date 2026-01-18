# Hackathon Policy Spec — Hackathon domain (v1.5, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Role/Participation service)
- `roles[]`: `{OWNER, ORGANIZER, MENTOR, JURY}`
- `particip.kind`: `{NONE, LOOKING_FOR_TEAM, SINGLE, TEAM}`

### 0.2 Hackathon context (Hackathon service)
- `published_at`: `timestamp | null`
- `result_published_at`: `timestamp | null`
- `times`:
  - `registration_opens_at`
  - `registration_closes_at`
  - `starts_at`
  - `ends_at`
  - `judging_ends_at`
- `policy`:
  - `allow_team` (bool)
  - `allow_individual` (bool)
  - `team_size_max` (int | null)
- `fields`:
  - `name` (string)
  - `desc` (string)
  - `sh_desc` (string)
  - `location` (string)
  - `links` (отдельная сущность)
  - `task` (отдельная сущность)
  - `result` (отдельная сущность)
- `messages` (отдельная сущность)

### 0.3 Stage (Hackathon service)
- `stage ∈ {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`
- `stage == DRAFT` ⇔ `published_at == null`
- если `published_at != null`, то:
  - `UPCOMING`     : `now < registration_opens_at`
  - `REGISTRATION` : `registration_opens_at <= now < registration_closes_at`
  - `PRESTART`     : `registration_closes_at <= now < starts_at`
  - `RUNNING`      : `starts_at <= now < ends_at`
  - `JUDGING`      : `ends_at <= now < judging_ends_at` и `result_published_at == null`
  - `FINISHED`     : `now >= judging_ends_at` или `result_published_at != null`

---

## 1) Словарь предикатов

### 1.1 Роли и участие
- `role in {…}`: actor имеет хотя бы одну роль из множества
- `particip.kind in {…}`: тип участия входит в множество

### 1.2 Время и сравнения
- `now < old.<field>`
- `now < new.<field>`
- `old.<field> < new.<field>`

### 1.3 Валидационные предикаты
- `TIME_RULE(x)`
- `at_least_one_true(x.allow_team, x.allow_individual)`

### 1.4 Предикаты доступа для сообщений
- `is_staff`: `role in {OWNER, ORGANIZER, MENTOR, JURY}`
- `is_participant`: `particip.kind in {LOOKING_FOR_TEAM, SINGLE, TEAM}`
- `can_read_messages`: `is_staff OR is_participant`

---

## 2) TIME_RULE

`TIME_RULE(x)` выполняется, если:
- `registration_opens_at < registration_closes_at - 1h <= starts_at < ends_at - 1h <= judging_ends_at`

---

## 3) Типы изменений времени

TYPE-A:
- поля: `registration_opens_at`, `judging_ends_at`
- ограничение вне DRAFT: `now < old.field && now < new.field`

TYPE-B:
- поля: `registration_closes_at`, `starts_at`, `ends_at`
- ограничение вне DRAFT: `now < old.field && old.field < new.field`

---

## 4) Режимы валидации состояния

### 4.1 DRAFT (мягкий режим)
- Обновления разрешены при выполнении условий доступа по роли/стадии.
- Сохранение допускается при наличии ошибок.
- Ответ содержит `validation_errors` для строгого режима.

### 4.2 Опубликованные стадии (строгий режим)
- Если `stage != DRAFT`, то при наличии строгих ошибок запрос отклоняется.
- Ошибки возвращаются в `validation_errors`.

---

## 5) Строгие требования для `stage != DRAFT`

### 5.1 Обязательные поля
- `name != ""`
- `location != ""`
- `task` задан и валиден
- временные поля заданы: `registration_opens_at`, `registration_closes_at`, `starts_at`, `ends_at`, `judging_ends_at`

`desc`, `links`, `sh_desc` не обязательны.

### 5.2 Время и ограничения редактирования
Для любых изменений времени в `stage != DRAFT`:
- всегда: `TIME_RULE(new)`

Дополнительно по полям:
- `registration_opens_at`: `now < old.registration_opens_at && now < new.registration_opens_at`
- `registration_closes_at`: `now < old.registration_closes_at && old.registration_closes_at < new.registration_closes_at`
- `starts_at`: `now < old.starts_at && old.starts_at < new.starts_at`
- `ends_at`: `now < old.ends_at && old.ends_at < new.ends_at`
- `judging_ends_at`: `now < old.judging_ends_at && now < new.judging_ends_at`

### 5.3 Policy
- всегда: `at_least_one_true(policy.allow_team, policy.allow_individual)`

Ограничения на изменение типов регистрации:
- выключать тип регистрации можно при `stage in {DRAFT, UPCOMING}`
- включать тип регистрации можно при `stage == DRAFT`

Ограничения на `team_size_max`:
- менять можно при `stage in {DRAFT, UPCOMING}`

### 5.4 Задание
- редактирование `task` запрещено при `stage in {JUDGING, FINISHED}`

---

## 6) PublishReady(old)

`PublishReady(old)` выполняется, если:
- `old.name != ""`
- `old.location != ""`
- временные поля заданы: `old.registration_opens_at`, `old.registration_closes_at`, `old.starts_at`, `old.ends_at`, `old.judging_ends_at`
- `TIME_RULE(old)`
- `at_least_one_true(old.allow_team, old.allow_individual)`
- `old.task` задан и валиден

`desc` и `sh_desc` могут быть пустыми.

---

## 7) ResultReady(old)

`ResultReady(old)` выполняется, если:
- `old.result` задан и валиден

---

## 8) Messages model

`HackathonMessage`:
- `id`
- `hackathon_id`
- `title` (string | null)
- `body` (string)
- `created_by_user_id`
- `created_at`
- `updated_at`
- `deleted_at` (timestamp | null)

---

## 9) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `Hackathon.ReadPublic` | `name, desc, sh_desc, location, links, times, policy` | Читать опубликованные данные хакатона | `Hackathon.ReadPublic @ stage != DRAFT` | Доступно только для опубликованных хакатонов |
| `Hackathon.ReadDraft` | `name, desc, sh_desc, location, links, times, policy, task` | Читать черновик хакатона | `Hackathon.ReadDraft @ stage == DRAFT && role in {OWNER, ORGANIZER}` | Доступно OWNER и ORGANIZER |
| `Hackathon.ReadTask` | `task` | Читать задание хакатона | `Hackathon.ReadTask @ (role in {OWNER, ORGANIZER}) OR (stage != DRAFT && role in {MENTOR, JURY}) OR (stage == RUNNING && particip.kind in {SINGLE, TEAM})` | Доступно по правилам доступа к заданию |
| `Hackathon.ReadResultPublic` | `result` | Читать опубликованный результат хакатона | `Hackathon.ReadResultPublic @ stage == FINISHED` | Доступно всем после публикации результата |
| `Hackathon.ReadResultDraft` | `result` | Читать черновик результата | `Hackathon.ReadResultDraft @ stage == JUDGING && result_published_at == null && role in {OWNER, ORGANIZER}` | Доступно OWNER и ORGANIZER на JUDGING до публикации результата |
| `Hackathon.Create` | все поля | Создать хакатон | `Hackathon.Create @ auth` | Доступно авторизованному пользователю; создатель получает роль OWNER |
| `Hackathon.Publish` | `published_at` | Опубликовать хакатон | `Hackathon.Publish @ stage == DRAFT && role in {OWNER} && now < old.registration_opens_at && PublishReady(old)` | Доступно OWNER; `published_at` устанавливается один раз |
| `Hackathon.UpdateBasics` | `name, desc, sh_desc` | Обновить контент хакатона | `Hackathon.UpdateBasics @ role in {OWNER, ORGANIZER}` | Доступно OWNER и ORGANIZER |
| `Hackathon.UpdateLocation` | `location` | Обновить локацию | `Hackathon.UpdateLocation @ role in {OWNER, ORGANIZER} && stage in {DRAFT, UPCOMING, REGISTRATION, PRESTART}` | Доступно OWNER и ORGANIZER |
| `Hackathon.UpdateLinks` | `links` | Обновить ссылки | `Hackathon.UpdateLinks @ role in {OWNER, ORGANIZER}` | Доступно OWNER и ORGANIZER |
| `Hackathon.UpdatePolicy.TeamSizeMax` | `team_size_max` | Обновить максимальный размер команды | `Hackathon.UpdatePolicy.TeamSizeMax @ role in {OWNER, ORGANIZER} && stage in {DRAFT, UPCOMING}` | Доступно OWNER и ORGANIZER |
| `Hackathon.UpdatePolicy.DisableType` | `allow_team` и/или `allow_individual` | Выключить тип регистрации | `Hackathon.UpdatePolicy.DisableType @ role in {OWNER, ORGANIZER} && stage in {DRAFT, UPCOMING} && at_least_one_true(new.allow_team, new.allow_individual)` | Доступно OWNER и ORGANIZER |
| `Hackathon.UpdatePolicy.EnableType` | `allow_team` и/или `allow_individual` | Включить тип регистрации | `Hackathon.UpdatePolicy.EnableType @ role in {OWNER, ORGANIZER} && stage == DRAFT && at_least_one_true(new.allow_team, new.allow_individual)` | Доступно OWNER и ORGANIZER |
| `Hackathon.UpdateSchedule.RegistrationOpensAt` | `registration_opens_at` | Обновить время открытия регистрации | `Hackathon.UpdateSchedule.RegistrationOpensAt @ role in {OWNER, ORGANIZER} && TIME_RULE(new) && (stage == DRAFT || (now < old.registration_opens_at && now < new.registration_opens_at))` | Правило TYPE-A |
| `Hackathon.UpdateSchedule.RegistrationClosesAt` | `registration_closes_at` | Обновить время закрытия регистрации | `Hackathon.UpdateSchedule.RegistrationClosesAt @ role in {OWNER, ORGANIZER} && TIME_RULE(new) && (stage == DRAFT || (now < old.registration_closes_at && old.registration_closes_at < new.registration_closes_at))` | Правило TYPE-B |
| `Hackathon.UpdateSchedule.StartsAt` | `starts_at` | Обновить время начала хакатона | `Hackathon.UpdateSchedule.StartsAt @ role in {OWNER, ORGANIZER} && TIME_RULE(new) && (stage == DRAFT || (now < old.starts_at && old.starts_at < new.starts_at))` | Правило TYPE-B |
| `Hackathon.UpdateSchedule.EndsAt` | `ends_at` | Обновить время окончания хакатона | `Hackathon.UpdateSchedule.EndsAt @ role in {OWNER, ORGANIZER} && TIME_RULE(new) && (stage == DRAFT || (now < old.ends_at && old.ends_at < new.ends_at))` | Правило TYPE-B |
| `Hackathon.UpdateSchedule.JudgingEndsAt` | `judging_ends_at` | Обновить время окончания судейства | `Hackathon.UpdateSchedule.JudgingEndsAt @ role in {OWNER, ORGANIZER} && TIME_RULE(new) && (stage == DRAFT || (now < old.judging_ends_at && now < new.judging_ends_at))` | Правило TYPE-A |
| `Hackathon.UpdateTask` | `task` | Обновить задание хакатона | `Hackathon.UpdateTask @ role in {OWNER, ORGANIZER} && stage in {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING} && new.task != ""` | Доступно OWNER и ORGANIZER |
| `Hackathon.UpdateResultDraft` | `result` | Обновить черновик результата | `Hackathon.UpdateResultDraft @ role in {OWNER, ORGANIZER} && stage == JUDGING && result_published_at == null` | Доступно OWNER и ORGANIZER |
| `Hackathon.PublishResult` | `result_published_at, judging_ends_at` | Опубликовать результат | `Hackathon.PublishResult @ role in {OWNER, ORGANIZER} && stage == JUDGING && result_published_at == null && ResultReady(old)` | Доступно OWNER и ORGANIZER; `result_published_at` устанавливается один раз; `judging_ends_at = now` |
| `HackathonMessage.Create` | message fields | Создать сообщение хакатона | `HackathonMessage.Create @ stage != DRAFT && role in {OWNER, ORGANIZER}` | Доступно OWNER и ORGANIZER; создание запрещено в DRAFT |
| `HackathonMessage.Read` | message fields | Читать сообщения хакатона | `HackathonMessage.Read @ stage != DRAFT && can_read_messages` | Доступно staff или любому participant; чтение запрещено в DRAFT |
| `HackathonMessage.Update` | message fields | Обновить сообщение хакатона | `HackathonMessage.Update @ stage != DRAFT && role in {OWNER, ORGANIZER}` | Доступно OWNER и ORGANIZER; обновление запрещено в DRAFT |
| `HackathonMessage.Delete` | message id | Удалить сообщение хакатона | `HackathonMessage.Delete @ stage != DRAFT && role in {OWNER, ORGANIZER}` | Доступно OWNER и ORGANIZER; удаление запрещено в DRAFT |

---

## 10) validation_errors

### 10.1 Назначение
- `stage == DRAFT`: возвращаются всегда
- `stage != DRAFT`: возвращаются при ошибке

### 10.2 Минимальная структура
- `code`: `REQUIRED | TIME_RULE | TIME_LOCKED | TYPE_RULE | POLICY_RULE | FORBIDDEN`
- `field`: имя поля или группы
- `message`: строка
