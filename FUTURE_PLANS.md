# Планы развития kaiten-cli

Документ фиксирует, что уже реализовано, и что остаётся в планах. Для подробного описания текущих команд см. `README.md` и `.skill/SKILL.md`.

## Уже реализовано

### Карточки: пагинация и компактный вывод

- `kaiten cards list --limit <n>` — ограничить размер страницы live API.
- `kaiten cards list --offset <n>` — смещение для пагинации live API.
- `kaiten cards list --all-pages` — загрузить все страницы карточек доски.
- `--output json|jsonl|table|yaml|csv` для `cards list`, `search cards`, `sync`.
- `--fields`, `--no-descriptions`, `--quiet/-q` для `cards list` и `search cards`.

### Расширенная модель карточки

CLI парсит и отдаёт поля, полезные для аналитики и локального индекса:

- `owner_id`
- `updater_id`
- `created` / `updated`
- `completed_at`
- `column_changed_at`
- `comments_total`
- `tag_ids` в live API output
- `sprint_id`
- `sort_order`

### Локальная SQLite БД, sync и поиск

- `KAITEN_DB_PATH`, по умолчанию `~/.kaiten/kaiten.db`.
- `kaiten sync --board-id <id>`, `--space-id <id>`, `--all`.
- `kaiten db status`, `kaiten db reset --yes`, `kaiten db vacuum`.
- `kaiten search cards` ищет по локальной БД без API-вызовов.
- Фильтры `search cards`: `--text`, `--board-id`, `--space-id`, `--owner-id`, `--member-id`, `--tag-id`, `--condition`, `--created-from/to`, `--updated-from/to`, `--completed-from/to`, `--limit`, `--offset`, `--sort`.
- SQLite schema хранит пространства, доски, карточки, пользователей, участников карточек, теги, связи карточка-тег и состояние sync.
- FTS5 индексирует `title` и `description` карточек.

### API-клиент

- In-memory rate limiter 5 запросов/сек.
- Retry для HTTP 429 и 5xx.
- Поддержка `Retry-After`.

## Ближайшие улучшения

### Инкрементальный sync

Сейчас `sync` выполняет полную синхронизацию выбранного scope. Возможные улучшения:

- инкрементальная синхронизация по `updated`/`updated_after`, если API позволит;
- режим `sync --since <date>`;
- отображение прогресса для больших пространств;
- отчёт о частичных ошибках вместо полного fail-fast, если это безопасно для консистентности индекса.

### Live API фильтры в `cards list`

Сейчас сложные фильтры доступны локально через `kaiten search cards`. Для live API можно добавить, если это реально поддерживается Kaiten API и полезно без локального индекса:

- `cards list --owner-id <id>`
- `cards list --member-id <id>`

Важно: API уже был проверен и игнорирует date/text/sort параметры; такие фильтры должны оставаться локальными через SQLite.

### Единый вывод для всех команд

Сейчас расширенные output controls покрывают основные карточные/аналитические команды. Возможные улучшения:

- `--output json|jsonl|table|yaml|csv` для всех read-only команд;
- `--fields` для `spaces`, `boards`, `columns`, `lanes`, `comments`, `tags`, `members`, `checklists`;
- шаблоны вывода для часто используемых отчётов.

## Остальные планы

### Назначение участников на карточки

Сейчас `members list` — только просмотр. Добавить:

- `members add --card-id <id> --user-id <id>` — назначить участника;
- `members remove --card-id <id> --user-id <id>` — убрать участника;
- исследовать API endpoint для управления участниками.

### Управление пользователями и командами

- `users list` — список пользователей компании;
- `users get <id>` — информация о пользователе;
- резолвинг имени пользователя в ID, аналогично `--column-name`.

### Создание и управление досками

- `boards create --space-id <id> --title "..."` — создание доски;
- `boards update <id> --title "..."` — обновление доски;
- `boards delete <id>` — удаление доски.

### Управление пространствами

- `spaces create --title "..."` — создание пространства;
- `spaces update <id> --title "..."` — обновление пространства;
- `spaces delete <id>` — удаление пространства.

### Конфигурационный файл

- Поддержка `~/.kaiten.yml` или `~/.config/kaiten/config.yml` как альтернатива env vars;
- профили для нескольких Kaiten-инстансов, например `--profile work`.

### Дочерние карточки

- `cards children <id>` — список дочерних карточек;
- `cards create-child --parent-id <id> --title "..."` — создание дочерней карточки;
- API поддерживает `children_count`, `children_done`, `children_ids` в ответе.

### Логи и аудит

- `cards history <id>` — история изменений карточки, если API поддерживает.

### Массовые операции

- `cards move --board-id <id> --from-column "Done" --to-column "Archive"` — массовое перемещение;
- `cards archive-all --board-id <id> --column-name "Done"` — архивация всех из колонки.

### Настройка rate limiter

- Настройка лимита через env/config;
- метрики retry и backoff в verbose/debug режиме.

### Автодополнение shell

- `kaiten completion bash|zsh|fish` — генерация автодополнения Cobra.

## Не поддерживается Kaiten API

Протестировано: API игнорирует эти параметры, поэтому CLI должен обрабатывать такие запросы через локальный SQLite индекс.

- Фильтрация по дате: `created_on`, `updated_on`.
- Поиск по тексту: `title`, `title_contains`, `search`.
- Сортировка: `sort_by`.
