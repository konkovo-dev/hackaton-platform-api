# Hackathon Policy Spec — Submissions domain (v1, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Role/Participation service)
- `roles[]`: `{OWNER, ORGANIZER, MENTOR, JURY}`
- `particip.kind`: `{NONE, LOOKING_FOR_TEAM, SINGLE, TEAM}`
- `particip.team_id`: `uuid | null`

### 0.2 Hackathon context (Hackathon service)
- `stage ∈ {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`
- `policy.allow_team` (bool)
- `policy.allow_individual` (bool)

### 0.3 Team context (Team service)
- `is_team_member(team_id, actor.user_id)` (bool)
- `is_team_captain(team_id, actor.user_id)` (bool)

### 0.4 Submissions context (Submissions service)
`Submission`:
- `submission_id`
- `hackathon_id`
- `team_id` (uuid | null)
- `author_user_id`
- `created_at`
- `payload` (структура посылки, пока не конкретизируем)

`FinalSelection`:
- `hackathon_id`
- `team_id` (uuid | null)
- `final_submission_id` (uuid | null)

---

## 1) Предикаты

### 1.1 Окна стадий
- `SubmissionsWriteWindow`: `stage == RUNNING`
- `SubmissionsReadWindow`: `stage in {RUNNING, JUDGING, FINISHED}`

### 1.2 Разрешённые типы участия
- `CanSubmitTeam`: `particip.kind == TEAM && policy.allow_team == true && particip.team_id != null`
- `CanSubmitSingle`: `particip.kind == SINGLE && policy.allow_individual == true`

### 1.3 Привязка к команде
- `is_my_team_submission(s)`: `particip.team_id != null && s.team_id == particip.team_id`
- `is_my_team_captain`: `particip.team_id != null && is_team_captain(particip.team_id, actor.user_id) == true`

---

## 2) Правила видимости

### 2.1 Командные посылки
Командные посылки доступны только участникам этой команды:
- `particip.kind == TEAM`
- `s.team_id == particip.team_id`

### 2.2 Индивидуальные посылки
Индивидуальные посылки доступны только автору:
- `particip.kind == SINGLE`
- `s.author_user_id == actor.user_id`

---

## 3) Политика чтения

### 3.1 Submissions.ReadMy
`Submissions.ReadMy @ auth && SubmissionsReadWindow && (particip.kind in {SINGLE, TEAM})`

Описание:
- участник читает свои посылки
- для `TEAM` возвращаются только посылки своей команды
- для `SINGLE` возвращаются только посылки автора

### 3.2 Submissions.ReadFinalMy
`Submissions.ReadFinalMy @ auth && SubmissionsReadWindow && (particip.kind in {SINGLE, TEAM})`

Описание:
- участник читает финальную посылку
- если финальная не выбрана явно, финальная определяется по правилу `DefaultFinal`

---

## 4) Политика записи

### 4.1 Submissions.Upload
`Submissions.Upload @ auth && SubmissionsWriteWindow && (CanSubmitTeam || CanSubmitSingle)`

Описание:
- участник загружает посылку
- для `TEAM` посылка привязывается к `team_id = particip.team_id`
- для `SINGLE` посылка создаётся без `team_id`, с `author_user_id = actor.user_id`

Ограничения:
- если `particip.kind == TEAM`, то `particip.team_id != null`

### 4.2 Submissions.SetFinal
`Submissions.SetFinal @ auth && SubmissionsWriteWindow && (is_my_team_captain || (particip.kind == SINGLE && CanSubmitSingle)) && submission принадлежит актору`

Описание:
- капитан выбирает финальную посылку своей команды
- индивидуальный участник выбирает финальную посылку для себя

Ограничения:
- выбранная посылка должна принадлежать актору:
  - `TEAM`: `submission.team_id == particip.team_id`
  - `SINGLE`: `submission.author_user_id == actor.user_id`

---

## 5) DefaultFinal

`DefaultFinal` применяется при чтении финальной посылки, если `FinalSelection.final_submission_id == null`.

Правило:
- финальная посылка = последняя по `created_at` среди посылок актора в пределах `stage == RUNNING`

Поведение при чтении после RUNNING:
- если финальная посылка не была выбрана явно в RUNNING, то финальной считается последняя посылка, загруженная до конца RUNNING.

---

## 6) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `Submissions.ReadMy` | submissions list | Читать свои посылки | `Submissions.ReadMy @ auth && SubmissionsReadWindow && particip.kind in {SINGLE, TEAM}` | Доступно участникам начиная с RUNNING |
| `Submissions.ReadFinalMy` | final submission | Читать финальную посылку | `Submissions.ReadFinalMy @ auth && SubmissionsReadWindow && particip.kind in {SINGLE, TEAM}` | Доступно участникам начиная с RUNNING |
| `Submissions.Upload` | submission payload | Загрузить посылку | `Submissions.Upload @ auth && SubmissionsWriteWindow && (CanSubmitTeam || CanSubmitSingle)` | Доступно только в RUNNING |
| `Submissions.SetFinal` | submission_id | Выбрать финальную посылку | `Submissions.SetFinal @ auth && SubmissionsWriteWindow && (is_my_team_captain || (particip.kind == SINGLE && CanSubmitSingle)) && submission принадлежит актору` | Доступно только в RUNNING |

---

## 7) validation_errors

### 7.1 Минимальная структура
- `code`: `REQUIRED | FORBIDDEN | STAGE_RULE | POLICY_RULE | NOT_FOUND`
- `field`: имя поля или группы
- `message`: строка

### 7.2 Примеры
- `{code:"STAGE_RULE", field:"stage", message:"операция разрешена только в RUNNING"}`
- `{code:"POLICY_RULE", field:"policy.allow_individual", message:"индивидуальные посылки запрещены"}`
- `{code:"FORBIDDEN", field:"team_id", message:"нет доступа к посылкам другой команды"}`
- `{code:"NOT_FOUND", field:"submission_id", message:"посылка не найдена"}`
