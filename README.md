# kaiten-cli

Консольный клиент для [Kaiten](https://kaiten.ru/) — системы управления проектами. Позволяет управлять пространствами, досками, колонками, дорожками и карточками через командную строку.

По умолчанию команды возвращают JSON. Для `cards list`, `search cards` и `sync` доступны форматы `json`, `jsonl`, `table`, `yaml`, `csv`; JSONL удобен для больших выгрузок и AI-agent workflows. `--fields`, `--no-descriptions` и `--quiet/-q` поддерживаются в карточных командах `cards list` и `search cards`.

## Требования

- Go 1.23+
- Аккаунт Kaiten с API-токеном

## Установка

### Сборка из исходников

```bash
git clone https://gitlab.life-pay.ru/ai/kaiten-cli.git
cd kaiten-cli
make build
```

Бинарник появится в `build/kaiten`.

### Кросс-платформенная сборка

```bash
make cross-build
```

Собирает под:
- Linux amd64 / arm64
- macOS amd64 / arm64
- Windows amd64

## Настройка

Конфигурация через переменные окружения:

| Переменная | Описание | Пример |
|---|---|---|
| `KAITEN_API_TOKEN` | API-токен Kaiten | `Токен из настроек Kaiten` |
| `KAITEN_URL` | URL вашего Kaiten | `https://mycompany.kaiten.ru` |
| `KAITEN_DB_PATH` | Путь к локальной SQLite БД | `~/.kaiten/kaiten.db` |

```bash
export KAITEN_API_TOKEN="ваш-токен"
export KAITEN_URL="https://mycompany.kaiten.ru"
# Опционально; по умолчанию ~/.kaiten/kaiten.db
export KAITEN_DB_PATH="$HOME/.kaiten/kaiten.db"
```

## Использование

### Пространства (spaces)

```bash
# Список всех пространств
kaiten spaces list

# Получить пространство по ID
kaiten spaces get 123
```

### Доски (boards)

```bash
# Список досок в пространстве
kaiten boards list --space-id 123

# Получить доску по ID
kaiten boards get 456
```

### Колонки (columns)

```bash
# Список колонок на доске
kaiten columns list --board-id 456
```

### Дорожки (lanes)

```bash
# Список дорожек на доске
kaiten lanes list --board-id 456
```

### Карточки (cards)

```bash
# Список карточек на доске (по умолчанию — активные)
kaiten cards list --board-id 456

# Архивные карточки
kaiten cards list --board-id 456 --archived

# Все карточки (активные + архивные)
kaiten cards list --board-id 456 --all

# Компактный список из live API: первые 50 карточек, только нужные поля, без описаний
kaiten cards list --board-id 456 --limit 50 \
  --fields id,title,condition,updated,completed_at \
  --no-descriptions --output table

# Все страницы API в JSONL для потоковой обработки
kaiten cards list --board-id 456 --all-pages --output jsonl \
  --fields id,title,board_id,updated

# Получить карточку по ID
kaiten cards get 789

# Создать карточку (по ID колонки)
kaiten cards create \
  --title "Новая задача" \
  --board-id 456 \
  --column-id 10 \
  --lane-id 20 \
  --position 1 \
  --description "Описание задачи" \
  --due-date "2025-12-31T23:59:59Z"

# Создать карточку (по имени колонки)
kaiten cards create \
  --title "Новая задача" \
  --board-id 456 \
  --column-name "Backlog" \
  --lane-name "Default" \
  --size 3

# Обновить карточку (только указанные поля)
kaiten cards update 789 \
  --title "Обновлённое название" \
  --column-id 11

# Переместить карточку по имени колонки
kaiten cards update 789 \
  --board-id 456 \
  --column-name "In Review"

# Удалить карточку
kaiten cards delete 789

# Архивировать карточку
kaiten cards archive 789

# Разархивировать карточку
kaiten cards unarchive 789
```

### Локальная SQLite БД и синхронизация

`kaiten sync` загружает данные из Kaiten в локальную SQLite БД. Для синхронизации нужен ровно один scope: `--board-id`, `--space-id` или `--all`.

```bash
# Начальная синхронизация одной доски
kaiten sync --board-id 456

# Синхронизация всех досок пространства
kaiten sync --space-id 123

# Синхронизация всех доступных данных
kaiten sync --all

# Проверить состояние БД
kaiten db status

# Освободить место после больших обновлений/удалений
kaiten db vacuum

# Полностью удалить локальную БД (без интерактива, для автоматизации)
kaiten db reset --yes
```

### Локальный поиск карточек

`kaiten search cards` ищет только в локальной SQLite БД и не делает API-вызовов. Перед поиском выполните `kaiten sync`.

```bash
# Карточки, завершённые в прошлом году конкретным участником; JSONL + короткий набор полей
kaiten search cards \
  --member-id 100 \
  --completed-from 2025-01-01 \
  --completed-to 2025-12-31 \
  --condition done \
  --fields id,title,board_id,completed_at,owner_id \
  --output jsonl

# То же по владельцу карточки
kaiten search cards \
  --owner-id 100 \
  --completed-from 2025-01-01 \
  --completed-to 2025-12-31 \
  --fields id,title,completed_at \
  --output jsonl

# Текстовый поиск по локальной БД
kaiten search cards --text "авторизация" --board-id 456 --limit 20 --output table
```

Фильтры поиска: `--text`, `--board-id`, `--space-id`, `--owner-id`, `--member-id`, `--tag-id`, `--condition`, `--created-from/to`, `--updated-from/to`, `--completed-from/to`, `--limit`, `--offset`, `--sort`.

### Комментарии (comments)

```bash
# Список комментариев карточки
kaiten comments list --card-id 789

# Добавить комментарий
kaiten comments add --card-id 789 --text "Текст комментария"

# Добавить комментарий в формате HTML
kaiten comments add --card-id 789 --text "<b>Важно</b>" --type 2
```

### Блокировки (blockers)

```bash
# Заблокировать карточку
kaiten blockers block --card-id 789 --reason "Ждём ответа от заказчика"

# Разблокировать (нужен ID блокировки — берётся из ответа block)
kaiten blockers unblock --card-id 789 --blocker-id 1234

# Удалить блокировку
kaiten blockers delete --card-id 789 --blocker-id 1234
```

### Теги (tags)

```bash
# Список всех тегов компании
kaiten tags list

# Теги на карточке
kaiten tags card-tags --card-id 789

# Добавить тег к карточке (по имени)
kaiten tags add --card-id 789 --name "urgent"

# Удалить тег с карточки
kaiten tags remove --card-id 789 --tag-id 123
```

### Участники (members)

```bash
# Список участников карточки
kaiten members list --card-id 789
```

### Чеклисты (checklists)

```bash
# Создать чеклист
kaiten checklists create --card-id 789 --name "Чеклист задачи"

# Получить чеклист с пунктами
kaiten checklists get --card-id 789 --checklist-id 111

# Добавить пункт
kaiten checklists add-item --card-id 789 --checklist-id 111 --text "Первый пункт"

# Отметить пункт выполненным
kaiten checklists check --card-id 789 --checklist-id 111 --item-id 222

# Снять отметку
kaiten checklists uncheck --card-id 789 --checklist-id 111 --item-id 222

# Удалить пункт
kaiten checklists delete-item --card-id 789 --checklist-id 111 --item-id 222

# Удалить чеклист
kaiten checklists delete --card-id 789 --checklist-id 111
```

### Флаги создания карточки

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | Да | Название карточки |
| `--board-id` | Да | ID доски |
| `--column-id` | Да* | ID колонки |
| `--column-name` | Да* | Имя колонки (вместо --column-id) |
| `--lane-id` | Нет | ID дорожки |
| `--lane-name` | Нет | Имя дорожки (вместо --lane-id) |
| `--position` | Нет | Позиция: `1` — вверху, `2` — внизу |
| `--description` | Нет | Описание |
| `--type-id` | Нет | ID типа карточки |
| `--size` | Нет | Размер карточки (число) |
| `--due-date` | Нет | Дедлайн (ISO 8601) |

* Обязателен один из двух: --column-id или --column-name

### Флаги обновления карточки

| Флаг | Описание |
|---|---|
| `--title` | Новое название |
| `--board-id` | ID доски (нужен при использовании --column-name или --lane-name) |
| `--column-id` | ID новой колонки (перемещение) |
| `--column-name` | Имя новой колонки (вместо --column-id) |
| `--lane-id` | ID новой дорожки |
| `--lane-name` | Имя новой дорожки (вместо --lane-id) |
| `--type-id` | Новый ID типа карточки |
| `--description` | Новое описание |
| `--size` | Новый размер карточки (число) |
| `--due-date` | Новый дедлайн (ISO 8601) |

## Формат вывода

По умолчанию команды возвращают JSON с отступами. Для `cards list`, `search cards` и `sync` можно выбрать формат:

```bash
kaiten search cards --board-id 456 --output jsonl
kaiten search cards --board-id 456 --output table --fields id,title,condition
kaiten cards list --board-id 456 --quiet
kaiten cards list --board-id 456 --no-descriptions
```

Доступные output controls для `cards list` и `search cards`:

- `--output json|jsonl|table|yaml|csv`
- `--fields id,title,...` — оставить только нужные поля
- `--no-descriptions` — исключить поля описаний
- `--quiet`, `-q` — вывести только ID карточек, по одному в строке

Для `sync` доступен выбор формата через `--output` / `-o`.

Пример JSON:

```json
{
  "id": 789,
  "title": "Моя карточка",
  "board_id": 456,
  "column_id": 10,
  "type_id": 5,
  "size_text": "3",
  "condition": "in_progress"
}
```

Для обработки вывода удобно использовать `jq`:

```bash
# Получить только названия карточек
kaiten cards list --board-id 456 | jq '.[].title'

# Отфильтровать карточки по колонке
kaiten cards list --board-id 456 | jq '[.[] | select(.column_id == 10)]'
```

## Коды ошибок API

| Код | Описание |
|---|---|
| 401 | Невалидный или отсутствующий токен |
| 402 | Функция недоступна на вашем тарифе |
| 403 | Недостаточно прав доступа |
| 404 | Ресурс не найден |
| 429 | Превышен лимит запросов (5 запросов/сек) |

## Структура проекта

```
├── main.go                     # Точка входа
├── Makefile                    # Сборка и линтинг
├── .goreleaser.yml             # Конфигурация релизов
├── go.mod / go.sum             # Зависимости
└── internal/
    ├── config/
    │   └── config.go           # Чтение KAITEN_API_TOKEN, KAITEN_URL и KAITEN_DB_PATH
    ├── api/
    │   ├── client.go           # HTTP-клиент с авторизацией
    │   ├── spaces.go           # API пространств
    │   ├── boards.go           # API досок
    │   ├── columns.go          # API колонок
    │   ├── lanes.go            # API дорожек
    │   ├── cards.go            # API карточек
    │   ├── comments.go         # API комментариев
    │   ├── blockers.go         # API блокировок
    │   ├── tags.go             # API тегов
    │   ├── members.go          # API участников
    │   └── checklists.go       # API чеклистов
    ├── db/                     # Локальная SQLite БД и поисковые запросы
    ├── output/                 # Форматы вывода и выбор полей
    ├── syncer/                 # Синхронизация Kaiten → SQLite
    └── cmd/
        ├── root.go             # Корневая команда
        ├── spaces.go           # Команды пространств
        ├── boards.go           # Команды досок
        ├── columns.go          # Команды колонок
        ├── lanes.go            # Команды дорожек
        ├── cards.go            # Команды карточек
        ├── comments.go         # Команды комментариев
        ├── blockers.go         # Команды блокировок
        ├── tags.go             # Команды тегов
        ├── members.go          # Команды участников
        ├── checklists.go       # Команды чеклистов
        ├── sync.go             # Команда sync
        ├── search.go           # Локальный поиск
        ├── db.go               # Управление локальной БД
        └── resolve.go          # Резолвинг имён в ID
```

## Ограничения текущей версии

- Нет назначения участников на карточки (только просмотр)
- Локальный поиск работает по последнему состоянию после `kaiten sync`
- API-клиент ограничивает запросы до 5 req/s и повторяет 429/5xx

## Разработка

```bash
# Сборка
make build

# Запуск тестов
make test

# Линтинг (нужен golangci-lint)
make lint

# Очистка
make clean
```

## Лицензия

Внутренний проект [Life Pay](https://life-pay.ru/).
