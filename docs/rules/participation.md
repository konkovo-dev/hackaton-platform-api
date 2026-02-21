# Hackathon Policy Spec — Participation domain (v1, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Role/Participation service)
- `roles[]`: `{OWNER, ORGANIZER, MENTOR, JURY}`
- `particip.status`: `{NONE, INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM, TEAM_MEMBER, TEAM_CAPTAIN}`
- `particip.team_id`: `uuid | null`

### 0.2 Hackathon context (Hackathon service)
- `stage ∈ {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`
- `policy.allow_individual` (bool)
- `policy.allow_team` (bool)

### 0.3 Participation context (Participation service)
`HackathonParticipation`:
- `hackathon_id`
- `user_id`
- `status ∈ {NONE, INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM, TEAM_MEMBER, TEAM_CAPTAIN}`
- `team_id` (uuid | null)
- `profile`:
  - `wished_roles[]` (list<enum>)
  - `motivation_text` (string | null)
- `registered_at`
- `updated_at`

---

## 1) Инварианты домена

### 1.1 Взаимоисключение staff и participation
В одном хакатоне запрещено одновременно:
- иметь любую роль из `{OWNER, ORGANIZER, MENTOR, JURY}`
- и иметь участие `particip.status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM, TEAM_MEMBER, TEAM_CAPTAIN}`

### 1.2 team_id заполнен только для TEAM статусов
- если `status in {TEAM_MEMBER, TEAM_CAPTAIN}`, то `team_id != null`
- если `status in {NONE, INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}`, то `team_id == null`

### 1.3 profile актуален только для non-TEAM статусов
- `profile.wished_roles` используется только при `status == LOOKING_FOR_TEAM`
- `profile.motivation_text` может быть задан при `status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}`
- при `status in {TEAM_MEMBER, TEAM_CAPTAIN}` profile сохраняется, но не используется

---

## 2) Предикаты

### 2.1 Окна стадий
- `ParticipationWriteWindow`: `stage == REGISTRATION`
- `ParticipationReadWindow`: `stage in {REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`

### 2.2 Разрешенные статусы
- `CanRegister`: `particip.status == NONE && !has_staff_role`
- `CanSwitch`: `particip.status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}`
- `CanUpdate`: `particip.status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}`
- `CanUnregister`: `particip.status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}`

### 2.3 Роли и доступ
- `is_staff`: `role in {OWNER, ORGANIZER, MENTOR, JURY}`
- `is_participant`: `particip.status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM, TEAM_MEMBER, TEAM_CAPTAIN}`

---

## 3) Политика чтения

### 3.1 Participation.ReadMy
`Participation.ReadMy @ auth && ParticipationReadWindow && is_participant`

Описание:
- участник читает свой профиль участия в хакатоне
- доступен начиная с REGISTRATION

### 3.2 Participation.ReadUser
`Participation.ReadUser @ auth && ParticipationReadWindow && is_staff`

Описание:
- OWNER/ORGANIZER/MENTOR читают профиль участника
- возвращается полный профиль включая wished_roles и motivation_text

### 3.3 Participation.ListParticipants
`Participation.ListParticipants @ auth && ParticipationReadWindow && is_staff`

Описание:
- OWNER/ORGANIZER/MENTOR читают список участников
- поддерживается фильтрация по статусам и wished_roles
- пагинация через cursor

---

## 4) Политика записи

### 4.1 Participation.Register
`Participation.Register @ auth && ParticipationWriteWindow && CanRegister && desired_status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM} && (policy.allow_individual || policy.allow_team)`

Описание:
- пользователь регистрируется на хакатон
- выбирает между INDIVIDUAL_ACTIVE или LOOKING_FOR_TEAM
- TEAM статусы не допускаются при регистрации

Ограничения:
- если `desired_status == INDIVIDUAL_ACTIVE`, то `policy.allow_individual == true`
- если `desired_status == LOOKING_FOR_TEAM`, то `policy.allow_team == true`
- если `desired_status == LOOKING_FOR_TEAM`, то `profile.wished_roles` рекомендуется заполнить

Побочные эффекты:
- создается `HackathonParticipation` со статусом `desired_status`
- при `status == LOOKING_FOR_TEAM` отправляется событие в Matchmaking Service

### 4.2 Participation.Update
`Participation.Update @ auth && ParticipationWriteWindow && CanUpdate`

Описание:
- участник обновляет профиль для матчмейкинга
- изменяет `wished_roles` и `motivation_text`

Ограничения:
- запрещено при `status in {TEAM_MEMBER, TEAM_CAPTAIN}`

Побочные эффекты:
- если `status == LOOKING_FOR_TEAM`, отправляется обновленное событие в Matchmaking Service

### 4.3 Participation.SwitchMode
`Participation.SwitchMode @ auth && ParticipationWriteWindow && CanSwitch && new_status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM} && new_status != current_status`

Описание:
- участник переключается между INDIVIDUAL_ACTIVE и LOOKING_FOR_TEAM
- профиль сохраняется

Ограничения:
- запрещено при `status in {TEAM_MEMBER, TEAM_CAPTAIN}`
- `new_status` должен отличаться от текущего
- если переключение в `LOOKING_FOR_TEAM`, то `policy.allow_team == true`
- если переключение в `INDIVIDUAL_ACTIVE`, то `policy.allow_individual == true`

Побочные эффекты:
- если `new_status == LOOKING_FOR_TEAM`, отправляется событие в Matchmaking Service
- если `old_status == LOOKING_FOR_TEAM`, отправляется событие удаления из Matchmaking Service

### 4.4 Participation.Unregister
`Participation.Unregister @ auth && ParticipationWriteWindow && CanUnregister`

Описание:
- участник отменяет регистрацию на хакатон

Ограничения:
- запрещено при `status in {TEAM_MEMBER, TEAM_CAPTAIN}`
- чтобы отменить участие в команде, нужно сначала выйти из команды

Побочные эффекты:
- `status` устанавливается в `NONE`
- `team_id` очищается
- если был `LOOKING_FOR_TEAM`, отправляется событие удаления из Matchmaking Service

---

## 5) Внутренние операции (service-to-service)

### 5.1 Participation.ConvertToTeam
`Participation.ConvertToTeam @ service_to_service && status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM} && team_id provided`

Описание:
- Team Service вызывает при создании команды, вступлении в команду, принятии приглашения
- конвертирует участие в TEAM_MEMBER или TEAM_CAPTAIN

Ограничения:
- может конвертировать только из `{INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}`
- `team_id` обязателен
- `is_captain` определяет итоговый статус

Побочные эффекты:
- `status` устанавливается в `TEAM_MEMBER` или `TEAM_CAPTAIN`
- `team_id` заполняется
- `profile` сохраняется, но не используется
- если был `LOOKING_FOR_TEAM`, отправляется событие удаления из Matchmaking Service

### 5.2 Participation.ConvertFromTeam
`Participation.ConvertFromTeam @ service_to_service && status in {TEAM_MEMBER, TEAM_CAPTAIN}`

Описание:
- Team Service вызывает при выходе из команды, исключении, удалении команды
- конвертирует обратно в LOOKING_FOR_TEAM

Ограничения:
- может конвертировать только из `{TEAM_MEMBER, TEAM_CAPTAIN}`

Побочные эффекты:
- `status` устанавливается в `LOOKING_FOR_TEAM`
- `team_id` очищается
- `profile` восстанавливается к прежнему использованию
- отправляется событие в Matchmaking Service

---

## 6) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `Participation.Register` | status, profile | Зарегистрироваться на хакатон | `Participation.Register @ auth && ParticipationWriteWindow && CanRegister && desired_status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}` | Доступно только на REGISTRATION |
| `Participation.ReadMy` | participation | Читать своё участие | `Participation.ReadMy @ auth && ParticipationReadWindow && is_participant` | Доступно участникам начиная с REGISTRATION |
| `Participation.Update` | profile | Обновить профиль | `Participation.Update @ auth && ParticipationWriteWindow && CanUpdate` | Доступно только на REGISTRATION, не в команде |
| `Participation.SwitchMode` | status | Переключить режим | `Participation.SwitchMode @ auth && ParticipationWriteWindow && CanSwitch` | Доступно только на REGISTRATION, не в команде |
| `Participation.Unregister` | status | Отменить участие | `Participation.Unregister @ auth && ParticipationWriteWindow && CanUnregister` | Доступно только на REGISTRATION, не в команде |
| `Participation.ReadUser` | participation | Читать профиль участника | `Participation.ReadUser @ auth && ParticipationReadWindow && is_staff` | Доступно staff начиная с REGISTRATION |
| `Participation.ListParticipants` | participants list | Список участников | `Participation.ListParticipants @ auth && ParticipationReadWindow && is_staff` | Доступно staff начиная с REGISTRATION |
| `Participation.ConvertToTeam` | status, team_id | Конвертировать в команду | `Participation.ConvertToTeam @ service_to_service && status in {INDIVIDUAL_ACTIVE, LOOKING_FOR_TEAM}` | Внутренний вызов от Team Service |
| `Participation.ConvertFromTeam` | status, team_id | Конвертировать из команды | `Participation.ConvertFromTeam @ service_to_service && status in {TEAM_MEMBER, TEAM_CAPTAIN}` | Внутренний вызов от Team Service |

---

## 7) validation_errors

### 7.1 Минимальная структура
- `code`: `REQUIRED | FORBIDDEN | STAGE_RULE | CONFLICT | NOT_FOUND`
- `field`: имя поля или группы
- `message`: строка

### 7.2 Примеры
- `{code:"STAGE_RULE", field:"stage", message:"регистрация доступна только на стадии REGISTRATION"}`
- `{code:"FORBIDDEN", field:"status", message:"нельзя зарегистрироваться, имея staff роль"}`
- `{code:"FORBIDDEN", field:"status", message:"нельзя переключить режим, находясь в команде"}`
- `{code:"CONFLICT", field:"participation", message:"пользователь уже зарегистрирован на хакатон"}`
- `{code:"REQUIRED", field:"profile.wished_roles", message:"для LOOKING_FOR_TEAM рекомендуется указать желаемые роли"}`
