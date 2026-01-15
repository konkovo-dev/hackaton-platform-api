# SkillsService REST Guide

## Описание

`SkillsService` предоставляет REST API для работы с каталогом навыков.

**Требует авторизации**: ✅ Все методы требуют Bearer token.

**Base URL:** `http://localhost:8080`

---

## Предусловия

### Получение access_token

Для тестирования `SkillsService` необходим access token. Используйте регистрацию через `AuthService`:

```bash
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser-skills",
    "email": "testuser-skills@example.com",
    "password": "SecurePass123",
    "first_name": "Test",
    "last_name": "Skills",
    "timezone": "UTC"
  }')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken')
echo "ACCESS_TOKEN=$ACCESS_TOKEN"
```

Сохраните токен в переменную для дальнейшего использования.

---

## Endpoints

### ListSkillCatalog

**POST** `/v1/skills:list`

Получить список навыков из каталога с поддержкой фильтрации, сортировки и пагинации.

**Поддерживаемые параметры:**

- `query.q` — поиск по имени навыка (case-insensitive, ILIKE %value%)
- `query.filter_groups` — фильтрация:
  - `name` (CONTAINS — точное совпадение, PREFIX — префиксный поиск)
- `query.sort` — сортировка по `name` (ASC/DESC), по умолчанию `name ASC`
- `query.page.page_size` — размер страницы (по умолчанию 50, максимум 100)
- `query.page.page_token` — токен для следующей страницы

---

## Тест-сценарии

> **Важно**: Все запросы требуют авторизации. Используйте `ACCESS_TOKEN` из раздела выше.

### 1. ListSkillCatalog (базовый запрос)

```bash
curl -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "page": {
        "page_size": 10
      }
    }
  }' | jq .
```

**Response:**
```json
{
  "skills": [
    {
      "id": "00000000-0000-0000-0000-000000000014",
      "name": "Angular"
    },
    {
      "id": "00000000-0000-0000-0000-000000000032",
      "name": "AWS"
    },
    {
      "id": "00000000-0000-0000-0000-000000000033",
      "name": "Azure"
    },
    {
      "id": "00000000-0000-0000-0000-000000000007",
      "name": "C++"
    },
    {
      "id": "00000000-0000-0000-0000-000000000035",
      "name": "CI/CD"
    },
    {
      "id": "00000000-0000-0000-0000-000000000042",
      "name": "Data Science"
    },
    {
      "id": "00000000-0000-0000-0000-000000000030",
      "name": "Docker"
    },
    {
      "id": "00000000-0000-0000-0000-000000000024",
      "name": "Elasticsearch"
    },
    {
      "id": "00000000-0000-0000-0000-000000000034",
      "name": "GCP"
    },
    {
      "id": "00000000-0000-0000-0000-000000000040",
      "name": "Git"
    }
  ],
  "page": {
    "nextPageToken": "eyJuYW1lIjoiR2l0IiwiaWQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwNDAifQ=="
  }
}
```

> **Примечание:** Навыки отсортированы по имени (ASC). Для получения следующей страницы используйте `nextPageToken`.

### 2. ListSkillCatalog с пагинацией

```bash
curl -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "page": {
        "page_size": 10,
        "page_token": "eyJuYW1lIjoiR2l0IiwiaWQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwNDAifQ=="
      }
    }
  }' | jq .
```

**Response:**
```json
{
  "skills": [
    {
      "id": "00000000-0000-0000-0000-000000000001",
      "name": "Go"
    },
    {
      "id": "00000000-0000-0000-0000-000000000043",
      "name": "GraphQL"
    },
    {
      "id": "00000000-0000-0000-0000-000000000044",
      "name": "gRPC"
    },
    {
      "id": "00000000-0000-0000-0000-000000000003",
      "name": "Java"
    },
    {
      "id": "00000000-0000-0000-0000-000000000010",
      "name": "JavaScript"
    },
    {
      "id": "00000000-0000-0000-0000-000000000031",
      "name": "Kubernetes"
    },
    {
      "id": "00000000-0000-0000-0000-000000000041",
      "name": "Machine Learning"
    },
    {
      "id": "00000000-0000-0000-0000-000000000045",
      "name": "Microservices"
    },
    {
      "id": "00000000-0000-0000-0000-000000000021",
      "name": "MongoDB"
    },
    {
      "id": "00000000-0000-0000-0000-000000000023",
      "name": "MySQL"
    }
  ],
  "page": {
    "nextPageToken": "eyJuYW1lIjoiTXlTUUwiLCJpZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAyMyJ9"
  }
}
```

### 3. ListSkillCatalog с поиском (query.q)

```bash
# Поиск навыков, содержащих "script"
curl -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "q": "script",
      "page": {
        "page_size": 10
      }
    }
  }' | jq .
```

**Response:**
```json
{
  "skills": [
    {
      "id": "00000000-0000-0000-0000-000000000010",
      "name": "JavaScript"
    },
    {
      "id": "00000000-0000-0000-0000-000000000011",
      "name": "TypeScript"
    }
  ],
  "page": {}
}
```

> **Примечание:** `query.q` выполняет поиск (ILIKE '%value%') по имени навыка.

### 4. ListSkillCatalog с фильтрацией (CONTAINS — точное совпадение)

```bash
# Найти навык с именем "Go" (точное совпадение, case-insensitive)
curl -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "name",
              "operation": "FILTER_OPERATION_CONTAINS",
              "string_value": "go"
            }
          ]
        }
      ],
      "page": {
        "page_size": 10
      }
    }
  }' | jq .
```

**Response:**
```json
{
  "skills": [
    {
      "id": "00000000-0000-0000-0000-000000000001",
      "name": "Go"
    }
  ],
  "page": {}
}
```

> **Примечание:** `FILTER_OPERATION_CONTAINS` для `name` выполняет **точное совпадение** (case-insensitive).

### 5. ListSkillCatalog с фильтрацией (PREFIX)

```bash
# Найти навыки, начинающиеся с "Ja"
curl -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "name",
              "operation": "FILTER_OPERATION_PREFIX",
              "string_value": "Ja"
            }
          ]
        }
      ],
      "page": {
        "page_size": 10
      }
    }
  }' | jq .
```

**Response:**
```json
{
  "skills": [
    {
      "id": "00000000-0000-0000-0000-000000000003",
      "name": "Java"
    },
    {
      "id": "00000000-0000-0000-0000-000000000010",
      "name": "JavaScript"
    }
  ],
  "page": {}
}
```

### 6. ListSkillCatalog с сортировкой DESC

```bash
curl -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "sort": [
        {
          "field": "name",
          "direction": "SORT_DIRECTION_DESC"
        }
      ],
      "page": {
        "page_size": 5
      }
    }
  }' | jq .
```

**Response:**
```json
{
  "skills": [
    {
      "id": "00000000-0000-0000-0000-000000000013",
      "name": "Vue.js"
    },
    {
      "id": "00000000-0000-0000-0000-000000000046",
      "name": "UI/UX Design"
    },
    {
      "id": "00000000-0000-0000-0000-000000000011",
      "name": "TypeScript"
    },
    {
      "id": "00000000-0000-0000-0000-000000000036",
      "name": "Terraform"
    },
    {
      "id": "00000000-0000-0000-0000-000000000008",
      "name": "Rust"
    }
  ],
  "page": {
    "nextPageToken": "eyJuYW1lIjoiUnVzdCIsImlkIjoiMDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDA4In0="
  }
}
```

---

## Полный тестовый сценарий

Сохраните как `test-skills-service-rest.sh`:

```bash
#!/bin/bash

echo "=== SkillsService REST Test ==="

# 0. Регистрация
echo -e "\n0. Registering test user..."
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser-skills",
    "email": "testuser-skills@example.com",
    "password": "SecurePass123",
    "first_name": "Test",
    "last_name": "Skills",
    "timezone": "UTC"
  }')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken')
echo "Got access token: ${ACCESS_TOKEN:0:20}..."

# 1. ListSkillCatalog (базовый)
echo -e "\n1. ListSkillCatalog (базовый)..."
curl -s -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "page": {
        "page_size": 10
      }
    }
  }' | jq .

# 2. ListSkillCatalog с поиском
echo -e "\n2. ListSkillCatalog с поиском 'script'..."
curl -s -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "q": "script",
      "page": {
        "page_size": 10
      }
    }
  }' | jq .

# 3. ListSkillCatalog с фильтрацией (CONTAINS)
echo -e "\n3. ListSkillCatalog с фильтрацией (точное совпадение 'Go')..."
curl -s -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "name",
              "operation": "FILTER_OPERATION_CONTAINS",
              "string_value": "go"
            }
          ]
        }
      ],
      "page": {
        "page_size": 10
      }
    }
  }' | jq .

# 4. ListSkillCatalog с фильтрацией (PREFIX)
echo -e "\n4. ListSkillCatalog с префиксом 'Ja'..."
curl -s -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "filter_groups": [
        {
          "filters": [
            {
              "field": "name",
              "operation": "FILTER_OPERATION_PREFIX",
              "string_value": "Ja"
            }
          ]
        }
      ],
      "page": {
        "page_size": 10
      }
    }
  }' | jq .

# 5. ListSkillCatalog с сортировкой DESC
echo -e "\n5. ListSkillCatalog с сортировкой DESC..."
curl -s -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "sort": [
        {
          "field": "name",
          "direction": "SORT_DIRECTION_DESC"
        }
      ],
      "page": {
        "page_size": 5
      }
    }
  }' | jq .

echo -e "\n=== All tests completed ==="
```

Запустите:
```bash
chmod +x test-skills-service-rest.sh
./test-skills-service-rest.sh
```

---

## Troubleshooting

### Ошибка: 401 Unauthorized

```json
{
  "code": 16,
  "message": "missing or invalid token",
  "details": []
}
```

**Решение:** Убедитесь, что вы передали Bearer token через header:
```bash
-H "Authorization: Bearer $ACCESS_TOKEN"
```

### Ошибка: 400 Bad Request (InvalidArgument)

```json
{
  "code": 3,
  "message": "invalid input: ...",
  "details": []
}
```

**Возможные причины:**
- Неправильная операция для поля (`name` поддерживает только CONTAINS/PREFIX)
- Пустое значение в фильтре
- Некорректный `page_token`

---

## Полезные команды

```bash
# Получить все навыки (без пагинации)
curl -s -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": {"page": {"page_size": 100}}}' | jq '.skills[] | {id, name}'

# Посчитать количество навыков
curl -s -X POST "http://localhost:8080/v1/skills:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": {"page": {"page_size": 100}}}' | jq '.skills | length'
```

