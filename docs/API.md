# API rw-pair — интеграция с основным ботом

> Документ для будущего внедрения в основном боте. Контракт сверен с кодом rw-pair: `main.go`, `bot.go`, `store.go`, `ratelimit.go`. Если меняется код сервиса — сначала обнови этот файл, потом код бота.

## 1. Что нужно сделать в боте — всего 2 вещи

1. **Хендлер выдачи кода**: по нажатию пользователя вызвать `POST /pair/create` в rw-pair, получить 6-значный код и показать его пользователю вместе с адресом `pair.dendenx.cloud`.
2. **Внутренний endpoint подписок**: принять `POST` от rw-pair с `telegram_id`, найти ключи пользователя в Panel API и вернуть их `shortUuid` + тип. Токен Panel остаётся только в боте.

Всё остальное уже делает rw-pair: хранение кодов, TTL, одноразовость, rate-limit, маппинг типов в домены и сборку `subscription_url`, страницу ввода кода.

## 2. Схема потока

```text
Бот --POST /pair/create {telegram_id}--> rw-pair        : код 123456, TTL 300 c
Бот --код + pair.dendenx.cloud--> пользователь
Пользователь --код на pair.dendenx.cloud--> страница --POST /pair/redeem--> rw-pair
rw-pair --POST BOT_SUBSCRIPTIONS_URL {telegram_id}--> Бот --Panel API--> ключи
rw-pair --{subscriptions:[{id,label,subscription_url}]}--> страница
```

## 3. Секреты и сеть

| Секрет | Направление | Где задаётся |
|---|---|---|
| `PAIR_INTERNAL_KEY` | бот -> rw-pair, заголовок `X-Internal-Key` на `POST /pair/create` | env rw-pair и env бота, значение одно и то же |
| `BOT_INTERNAL_KEY` | rw-pair -> бот, заголовок `X-Internal-Key` на внутреннем endpoint бота | env rw-pair и env бота; если в rw-pair пустой — используется `PAIR_INTERNAL_KEY`, то есть допустим один общий секрет на оба направления |

Правила сети:

- `POST /pair/create` вызывать **только напрямую по docker-сети**, например `http://rw-pair:8080/pair/create`. Не через Nginx и не через публичный `pair.dendenx.cloud`: Nginx на публичный `create` отвечает 404, а сам сервис принимает запрос только с loopback/private IP (`RemoteAddr`), иначе 403.
- Бот и rw-pair должны быть в одной docker-сети. В `docker-compose.yml` rw-pair сеть называется `pair-net`, порт наружу не публикуется.
- Секреты не коммитить, не логировать. Код и `shortUuid` целиком в логи не писать.

## 4. POST /pair/create — бот -> rw-pair

### Запрос

```http
POST http://rw-pair:8080/pair/create
Content-Type: application/json
X-Internal-Key: <PAIR_INTERNAL_KEY>
```

```json
{"telegram_id": 123456789, "ttl": 300}
```

| Поле | Тип | Обязательно | Правила |
|---|---|---|---|
| `telegram_id` | int64, число или числовая строка | да | Должен быть > 0. Сервис принимает и `123456789`, и `"123456789"`, хранит как int64 |
| `ttl` | int, секунды | нет | По умолчанию 300. `0` = по умолчанию. Допустимо 1..3600, для MVP всегда передавай 300 или не передавай поле |

Тело больше 4096 байт сервис не примет.

### Успех: 200

```json
{"code": "012345", "expires_in": 300}
```

`code` — строка из 6 цифр с сохранением ведущих нулей. Не приводи к числу: `012345` и `12345` — разные коды.

### Ошибки

| Статус | Тело `error` | Причина | Что делать в боте |
|---|---|---|---|
| 400 | `invalid body` / `telegram_id is required` / `ttl out of range` | Битый JSON, `telegram_id` <= 0 или не передан, `ttl` вне диапазона | Баг на стороне бота, пользователю код не показывать |
| 401 | `unauthorized` | Нет или неверный `X-Internal-Key` | Проверить, что бот шлёт именно `PAIR_INTERNAL_KEY` rw-pair |
| 403 | `forbidden` | Источник запроса не loopback/private IP | Вызывать сервис напрямую по docker-сети, не через публичный адрес |
| 405 | `method not allowed` | Не POST | — |
| 500 | `failed to create code` | Не удалось сгенерировать уникальный код после 10 попыток коллизий | Повторить один раз, дальше — сообщение о временной ошибке |

### Пример на Python (httpx)

```python
import os, httpx

PAIR_URL = os.environ['PAIR_SERVICE_URL']  # http://rw-pair:8080
PAIR_KEY = os.environ['PAIR_INTERNAL_KEY']

async def create_pair_code(telegram_id: int) -> str:
    async with httpx.AsyncClient(timeout=5.0) as client:
        r = await client.post(
            PAIR_URL + '/pair/create',
            json={'telegram_id': telegram_id, 'ttl': 300},
            headers={'X-Internal-Key': PAIR_KEY},
        )
    r.raise_for_status()
    return r.json()['code']  # строка, не int
```

### Что показать пользователю

Код, адрес страницы и срок действия. Например: код `012345`, откройте `pair.dendenx.cloud`, код действует 5 минут и работает один раз.

## 5. Внутренний endpoint бота — rw-pair -> бот

URL задаётся в rw-pair переменной `BOT_SUBSCRIPTIONS_URL`, например `http://bot:8081/internal/subscriptions`. Путь и порт — на стороне бота, главное, чтобы совпадали с этой переменной.

### Запрос от rw-pair

```http
POST <BOT_SUBSCRIPTIONS_URL>
Content-Type: application/json
X-Internal-Key: <BOT_INTERNAL_KEY или PAIR_INTERNAL_KEY, если BOT_INTERNAL_KEY пустой>
```

```json
{"telegram_id": 123456789}
```

`telegram_id` всегда приходит JSON-числом (int64).

Таймаут вызова на стороне rw-pair: `BOT_TIMEOUT_SECONDS`, по умолчанию 5 секунд. Endpoint бота должен отвечать быстрее.

### Что должен сделать бот

1. Проверить `X-Internal-Key` сравнением в constant time (`hmac.compare_digest` / `secrets.compare_digest`) с секретом из п. 3.
2. Найти пользователя по `telegram_id` в своей БД и запросить его ключи в Panel API.
3. Вернуть **все** активные ключи пользователя. Выбор ключа делает пользователь на странице, бот ничего не выбирает и `subscription_url` не собирает.

### Успех: 200

Предпочтительный формат:

```json
{"keys": [{"shortUuid": "<shortUuid-wh>", "type": "_wh"}, {"shortUuid": "<shortUuid-basic>", "type": "_basic"}]}
```

Допустимые варианты, которые rw-pair тоже понимает:

- Корневой ключ `subscriptions` вместо `keys`.
- Поле `short_uuid` вместо `shortUuid`.
- Поле `kind` вместо `type`.

Порядок ключей в ответе сохраняется и в ответе `redeem`, страница показывает их в этом же порядке.

### Маппинг типов — делает rw-pair, не бот

| `type` от бота | `id` в redeem | `label` | База URL |
|---|---|---|---|
| `_wh`, `wh`, `whitelist` | `_wh` | `Обход` | `WHITELIST_BASE_URL`, по умолчанию `https://whitelist.dendenx.cloud` |
| `_basic`, `basic`, пустая строка | `_basic` | `Базовый` | `BASIC_BASE_URL`, по умолчанию `https://sub.dendenx.cloud` |
| Любой другой | — | — | Ключ игнорируется |

`subscription_url` собирает rw-pair: `<база>/<shortUuid>`, завершающий `/` у базы обрезается. Боту нельзя присылать готовый URL или label — они будут проигнорированы. Ключи без `shortUuid` пропускаются.

### Статусы ответа бота — семантика критична

| Статус бота | Что сделает rw-pair |
|---|---|
| 200 с непустым списком распознаваемых ключей | `redeem` отвечает 200, код сгорает |
| 404 | `redeem` отвечает 404 `no subscriptions`. Код **не** сгорает сразу, но счётчик неудачных попыток по коду +1; после 5 таких ответов код сгорит |
| 200 с пустым списком, или все ключи нераспознаваемых типов / без `shortUuid` | То же, что 404 выше |
| Любой другой статус (401, 500 и т.д.), таймаут, недоступность, битый JSON | `redeem` отвечает 502, код **не** сгорает и счётчик не растёт — пользователь может повторить ввод до конца TTL |

Вывод для бота: «у пользователя нет ключей» — это строго 404 или 200 с пустым списком. Временная ошибка Panel — это 5xx, но тогда код у пользователя не сгорит и он сможет повторить.

### Псевдокод endpoint (framework не важен)

```python
async def internal_subscriptions(request):
    key = request.headers.get('X-Internal-Key', '')
    if not secrets.compare_digest(key, BOT_KEY):
        return json_response({'error': 'unauthorized'}, status=401)
    telegram_id = (await request.json())['telegram_id']
    keys = await panel_keys_by_telegram_id(telegram_id)  # запрос в Panel API, токен только здесь
    if not keys:
        return json_response({'error': 'not found'}, status=404)
    return json_response({'keys': [{'shortUuid': k.short_uuid, 'type': k.type} for k in keys]})
```

## 6. POST /pair/redeem — страница -> rw-pair

Бот этот endpoint не вызывает, раздел — чтобы понимать поведение кода.

```http
POST https://pair.dendenx.cloud/pair/redeem
Content-Type: application/json
```

```json
{"code": "012345"}
```

Успех 200:

```json
{"telegram_id": 123456789, "subscriptions": [{"id": "_wh", "label": "Обход", "subscription_url": "https://whitelist.dendenx.cloud/<shortUuid>"}, {"id": "_basic", "label": "Базовый", "subscription_url": "https://sub.dendenx.cloud/<shortUuid>"}]}
```

Ошибки: 400 `invalid body` / `invalid code` (не 6 цифр), 404 `code not found`, 404 `no subscriptions`, 410 `code expired`, 429 `too many attempts` с заголовком `Retry-After: 60`, 502 `upstream unavailable`.

Лимиты и жизненный цикл кода:

- Rate-limit: не более 5 запросов в минуту с одного IP. IP берётся из первого значения `X-Forwarded-For`, иначе `X-Real-IP`, иначе `RemoteAddr` — за Nginx обязательно проксировать реальные заголовки, конфиг уже есть в `nginx/pair.conf`.
- Код одноразовый: сгорает только при успешном ответе со списком подписок.
- При ошибках бота/Panel (502 на стороне rw-pair) код не сгорает.
- Хранилище in-memory, протухшие записи чистятся раз в минуту. Перезапуск rw-pair = все живые коды сгорели, приемлемо при TTL 5 минут. Сервис — одна реплика, Redis/БД нет.

## 7. GET /health

`GET /health` и `GET /pair/health` — 200 `{"status": "ok"}`, без аутентификации. Для healthcheck контейнера и проверки после деплоя.

## 8. Smoke-тест интеграции

1. Поднять rw-pair и бота в одной docker-сети, секреты совпадают.
2. `curl` в health rw-pair — 200.
3. Из сети бота вызвать `POST /pair/create` с ключом — 200 и 6 цифр; без ключа — 401; с публичного адреса — 404/403.
4. Открыть `pair.dendenx.cloud`, ввести код — при двух ключах страница показывает выбор «Базовый» / «Обход», при одном — сразу экран ссылки.
5. Повторно ввести тот же код — 404.
6. Остановить бота, ввести новый код — 502, код не сгорел: после запуска бота тот же код в пределах TTL срабатывает.
7. Пользователь без ключей: бот возвращает 404 — страница показывает отсутствие подписок, после 5 попыток код сгорает.

## 9. Частые проблемы

| Симптом | Причина |
|---|---|
| create всегда 401 | Бот шлёт не тот секрет: нужен `PAIR_INTERNAL_KEY` rw-pair |
| create 403/404 | Вызов идёт через публичный Nginx или не из docker-сети |
| redeem всегда 502 | `BOT_SUBSCRIPTIONS_URL` не бьётся, бот отвечает не 200/404, битый JSON или дольше 5 секунд |
| redeem 404 при живых ключах | Бот вернул пустой список, неизвестный `type` или ключи без `shortUuid` |
| В URL не тот домен | Правятся `WHITELIST_BASE_URL` / `BASIC_BASE_URL` в env rw-pair, не код бота |
| Коды пропали после рестарта | Норма для MVP: in-memory хранилище, выдать новый код |
