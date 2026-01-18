# Hackathon Policy Spec — Judging domain (v1.1, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Role/Participation service)
- `roles[]`: `{OWNER, ORGANIZER, MENTOR, JURY}`
- `particip.kind`: `{NONE, LOOKING_FOR_TEAM, SINGLE, TEAM}`
- `particip.team_id`: `uuid | null`

### 0.2 Hackathon context (Hackathon service)
- `stage ∈ {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`
- `published_at`: `timestamp | null`

`Result`:
- `result_draft` (struct | null)
- `result_published_at` (timestamp | null)

Условия:
- `ResultDraftWindow`: `stage == JUDGING && result_published_at == null`
- `ResultPublished`: `result_published_at != null`

### 0.3 Team context (Team service)
- `teams_in_hackathon(hackathon_id)` (list<team_id>)

### 0.4 Submissions context (Submissions service)
- `final_submission_id_for_team(hackathon_id, team_id)` (uuid | null)
- `final_submission_payload(submission_id)` (struct)

### 0.5 Judging context (Judging service)

`JudgingItem`:
- `item_id`
- `hackathon_id`
- `team_id`
- `final_submission_id` (uuid | null)

`Assignment`:
- `assignment_id`
- `hackathon_id`
- `jury_user_id`
- `item_id`
- `status ∈ {PENDING, DONE}`

`Verdict`:
- `verdict_id`
- `hackathon_id`
- `jury_user_id`
- `item_id`
- `score` (int)
- `comment` (string)
- `created_at`
- `updated_at`

---

## 1) Предикаты

### 1.1 Окна стадий
- `JudgingWindow`: `stage == JUDGING`
- `VerdictWriteWindow`: `stage == JUDGING && result_published_at == null`
- `VerdictReadWindow`: `stage in {JUDGING, FINISHED}`

### 1.2 Роли
- `is_owner_or_organizer`: `role in {OWNER, ORGANIZER}`
- `is_jury`: `role in {JURY}`

### 1.3 Принадлежность назначений
- `is_my_assignment(item_id)`: существует `Assignment` с `jury_user_id == actor.user_id` и `item_id == item_id`

---

## 2) Основные ограничения

### 2.1 Анонимность работ для жюри
Для любых действий чтения со стороны роли `JURY`:
- не возвращаются `team_id`, данные команды, состав команды и ссылки на команду
- объект оценки возвращается как `JudgingItemView(item_id, payload, status)` без идентичности команды

Для OWNER/ORGANIZER анонимность не применяется:
- им доступен `team_id` и полная связка для формирования результата

### 2.2 В судействе участвуют только команды
- оценка выполняется по `TEAM`-работам
- объект оценки берётся из финальной посылки команды

### 2.3 Вердикты закрываются публикацией результата
Если `ResultPublished`, то:
- создание и редактирование вердиктов запрещено
- чтение работ и вердиктов разрешено

---

## 3) Политика чтения

### 3.1 Judging.ReadItems
`Judging.ReadItems @ auth && is_jury && VerdictReadWindow`

Описание:
- жюри читает список всех объектов оценки в хакатоне
- в ответе возвращаются `item_id`, `payload`, агрегированные статусы по назначению самого жюри
- `team_id` не возвращается

### 3.2 Judging.ReadMyAssignments
`Judging.ReadMyAssignments @ auth && is_jury && VerdictReadWindow`

Описание:
- жюри читает только свои назначения
- в ответе возвращаются `assignment_id`, `item_id`, `status`
- `team_id` не возвращается

### 3.3 Judging.ReadVerdicts
`Judging.ReadVerdicts @ auth && is_jury && VerdictReadWindow`

Описание:
- жюри читает все вердикты по всем объектам оценки
- в ответе возвращаются `item_id`, `jury_user_id`, `score`, `comment`
- `team_id` не возвращается

### 3.4 Judging.ReadAllItems
`Judging.ReadAllItems @ auth && is_owner_or_organizer && VerdictReadWindow`

Описание:
- OWNER/ORGANIZER читают все объекты оценки
- в ответе возвращаются `item_id`, `team_id`, `final_submission_id`, `payload`

### 3.5 Judging.ReadAllAssignments
`Judging.ReadAllAssignments @ auth && is_owner_or_organizer && VerdictReadWindow`

Описание:
- OWNER/ORGANIZER читают все назначения

### 3.6 Judging.ReadAllVerdicts
`Judging.ReadAllVerdicts @ auth && is_owner_or_organizer && VerdictReadWindow`

Описание:
- OWNER/ORGANIZER читают все вердикты

---

## 4) Политика записи

### 4.1 Judging.AssignRandom
`Judging.AssignRandom @ auth && is_owner_or_organizer && JudgingWindow && result_published_at == null`

Описание:
- OWNER/ORGANIZER запускают распределение работ между жюри
- распределение выполняется случайно
- создаются `JudgingItem` для каждой команды, у которой есть финальная посылка
- создаются `Assignment` для жюри по `item_id`

Ограничения:
- назначаются только команды, у которых `final_submission_id != null`
- количество назначений на одного жюри вычисляется как:
  - `target_per_jury = ceil(items_count / jury_count)`
- повторный запуск допускается только если назначения не созданы, либо явно выполнен `Judging.ResetAssignments`

### 4.2 Judging.ResetAssignments
`Judging.ResetAssignments @ auth && is_owner_or_organizer && JudgingWindow && result_published_at == null`

Описание:
- OWNER/ORGANIZER сбрасывают `Assignment` и `Verdict`

Ограничения:
- действие запрещено, если `ResultPublished`

### 4.3 Verdict.UpsertMy
`Verdict.UpsertMy @ auth && is_jury && VerdictWriteWindow && is_my_assignment(item_id)`

Описание:
- жюри создаёт или обновляет свой вердикт по объекту оценки, который назначен этому жюри

Ограничения:
- `score` обязателен
- `comment` может быть пустым
- вердикт возможен только по `item_id`, который назначен этому жюри

---

## 5) Связь с публикацией результата

### 5.1 ResultDraft
`Hackathon.Result` редактируется OWNER/ORGANIZER только в `ResultDraftWindow`.

### 5.2 ResultPublish
После `Hackathon.Result.Publish`:
- `result_published_at` устанавливается один раз
- `stage` переходит в `FINISHED`
- `VerdictWriteWindow` закрывается

---

## 6) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `Judging.ReadItems` | items | Читать все работы | `Judging.ReadItems @ auth && is_jury && VerdictReadWindow` | Доступно жюри в JUDGING/FINISHED, анонимно |
| `Judging.ReadMyAssignments` | assignments | Читать свои назначения | `Judging.ReadMyAssignments @ auth && is_jury && VerdictReadWindow` | Доступно жюри в JUDGING/FINISHED |
| `Judging.ReadVerdicts` | verdicts | Читать все вердикты | `Judging.ReadVerdicts @ auth && is_jury && VerdictReadWindow` | Доступно жюри в JUDGING/FINISHED, анонимно |
| `Judging.ReadAllItems` | items + payload | Читать все работы | `Judging.ReadAllItems @ auth && is_owner_or_organizer && VerdictReadWindow` | Доступно OWNER/ORGANIZER в JUDGING/FINISHED |
| `Judging.ReadAllAssignments` | assignments | Читать все назначения | `Judging.ReadAllAssignments @ auth && is_owner_or_organizer && VerdictReadWindow` | Доступно OWNER/ORGANIZER в JUDGING/FINISHED |
| `Judging.ReadAllVerdicts` | verdicts | Читать все вердикты | `Judging.ReadAllVerdicts @ auth && is_owner_or_organizer && VerdictReadWindow` | Доступно OWNER/ORGANIZER в JUDGING/FINISHED |
| `Judging.AssignRandom` | assignments | Случайно распределить работы | `Judging.AssignRandom @ auth && is_owner_or_organizer && JudgingWindow && result_published_at == null` | Доступно OWNER/ORGANIZER в JUDGING до публикации результата |
| `Judging.ResetAssignments` | assignments + verdicts | Сбросить распределение | `Judging.ResetAssignments @ auth && is_owner_or_organizer && JudgingWindow && result_published_at == null` | Доступно OWNER/ORGANIZER в JUDGING до публикации результата |
| `Verdict.UpsertMy` | score, comment | Создать/обновить вердикт | `Verdict.UpsertMy @ auth && is_jury && VerdictWriteWindow && is_my_assignment(item_id)` | Доступно жюри только по своим назначениям в JUDGING |

---

## 7) validation_errors

### 7.1 Минимальная структура
- `code`: `REQUIRED | FORBIDDEN | STAGE_RULE | CONFLICT | NOT_FOUND`
- `field`: имя поля или группы
- `message`: строка

### 7.2 Примеры
- `{code:"STAGE_RULE", field:"stage", message:"операция разрешена только в JUDGING"}`
- `{code:"FORBIDDEN", field:"roles", message:"доступно только для роли JURY"}`
- `{code:"FORBIDDEN", field:"item_id", message:"нет назначения на этот объект оценки"}`
- `{code:"REQUIRED", field:"score", message:"score обязателен"}`
- `{code:"CONFLICT", field:"result_published_at", message:"вердикты запрещены после публикации результата"}`
