# Identity Policy Spec — Users domain (v1, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Auth)
- `auth`: пользователь авторизован
- `actor.user_id`: идентификатор пользователя

### 0.2 Target context (Identity service)
- `target.user_id`: идентификатор профиля, к которому идёт обращение
- `is_me`: `actor.user_id == target.user_id`

### 0.3 Поля профиля (Identity service)
- `username` (string)
- `first_name` (string)
- `last_name` (string)
- `timezone` (string)
- `avatar_url` (string)
- `skills` (list<struct>)
- `skills_visibility` (bool)
- `contacts_visibility` (bool)
- `contacts` (list<Contact>), где `Contact` включает `per_contact_visibility` (bool)

---

## 1) Словарь предикатов

### 1.1 Актор и принадлежность профиля
- `auth`
- `is_me`

### 1.2 Видимость
- `skills_visibility == true`
- `contacts_visibility == true`
- `contact.per_contact_visibility == true`

---

## 2) Режимы времени
Правила Identity не зависят от времени.

---

## 3) Политика чтения (фильтрация полей)

### 3.1 Read: авторизованный + свой профиль
`Identity.ReadMe @ auth && is_me`

Поля в ответе:
- `username`
- `first_name`
- `last_name`
- `timezone`
- `avatar_url`
- `skills`
- `skills_visibility`
- `contacts_visibility`
- `contacts` (включая `per_contact_visibility`)

### 3.2 Read: авторизованный + чужой профиль
`Identity.ReadUser @ auth && !is_me`

Поля в ответе:
- всегда:
  - `username`
  - `first_name`
  - `last_name`
  - `timezone`
  - `avatar_url`
- условно:
  - `skills` — только если `skills_visibility == true`
  - `contacts` — только если `contacts_visibility == true`, и внутри только те элементы `contacts`, у которых `per_contact_visibility == true`

Поля, которые не возвращаются:
- `skills_visibility`
- `contacts_visibility`
- `contacts.per_contact_visibility`

### 3.3 Read: неавторизованный
`Identity.ReadUser @ !auth`

Поля в ответе:
- не возвращаются (доступ запрещён)

---

## 4) Политика записи (CRUD)

> Оговорка про `C`:
> - `C` для `username/email/first_name/last_name/timezone/avatar_url` реализуется процессом регистрации.
> - `skills` и `contacts` не заполняются в регистрации и изменяются отдельными методами после регистрации.

### 4.1 Регистрация (создание профиля)
`Identity.Register @ auth`

Создаёт профиль и устанавливает:
- `username` (C)
- `email` (C)
- `first_name` (C)
- `last_name` (C)
- `timezone` (C)
- `avatar_url` (C, опционально)

### 4.2 Обновление своего профиля (основные поля)
`Identity.UpdateMe.Profile @ auth && is_me`

Разрешены изменения:
- `first_name` (U)
- `last_name` (U)
- `timezone` (U)
- `avatar_url` (U/D)

Запрещены изменения:
- `username` (immutable после регистрации)
- `email` (immutable после регистрации. не является частью identity, хранится только в auth)

### 4.3 Обновление skills
`Identity.UpdateMe.Skills @ auth && is_me`

Разрешены изменения:
- `skills` (CRUD)
- `skills_visibility` (CRUD)

### 4.4 Обновление contacts
`Identity.UpdateMe.Contacts @ auth && is_me`

Разрешены изменения:
- `contacts` (CRUD)
- `contacts_visibility` (CRUD)
- `contacts.per_contact_visibility` (CRUD)

---

## 5) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `Identity.ReadMe` | все поля профиля | Читать свой профиль | `Identity.ReadMe @ auth && is_me` | Доступно только владельцу профиля |
| `Identity.ReadUser` | публичные поля + условные поля | Читать чужой профиль | `Identity.ReadUser @ auth && !is_me` | Доступно любому авторизованному пользователю; чувствительные поля фильтруются |
| `Identity.ReadUser.Anonymous` | — | Читать профиль без авторизации | `Identity.ReadUser.Anonymous @ !auth` | Доступ запрещён |
| `Identity.Register` | `username,email,first_name,last_name,timezone,avatar_url` | Зарегистрировать профиль | `Identity.Register @ auth` | Создание профиля происходит в процессе регистрации |
| `Identity.UpdateMe.Profile` | `first_name,last_name,timezone,avatar_url` | Обновить основные поля профиля | `Identity.UpdateMe.Profile @ auth && is_me` | Изменение разрешено только владельцу; username/email неизменяемы |
| `Identity.UpdateMe.Skills` | `skills,skills_visibility` | Обновить навыки и их видимость | `Identity.UpdateMe.Skills @ auth && is_me` | Доступно только владельцу |
| `Identity.UpdateMe.Contacts` | `contacts,contacts_visibility,contacts.per_contact_visibility` | Обновить контакты и их видимость | `Identity.UpdateMe.Contacts @ auth && is_me` | Доступно только владельцу |

---

## 6) Правила фильтрации (детализация)

### 6.1 Skills при чтении чужого профиля
Если `skills_visibility == false`, поле `skills` не возвращается.

### 6.2 Contacts при чтении чужого профиля
Если `contacts_visibility == false`, поле `contacts` не возвращается.
Если `contacts_visibility == true`, возвращаются только элементы, где `per_contact_visibility == true`.
Поле `per_contact_visibility` в ответе не возвращается.

---

## 7) validation_errors
Для Identity строгие ошибки зависят от форматов полей (валидации не завязаны на время).

Рекомендуемые коды:
- `REQUIRED`
- `FORMAT`
- `CONFLICT`
- `FORBIDDEN`

Примеры:
- `{code:"REQUIRED", field:"username", message:"username обязателен"}`
- `{code:"FORMAT", field:"timezone", message:"timezone должен быть валидным IANA идентификатором"}`
- `{code:"CONFLICT", field:"username", message:"username уже занят"}`
- `{code:"FORBIDDEN", field:"email", message:"email нельзя изменить после регистрации"}`
