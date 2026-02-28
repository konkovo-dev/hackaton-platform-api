# Integration Testing Guide

Гайд по написанию интеграционных тестов для микросервисной архитектуры hackathon-platform-api.

## Оглавление

1. [Архитектура тестов](#архитектура-тестов)
2. [Настройка окружения](#настройка-окружения)
3. [Структура теста](#структура-теста)
4. [Хелперы и утилиты](#хелперы-и-утилиты)
5. [Паттерны тестирования](#паттерны-тестирования)
6. [Работа с БД](#работа-с-бд)
7. [Типичные ошибки](#типичные-ошибки)
8. [Примеры](#примеры)

---

## Архитектура тестов

### Принципы

- **E2E тесты**: Тесты работают через HTTP API (gateway), а не напрямую с gRPC сервисами
- **Изоляция**: Каждый тест создает свои данные (пользователи, хакатоны, команды)
- **Cleanup**: Тесты НЕ чистят за собой БД - используется флаг `-count=1` для отключения кеша
- **Реалистичность**: Тесты имитируют реальные пользовательские сценарии

### Структура файлов

```
tests/integration/
├── setup_test.go           # Общая инфраструктура (TestContext, хелперы)
├── auth_test.go            # Тесты auth-service
├── identity_test.go        # Тесты identity-service
├── hackathon_test.go       # Тесты hackathon-service
├── participation_test.go   # Тесты participation-service
├── team_test.go            # Тесты team-service
└── your_service_test.go    # Тесты вашего сервиса
```

---

## Настройка окружения

### Переменные окружения

Настройки в `.vscode/settings.json`:

```json
{
  "go.testEnvVars": {
    "API_BASE_URL": "http://localhost:8080",
    "DB_DSN": "postgres://hackathon:hackathon_dev_password@localhost:5432/hackathon?sslmode=disable"
  },
  "go.testFlags": ["-count=1"]
}
```

- `API_BASE_URL` - адрес gateway (HTTP)
- `DB_DSN` - подключение к БД для прямых SQL запросов (когда API недостаточно)
- `-count=1` - отключает кеширование результатов тестов

### Запуск тестов

```bash
# Запустить все тесты
go test -v ./tests/integration

# Запустить конкретный тест
go test -v -run TestCreateTeam_AsCaptain_ShouldCreateTeam ./tests/integration

# Запустить все тесты для одного сервиса
go test -v -run "^TestTeam" ./tests/integration
```

---

## Структура теста

### Базовый шаблон

```go
func TestFeatureName_Scenario_ExpectedOutcome(t *testing.T) {
    // 1. Setup: создать контекст и подготовить данные
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    
    // 2. Arrange: подготовить специфичные для теста данные
    hackathonID := createHackathon(tc, user)
    
    // 3. Act: выполнить тестируемое действие
    requestBody := map[string]interface{}{
        "name": "Test Team",
    }
    resp, body := tc.DoAuthenticatedRequest(
        "POST", 
        fmt.Sprintf("/v1/hackathons/%s/teams", hackathonID),
        user.AccessToken,
        requestBody,
    )
    
    // 4. Assert: проверить результат
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    data := tc.ParseJSON(body)
    assert.Equal(t, "Test Team", data["name"])
}
```

### Naming Convention

Формат: `Test<Feature>_<Scenario>_<ExpectedOutcome>`

**Примеры:**
- `TestCreateTeam_AsCaptain_ShouldCreateTeam` - позитивный сценарий
- `TestCreateTeam_AsNonParticipant_ShouldFail` - негативный сценарий
- `TestUpdateTeam_WithDuplicateName_ShouldFail` - граничный случай
- `TestDeleteTeam_SoleMember_ShouldDeleteAndConvertParticipation` - сложный сценарий

---

## Хелперы и утилиты

### TestContext

Основной объект для работы с тестами:

```go
type TestContext struct {
    BaseURL         string          // URL gateway
    HTTPClient      *http.Client    // HTTP клиент
    T               *testing.T      // Testing context
    DB              *pgxpool.Pool   // Прямое подключение к БД
    HackathonDBName string          // Имя таблицы хакатонов (зависит от окружения)
}
```

### Основные методы TestContext

#### 1. Регистрация пользователя

```go
user := tc.RegisterUser()
// Возвращает UserCredentials с AccessToken, UserID, Email, Password
```

#### 2. HTTP запросы

```go
// Без авторизации
resp, body := tc.DoRequest("GET", "/v1/ping", nil, nil)

// С авторизацией
resp, body := tc.DoAuthenticatedRequest(
    "POST",
    "/v1/hackathons",
    user.AccessToken,
    requestBody,
)
```

#### 3. Парсинг JSON

```go
data := tc.ParseJSON(body)
name := data["name"].(string)
id := data["id"].(string)

// Для вложенных структур
team := data["team"].(map[string]interface{})
teamName := team["name"].(string)
```

#### 4. Ожидание асинхронных операций

```go
// Ждать пока пользователю назначится роль
tc.WaitForHackathonOwnerRole(hackathonID, user.AccessToken)

// Общий паттерн ожидания
time.Sleep(500 * time.Millisecond)
```

---

## Паттерны тестирования

### 1. Создание тестовых данных через хелперы

Вместо дублирования кода создайте хелперы:

```go
// В конце файла your_service_test.go

func createTicket(tc *TestContext, hackathonID string, user *UserCredentials, subject string) string {
    body := map[string]interface{}{
        "subject": subject,
        "message": "Test message",
    }
    
    resp, respBody := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/tickets", hackathonID),
        user.AccessToken,
        body,
    )
    
    require.Equal(tc.T, http.StatusOK, resp.StatusCode, 
        "Failed to create ticket: %s", string(respBody))
    
    data := tc.ParseJSON(respBody)
    return data["ticketId"].(string)
}
```

**Использование:**

```go
func TestReplyToTicket_AsMentor_ShouldAddReply(t *testing.T) {
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    mentor := tc.RegisterUser()
    
    hackathonID := createHackathon(tc, user)
    ticketID := createTicket(tc, hackathonID, user, "Need help")
    
    // Теперь можно сфокусироваться на тестировании ответа
    replyBody := map[string]interface{}{
        "message": "Here's the answer",
    }
    resp, body := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s/replies", hackathonID, ticketID),
        mentor.AccessToken,
        replyBody,
    )
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### 2. Тестирование авторизации

Всегда проверяйте права доступа:

```go
func TestDeleteTicket_AsNonAuthor_ShouldFail(t *testing.T) {
    tc := NewTestContext(t)
    author := tc.RegisterUser()
    otherUser := tc.RegisterUser()
    
    hackathonID := createHackathon(tc, author)
    ticketID := createTicket(tc, hackathonID, author, "My ticket")
    
    // Попытка удалить чужой тикет
    resp, body := tc.DoAuthenticatedRequest(
        "DELETE",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s", hackathonID, ticketID),
        otherUser.AccessToken,  // ← Другой пользователь!
        nil,
    )
    
    assert.Equal(t, http.StatusForbidden, resp.StatusCode,
        "Should not allow deleting other user's ticket: %s", string(body))
}
```

### 3. Тестирование валидации

Проверяйте граничные случаи:

```go
func TestCreateTicket_EmptySubject_ShouldFail(t *testing.T) {
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    hackathonID := createHackathon(tc, user)
    
    body := map[string]interface{}{
        "subject": "",  // ← Пустая тема
        "message": "Test",
    }
    
    resp, respBody := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/tickets", hackathonID),
        user.AccessToken,
        body,
    )
    
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
        "Should reject empty subject: %s", string(respBody))
}

func TestCreateTicket_SubjectTooLong_ShouldFail(t *testing.T) {
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    hackathonID := createHackathon(tc, user)
    
    body := map[string]interface{}{
        "subject": strings.Repeat("a", 256),  // ← Слишком длинная тема
        "message": "Test",
    }
    
    resp, respBody := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/tickets", hackathonID),
        user.AccessToken,
        body,
    )
    
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
        "Should reject too long subject: %s", string(respBody))
}
```

### 4. Тестирование бизнес-логики

Проверяйте побочные эффекты:

```go
func TestAcceptInvitation_ShouldJoinTeamAndCancelOtherInvitations(t *testing.T) {
    tc := NewTestContext(t)
    captain1 := tc.RegisterUser()
    captain2 := tc.RegisterUser()
    user := tc.RegisterUser()
    
    hackathonID := createHackathon(tc, captain1)
    team1ID := createTeam(tc, hackathonID, captain1, "Team 1")
    team2ID := createTeam(tc, hackathonID, captain2, "Team 2")
    
    // Создать 2 приглашения
    invite1ID := createInvitation(tc, hackathonID, team1ID, captain1, user.UserID)
    invite2ID := createInvitation(tc, hackathonID, team2ID, captain2, user.UserID)
    
    // Принять первое
    resp, _ := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/users/me/team-invitations/%s/accept", invite1ID),
        user.AccessToken,
        nil,
    )
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    time.Sleep(500 * time.Millisecond)
    
    // Проверить что пользователь в команде 1
    resp, body := tc.DoAuthenticatedRequest(
        "GET",
        fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, team1ID),
        user.AccessToken,
        nil,
    )
    data := tc.ParseJSON(body)
    teamData := data["team"].(map[string]interface{})
    team := teamData["team"].(map[string]interface{})
    members := team["members"].([]interface{})
    assert.Len(t, members, 2, "Team should have 2 members")
    
    // Проверить что второе приглашение отменено
    resp, body = tc.DoAuthenticatedRequest(
        "GET",
        "/v1/users/me/team-invitations",
        user.AccessToken,
        nil,
    )
    data = tc.ParseJSON(body)
    invitations := data["invitations"].([]interface{})
    
    for _, inv := range invitations {
        invitation := inv.(map[string]interface{})
        if invitation["invitationId"].(string) == invite2ID {
            assert.Equal(t, "TEAM_INBOX_CANCELED", invitation["status"],
                "Second invitation should be canceled")
        }
    }
}
```

### 5. Тестирование списков и пагинации

```go
func TestListTickets_WithPagination_ShouldReturnCorrectPage(t *testing.T) {
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    hackathonID := createHackathon(tc, user)
    
    // Создать 5 тикетов
    for i := 1; i <= 5; i++ {
        createTicket(tc, hackathonID, user, fmt.Sprintf("Ticket %d", i))
    }
    
    // Запросить первую страницу (2 элемента)
    resp, body := tc.DoAuthenticatedRequest(
        "GET",
        fmt.Sprintf("/v1/hackathons/%s/tickets?page_size=2&page_token=", hackathonID),
        user.AccessToken,
        nil,
    )
    
    require.Equal(t, http.StatusOK, resp.StatusCode)
    data := tc.ParseJSON(body)
    tickets := data["tickets"].([]interface{})
    
    assert.Len(t, tickets, 2, "Should return 2 tickets")
    assert.NotEmpty(t, data["nextPageToken"], "Should have next page token")
    
    // Запросить вторую страницу
    nextToken := data["nextPageToken"].(string)
    resp, body = tc.DoAuthenticatedRequest(
        "GET",
        fmt.Sprintf("/v1/hackathons/%s/tickets?page_size=2&page_token=%s", hackathonID, nextToken),
        user.AccessToken,
        nil,
    )
    
    data = tc.ParseJSON(body)
    tickets = data["tickets"].([]interface{})
    assert.Len(t, tickets, 2, "Should return 2 more tickets")
}
```

---

## Работа с БД

### Когда использовать прямые SQL запросы

**Используйте SQL только когда:**
1. API не предоставляет нужную функциональность
2. Нужно обойти валидацию для тестирования граничных случаев
3. Нужно проверить состояние БД напрямую

**Примеры:**

#### 1. Обход валидации дат

```go
func TestCreateTeam_InRegistrationStage_ShouldSucceed(t *testing.T) {
    tc := NewTestContext(t)
    owner := tc.RegisterUser()
    captain := tc.RegisterUser()
    
    // Создать хакатон через API (он будет в статусе DRAFT)
    hackathonID := createHackathon(tc, owner)
    
    // Опубликовать хакатон
    resp, body := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/publish", hackathonID),
        owner.AccessToken,
        nil,
    )
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Напрямую в БД перевести хакатон в стадию REGISTRATION
    // (API не позволяет это сделать, т.к. проверяет даты)
    _, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
        UPDATE %s 
        SET registration_opens_at = $1,
            stage = 'registration'
        WHERE id = $2
    `, tc.HackathonDBName), time.Now().Add(-24*time.Hour), hackathonID)
    require.NoError(t, err)
    
    time.Sleep(500 * time.Millisecond)
    
    // Теперь можно создать команду
    registerParticipant(tc, hackathonID, captain, "PART_LOOKING_FOR_TEAM")
    teamID := createTeam(tc, hackathonID, captain, "Test Team")
    
    assert.NotEmpty(t, teamID)
}
```

#### 2. Проверка состояния БД

```go
func TestDeleteTeam_ShouldRemoveFromDatabase(t *testing.T) {
    tc := NewTestContext(t)
    captain := tc.RegisterUser()
    hackathonID := createHackathon(tc, captain)
    teamID := createTeam(tc, hackathonID, captain, "Test Team")
    
    // Удалить команду через API
    resp, _ := tc.DoAuthenticatedRequest(
        "DELETE",
        fmt.Sprintf("/v1/hackathons/%s/teams/%s", hackathonID, teamID),
        captain.AccessToken,
        nil,
    )
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Проверить что команда удалена из БД
    var count int
    err := tc.DB.QueryRow(context.Background(),
        "SELECT COUNT(*) FROM team.teams WHERE id = $1",
        teamID,
    ).Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 0, count, "Team should be deleted from database")
}
```

### Важно: Различия локальной и продовой БД

**Локальная БД** (docker-compose):
- Одна БД `hackathon` с разными схемами: `auth`, `identity`, `hackathon`, `participation`, `team`
- Таблицы: `hackathon.hackathons`, `team.teams`, и т.д.

**Продовая БД**:
- Разные БД для каждого сервиса: `hackathon_auth`, `hackathon_identity`, `hackathon_hackaton`, `hackathon_participation`, `hackathon_team`
- Таблицы без префикса схемы: `hackathons`, `teams`, и т.д.

**Решение:** Используйте `tc.HackathonDBName` для динамического определения:

```go
// ✅ Правильно - работает и локально, и на проде
_, err := tc.DB.Exec(context.Background(), fmt.Sprintf(`
    UPDATE %s 
    SET stage = 'registration'
    WHERE id = $1
`, tc.HackathonDBName), hackathonID)

// ❌ Неправильно - работает только локально
_, err := tc.DB.Exec(context.Background(), `
    UPDATE hackathon.hackathons 
    SET stage = 'registration'
    WHERE id = $1
`, hackathonID)
```

---

## Типичные ошибки

### 1. Забыли require.NoError после критичных операций

```go
// ❌ Плохо - тест продолжится даже если создание не удалось
resp, body := tc.DoAuthenticatedRequest("POST", "/v1/teams", user.AccessToken, body)
teamID := tc.ParseJSON(body)["teamId"].(string)  // Паника если resp != 200!

// ✅ Хорошо
resp, body := tc.DoAuthenticatedRequest("POST", "/v1/teams", user.AccessToken, body)
require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create team: %s", string(body))
teamID := tc.ParseJSON(body)["teamId"].(string)
```

### 2. Неправильный парсинг вложенных JSON структур

```go
// API возвращает: {"team": {"team": {...}, "vacancies": [...]}}

// ❌ Плохо - получим nil
data := tc.ParseJSON(body)
team := data["team"].(map[string]interface{})
name := team["name"].(string)  // Паника! name не на этом уровне

// ✅ Хорошо
data := tc.ParseJSON(body)
teamWithVacancies := data["team"].(map[string]interface{})
team := teamWithVacancies["team"].(map[string]interface{})
name := team["name"].(string)
```

### 3. Забыли time.Sleep после асинхронных операций

```go
// ❌ Плохо - проверка может выполниться до обновления
tc.DoAuthenticatedRequest("POST", "/v1/invitations/123/accept", user.AccessToken, nil)
resp, body := tc.DoAuthenticatedRequest("GET", "/v1/teams/456", user.AccessToken, nil)
// Может не увидеть нового участника!

// ✅ Хорошо
tc.DoAuthenticatedRequest("POST", "/v1/invitations/123/accept", user.AccessToken, nil)
time.Sleep(500 * time.Millisecond)  // Дать время на обработку
resp, body := tc.DoAuthenticatedRequest("GET", "/v1/teams/456", user.AccessToken, nil)
```

### 4. Неправильное сравнение типов из JSON

```go
data := tc.ParseJSON(body)

// ❌ Плохо - JSON парсит числа как float64
assert.Equal(t, 2, data["count"])  // Упадет! Ожидается int, получен float64

// ✅ Хорошо
assert.Equal(t, float64(2), data["count"])

// Или еще лучше - если API возвращает строку
assert.Equal(t, "2", data["count"])
```

### 5. Не проверили статус код перед парсингом

```go
// ❌ Плохо - паника если resp != 200
resp, body := tc.DoAuthenticatedRequest("POST", "/v1/teams", user.AccessToken, body)
data := tc.ParseJSON(body)  // Паника если body содержит ошибку!

// ✅ Хорошо
resp, body := tc.DoAuthenticatedRequest("POST", "/v1/teams", user.AccessToken, body)
if resp.StatusCode != http.StatusOK {
    t.Fatalf("Failed to create team: %s", string(body))
}
data := tc.ParseJSON(body)
```

---

## Примеры

### Пример 1: Простой CRUD тест

```go
func TestCreateTicket_ValidData_ShouldCreateTicket(t *testing.T) {
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    hackathonID := createHackathon(tc, user)
    
    body := map[string]interface{}{
        "subject": "Need help with API",
        "message": "How do I authenticate?",
    }
    
    resp, respBody := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/tickets", hackathonID),
        user.AccessToken,
        body,
    )
    
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    data := tc.ParseJSON(respBody)
    ticketID := data["ticketId"].(string)
    
    assert.NotEmpty(t, ticketID)
    assert.Equal(t, "Need help with API", data["subject"])
    assert.Equal(t, "open", data["status"])
}
```

### Пример 2: Тест с проверкой авторизации

```go
func TestGetTicket_AsAuthor_ShouldReturnTicket(t *testing.T) {
    tc := NewTestContext(t)
    author := tc.RegisterUser()
    otherUser := tc.RegisterUser()
    
    hackathonID := createHackathon(tc, author)
    ticketID := createTicket(tc, hackathonID, author, "My ticket")
    
    // Автор может видеть свой тикет
    resp, body := tc.DoAuthenticatedRequest(
        "GET",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s", hackathonID, ticketID),
        author.AccessToken,
        nil,
    )
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    data := tc.ParseJSON(body)
    assert.Equal(t, ticketID, data["ticketId"])
    
    // Другой пользователь не может видеть чужой тикет
    resp, body = tc.DoAuthenticatedRequest(
        "GET",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s", hackathonID, ticketID),
        otherUser.AccessToken,
        nil,
    )
    
    assert.Equal(t, http.StatusForbidden, resp.StatusCode,
        "Other user should not see the ticket: %s", string(body))
}
```

### Пример 3: Тест со сложной бизнес-логикой

```go
func TestCloseTicket_WithOpenReplies_ShouldMarkRepliesAsRead(t *testing.T) {
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    mentor := tc.RegisterUser()
    
    hackathonID := createHackathon(tc, user)
    assignMentorRole(tc, hackathonID, mentor)
    
    ticketID := createTicket(tc, hackathonID, user, "Need help")
    
    // Ментор отвечает
    replyBody := map[string]interface{}{
        "message": "Here's the answer",
    }
    resp, body := tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s/replies", hackathonID, ticketID),
        mentor.AccessToken,
        replyBody,
    )
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    time.Sleep(500 * time.Millisecond)
    
    // Пользователь закрывает тикет
    resp, body = tc.DoAuthenticatedRequest(
        "POST",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s/close", hackathonID, ticketID),
        user.AccessToken,
        nil,
    )
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    time.Sleep(500 * time.Millisecond)
    
    // Проверить что тикет закрыт
    resp, body = tc.DoAuthenticatedRequest(
        "GET",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s", hackathonID, ticketID),
        user.AccessToken,
        nil,
    )
    
    data := tc.ParseJSON(body)
    assert.Equal(t, "closed", data["status"])
    
    // Проверить что все ответы помечены как прочитанные
    replies := data["replies"].([]interface{})
    for _, r := range replies {
        reply := r.(map[string]interface{})
        assert.True(t, reply["isRead"].(bool), "All replies should be marked as read")
    }
}
```

### Пример 4: Тест с использованием БД

```go
func TestDeleteTicket_ShouldDeleteReplies(t *testing.T) {
    tc := NewTestContext(t)
    user := tc.RegisterUser()
    mentor := tc.RegisterUser()
    
    hackathonID := createHackathon(tc, user)
    ticketID := createTicket(tc, hackathonID, user, "Test")
    
    // Создать несколько ответов
    for i := 0; i < 3; i++ {
        createReply(tc, hackathonID, ticketID, mentor, fmt.Sprintf("Reply %d", i))
    }
    
    // Удалить тикет
    resp, _ := tc.DoAuthenticatedRequest(
        "DELETE",
        fmt.Sprintf("/v1/hackathons/%s/tickets/%s", hackathonID, ticketID),
        user.AccessToken,
        nil,
    )
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Проверить что ответы тоже удалены из БД
    var replyCount int
    err := tc.DB.QueryRow(context.Background(),
        "SELECT COUNT(*) FROM mentors.ticket_replies WHERE ticket_id = $1",
        ticketID,
    ).Scan(&replyCount)
    require.NoError(t, err)
    assert.Equal(t, 0, replyCount, "All replies should be deleted")
}
```

---

## Чеклист перед коммитом

- [ ] Все тесты проходят локально: `go test -v ./tests/integration`
- [ ] Названия тестов следуют конвенции: `Test<Feature>_<Scenario>_<ExpectedOutcome>`
- [ ] Используются `require` для критичных проверок, `assert` для некритичных
- [ ] Добавлены сообщения к ассертам: `assert.Equal(t, expected, actual, "Helpful message: %s", context)`
- [ ] Тесты изолированы - каждый создает свои данные
- [ ] Проверены как позитивные, так и негативные сценарии
- [ ] Проверена авторизация (кто может/не может выполнить действие)
- [ ] Добавлены `time.Sleep()` после асинхронных операций
- [ ] SQL запросы используют `tc.HackathonDBName` вместо хардкода
- [ ] Хелперы вынесены в отдельные функции для переиспользования

---

## Полезные ссылки

- [Testify documentation](https://github.com/stretchr/testify)
- [Table-driven tests in Go](https://go.dev/wiki/TableDrivenTests)
- Примеры тестов: `tests/integration/team_test.go`

---

**Удачи в написании тестов! 🚀**
