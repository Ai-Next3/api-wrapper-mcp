# API-обертка для протокола MCP

Этот проект представляет собой сервер-обертку на Go, который позволяет превращать любые REST API в инструменты, совместимые с Model-Client Protocol (MCP). Это дает возможность большим языковым моделям (LLM), таким как Claude, взаимодействовать с внешними API, используя их как свои собственные инструменты.

## Архитектура

Проект состоит из двух основных частей:

1.  **Консольный MCP-сервер** (`cmd/stdio_server/main.go`): Ядро, которое преобразует вызовы API в MCP-инструменты. Оно общается через стандартные потоки ввода/вывода (stdin/stdout).
2.  **Веб-сервер-обертка** (`main.go`): Запускает консольный сервер как дочерний процесс и "пробрасывает" его общение в интернет через WebSocket. Это позволяет подключаться к серверу из любого места, включая облачные платформы.

## Запуск и использование

Есть два основных способа запуска сервера:

### 1. Локальный запуск (для разработки и отладки)

Этот режим идеально подходит для быстрой проверки и добавления новых инструментов в `config.yaml`. В этом режиме мы запускаем только консольную часть, без веб-обертки.

**Шаги для запуска:**

1.  **Установите API-ключ** (если требуется для вашего API). В нашем примере с Wildberries используется переменная `WB_API_TOKEN`.
    ```bash
    export WB_API_TOKEN="ВАШ_API_КЛЮЧ"
    ```

2.  **Скомпилируйте и запустите сервер**, указав путь к вашему файлу конфигурации.
    ```bash
    # Перейдите в папку с консольным приложением
    cd cmd/stdio_server

    # Запустите, передав конфиг из корня проекта
    go run main.go ../../wildberries-config.yaml
    ```
После этого сервер будет запущен и готов к подключению от локальных MCP-клиентов (например, Claude Desktop).

### 2. Запуск как Веб-сервер (Replit, VDS, Docker)

Этот режим запускает WebSocket-сервер, который делает ваш MCP-сервер доступным по сети.

**Шаги для запуска:**

1.  **Установите переменные окружения.**
    - `WB_API_TOKEN`: Ваш ключ доступа к API.
    - `PORT` (опционально): Порт, на котором будет работать сервер. Если не указан, по умолчанию используется `8081`. На платформах вроде Replit эта переменная обычно устанавливается автоматически.

    **На VDS или локально:**
    ```bash
    export WB_API_TOKEN="ВАШ_API_КЛЮЧ"
    export PORT="8080" # по желанию
    ```

2.  **Запустите главный сервер.**
    ```bash
    # Запускать нужно из корня проекта
    go run main.go
    ```

После запуска в консоли появится сообщение:
`Starting WebSocket proxy on http://localhost:8081/ws`

Теперь ваш сервер доступен для подключения по WebSocket по адресу `ws://<адрес_вашего_сервера>:<порт>/ws`.

## Конфигурация

Вся настройка инструментов происходит в YAML-файле (например, `wildberries-config.yaml`).

```yaml
# Информация о сервере
server:
  name: "Wildberries API"
  description: "Обертка для API Wildberries"
  version: "1.0.0"
  
# Аутентификация
auth:
  # Название переменной окружения, где хранится API-ключ
  token_env_var: "WB_API_TOKEN" 

# Определения инструментов
tools:
  - name: "wb_get_balance"
    description: "Получить баланс продавца Wildberries."
    endpoint: "https://finance-api.wildberries.ru/api/v1/account/balance"
    method: "GET"
    timeout: 30
    parameters: {} # Параметры не требуются

  - name: "wb_get_documents_list"
    description: "Получить список документов продавца за период."
    endpoint: "https://documents-api.wildberries.ru/api/v1/documents/list"
    method: "GET"
    timeout: 30
    # Параметры для GET-запроса
    query_params:
      beginTime: "{{beginTime}}"
      endTime: "{{endTime}}"
    # Описание параметров для LLM
    parameters:
      beginTime:
        type: "string"
        description: "Начало периода в формате 'ГГГГ-ММ-ДД'"
        required: false
      endTime:
        type: "string"
        description: "Конец периода в формате 'ГГГГ-ММ-ДД'"
        required: false
```

### Добавление POST-запроса

Для POST-запросов используется поле `template` для формирования тела запроса.

```yaml
  - name: "example_post_request"
    description: "Пример POST-запроса."
    endpoint: "https://api.example.com/v1/items"
    method: "POST"
    timeout: 30
    # Шаблон для тела JSON-запроса
    template: |
      {
        "name": "{{itemName}}",
        "quantity": {{itemQuantity}}
      }
    # Описание параметров для LLM
    parameters:
      itemName:
        type: "string"
        description: "Название товара"
        required: true
      itemQuantity:
        type: "number"
        description: "Количество товара"
        required: true
```

## Claude Desktop Integration

To use with Claude Desktop, add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "api-wrapper": {
      "command": "path/to/api_wrapper",
      "args": ["path/to/your-config.yaml"],
      "env": {
        "API_GATEWAY_TOKEN": "your-api-token"
      }
    }
  }
}
```

## Examples

Check out `example-config.yaml` for sample API configurations.

## Environment Variables

- Set the main authentication token using the environment variable specified in the `auth.token_env_var` field.
- You can also reference other environment variables in your templates using `{{env:VARIABLE_NAME}}` syntax.
