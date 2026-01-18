# Hackathon Policy Spec — Mentors domain (v1, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Role/Participation service)
- `roles[]`: `{OWNER, ORGANIZER, MENTOR, JURY}`
- `particip.kind`: `{NONE, LOOKING_FOR_TEAM, SINGLE, TEAM}`
- `particip.team_id`: `uuid | null`

### 0.2 Hackathon context (Hackathon service)
- `stage ∈ {DRAFT, UPCOMING, REGISTRATION, PRESTART, RUNNING, JUDGING, FINISHED}`

### 0.3 Team context (Team service)
- `is_team_member(team_id, actor.user_id)` (bool)

### 0.4 Mentors context (Mentors service)

`SupportThread`:
- `thread_id`
- `hackathon_id`
- `scope ∈ {TEAM, SINGLE}`
- `team_id` (uuid | null)
- `user_id` (uuid | null)
- `is_open` (bool)

`Ticket`:
- `ticket_id`
- `thread_id`
- `hackathon_id`
- `status ∈ {OPEN, CLOSED}`
- `assigned_mentor_user_id` (uuid | null)
- `created_at`
- `closed_at` (timestamp | null)

`Message`:
- `message_id`
- `ticket_id`
- `author_kind ∈ {PARTICIPANT, MENTOR, ORGANIZER}`
- `author_user_id`
- `text`
- `created_at`

`ChatView`:
- единая лента сообщений для участника в рамках `SupportThread`
- сообщения агрегируются по всем тикетам внутри `SupportThread` в порядке `created_at`

---

## 1) Предикаты

### 1.1 Окно стадий
- `MentorsWindow`: `stage == RUNNING`

### 1.2 Роли
- `is_owner_or_organizer`: `role in {OWNER, ORGANIZER}`
- `is_mentor`: `role in {MENTOR}`

### 1.3 Идентификация участника
- `is_team_participant`: `particip.kind == TEAM && particip.team_id != null`
- `is_single_participant`: `particip.kind == SINGLE`

### 1.4 Доступ к thread
- `is_my_thread(thread)`:
  - `thread.scope == TEAM && particip.kind == TEAM && thread.team_id == particip.team_id`
  - OR `thread.scope == SINGLE && particip.kind == SINGLE && thread.user_id == actor.user_id`

### 1.5 Доступ к ticket
- `is_my_ticket(ticket)`: `ticket.thread_id принадлежит thread, где is_my_thread(thread) == true`

### 1.6 Назначение ментора
- `is_assigned_mentor(ticket)`: `ticket.assigned_mentor_user_id == actor.user_id`

---

## 2) Основные правила модели

### 2.1 Обезличенный чат
- для одного `team_id` существует не более одного открытого `SupportThread` со `scope == TEAM`
- для одного `user_id` существует не более одного открытого `SupportThread` со `scope == SINGLE`

### 2.2 Тикеты
- каждое обращение участника создаёт новый `Ticket` в пределах `SupportThread`
- переписка ведётся сообщениями `Message` внутри `Ticket`
- тикет закрывается ментором или организатором

### 2.3 Пользовательский чат
- пользователь видит один чат на `SupportThread`
- если пользователь отправляет сообщение и нет открытого тикета, создаётся новый тикет и сообщение кладётся в него

---

## 3) Политика чтения

### 3.1 Mentors.ReadMyThreads
`Mentors.ReadMyThreads @ auth && MentorsWindow && (particip.kind in {SINGLE, TEAM})`

Описание:
- участник читает свои треды (в v1 обычно будет один открытый)

### 3.2 Mentors.ReadMyTickets
`Mentors.ReadMyTickets @ auth && MentorsWindow && (particip.kind in {SINGLE, TEAM})`

Описание:
- участник читает свои тикеты и их статусы

### 3.3 Mentors.ReadTicketMessages.My
`Mentors.ReadTicketMessages.My @ auth && MentorsWindow && (particip.kind in {SINGLE, TEAM}) && is_my_ticket(ticket)`

Описание:
- участник читает сообщения только своего тикета

### 3.3a Mentors.ReadChat.My
`Mentors.ReadChat.My @ auth && MentorsWindow && (particip.kind in {SINGLE, TEAM})`

Описание:
- участник читает единую ленту сообщений своего `SupportThread`
- в ответе возвращаются сообщения всех тикетов этого треда в порядке `created_at`

### 3.4 Mentors.ReadAssignedTickets
`Mentors.ReadAssignedTickets @ auth && MentorsWindow && is_mentor`

Описание:
- ментор читает список тикетов, назначенных ему

### 3.5 Mentors.ReadTicketMessages.Mentor
`Mentors.ReadTicketMessages.Mentor @ auth && MentorsWindow && is_mentor && is_assigned_mentor(ticket)`

Описание:
- ментор читает сообщения назначенного ему тикета

### 3.6 Mentors.ReadAllTickets
`Mentors.ReadAllTickets @ auth && MentorsWindow && is_owner_or_organizer`

Описание:
- OWNER/ORGANIZER читают все тикеты

### 3.7 Mentors.ReadTicketMessages.All
`Mentors.ReadTicketMessages.All @ auth && MentorsWindow && is_owner_or_organizer`

Описание:
- OWNER/ORGANIZER читают сообщения любого тикета

---

## 4) Политика записи

### 4.1 Chat.SendMessage.ByParticipant
`Chat.SendMessage.ByParticipant @ auth && MentorsWindow && (is_team_participant || is_single_participant)`

Описание:
- участник отправляет сообщение в свой чат
- если `SupportThread` отсутствует, он создаётся
- если нет открытого тикета, создаётся новый `Ticket` со статусом `OPEN`
- новый тикет назначается случайному ментору из пула доступных
- сообщение сохраняется в открытом тикете

Ограничения:
- действие запрещено при `particip.kind in {NONE, LOOKING_FOR_TEAM}`

### 4.2 Ticket.CloseByMentor
`Ticket.CloseByMentor @ auth && MentorsWindow && is_mentor && is_assigned_mentor(ticket) && ticket.status == OPEN`

Описание:
- ментор закрывает назначенный ему тикет

### 4.3 Ticket.CloseByOrganizer
`Ticket.CloseByOrganizer @ auth && MentorsWindow && is_owner_or_organizer && ticket.status == OPEN`

Описание:
- OWNER/ORGANIZER закрывают любой тикет

### 4.4 Message.SendByParticipant
`Message.SendByParticipant @ auth && MentorsWindow && (particip.kind in {SINGLE, TEAM}) && is_my_ticket(ticket) && ticket.status == OPEN`

Описание:
- участник отправляет сообщение в свой тикет

Примечание:
- в пользовательском API рекомендуется использовать `Chat.SendMessage.ByParticipant`

### 4.5 Message.SendByMentor
`Message.SendByMentor @ auth && MentorsWindow && is_mentor && is_assigned_mentor(ticket) && ticket.status == OPEN`

Описание:
- ментор отправляет сообщение в назначенный ему тикет

### 4.6 Message.SendByOrganizer
`Message.SendByOrganizer @ auth && MentorsWindow && is_owner_or_organizer && ticket.status == OPEN`

Описание:
- OWNER/ORGANIZER отправляют сообщение в любой тикет

---

## 5) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `Mentors.ReadChat.My` | messages | Читать чат | `Mentors.ReadChat.My @ auth && MentorsWindow && particip.kind in {SINGLE, TEAM}` | Доступно участникам в RUNNING |
| `Mentors.ReadMyThreads` | threads | Читать свои треды | `Mentors.ReadMyThreads @ auth && MentorsWindow && particip.kind in {SINGLE, TEAM}` | Доступно участникам в RUNNING |
| `Mentors.ReadMyTickets` | tickets | Читать свои тикеты | `Mentors.ReadMyTickets @ auth && MentorsWindow && particip.kind in {SINGLE, TEAM}` | Доступно участникам в RUNNING |
| `Mentors.ReadTicketMessages.My` | messages | Читать сообщения своего тикета | `Mentors.ReadTicketMessages.My @ auth && MentorsWindow && particip.kind in {SINGLE, TEAM} && is_my_ticket(ticket)` | Доступно участникам в RUNNING |
| `Mentors.ReadAssignedTickets` | tickets | Читать назначенные тикеты | `Mentors.ReadAssignedTickets @ auth && MentorsWindow && is_mentor` | Доступно менторам в RUNNING |
| `Mentors.ReadTicketMessages.Mentor` | messages | Читать сообщения назначенного тикета | `Mentors.ReadTicketMessages.Mentor @ auth && MentorsWindow && is_mentor && is_assigned_mentor(ticket)` | Доступно менторам в RUNNING |
| `Mentors.ReadAllTickets` | tickets | Читать все тикеты | `Mentors.ReadAllTickets @ auth && MentorsWindow && is_owner_or_organizer` | Доступно OWNER/ORGANIZER в RUNNING |
| `Mentors.ReadTicketMessages.All` | messages | Читать сообщения любого тикета | `Mentors.ReadTicketMessages.All @ auth && MentorsWindow && is_owner_or_organizer` | Доступно OWNER/ORGANIZER в RUNNING |
| `Chat.SendMessage.ByParticipant` | message | Отправить сообщение в чат | `Chat.SendMessage.ByParticipant @ auth && MentorsWindow && (is_team_participant || is_single_participant)` | Доступно участникам в RUNNING |
| `Ticket.CloseByMentor` | ticket | Закрыть тикет | `Ticket.CloseByMentor @ auth && MentorsWindow && is_mentor && is_assigned_mentor(ticket) && ticket.status == OPEN` | Доступно назначенному ментору |
| `Ticket.CloseByOrganizer` | ticket | Закрыть тикет | `Ticket.CloseByOrganizer @ auth && MentorsWindow && is_owner_or_organizer && ticket.status == OPEN` | Доступно OWNER/ORGANIZER |
| `Message.SendByParticipant` | message | Отправить сообщение | `Message.SendByParticipant @ auth && MentorsWindow && particip.kind in {SINGLE, TEAM} && is_my_ticket(ticket) && ticket.status == OPEN` | Доступно участникам в RUNNING |
| `Message.SendByMentor` | message | Отправить сообщение | `Message.SendByMentor @ auth && MentorsWindow && is_mentor && is_assigned_mentor(ticket) && ticket.status == OPEN` | Доступно назначенному ментору |
| `Message.SendByOrganizer` | message | Отправить сообщение | `Message.SendByOrganizer @ auth && MentorsWindow && is_owner_or_organizer && ticket.status == OPEN` | Доступно OWNER/ORGANIZER |

---

## 6) validation_errors

### 6.1 Минимальная структура
- `code`: `REQUIRED | FORBIDDEN | STAGE_RULE | NOT_FOUND | CONFLICT`
- `field`: имя поля или группы
- `message`: строка

### 6.2 Примеры
- `{code:"STAGE_RULE", field:"stage", message:"операция разрешена только в RUNNING"}`
- `{code:"FORBIDDEN", field:"particip.kind", message:"доступно только участникам хакатона"}`
- `{code:"FORBIDDEN", field:"ticket_id", message:"нет доступа к тикету"}`
- `{code:"CONFLICT", field:"ticket.status", message:"тикет закрыт"}`
- `{code:"NOT_FOUND", field:"ticket_id", message:"тикет не найден"}`
