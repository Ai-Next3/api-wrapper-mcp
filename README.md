# API-обертка для протокола MCP

Этот проект представляет собой сервер-обертку на Go, который позволяет превращать любые REST API в инструменты, совместимые с Model-Client Protocol (MCP). Это дает возможность большим языковым моделям (LLM), таким как Claude, взаимодействовать с внешними API, используя их как свои собственные инструменты.

## Архитектура

Проект представляет собой единое приложение на Go, которое может работать в двух режимах:
1.  **stdio (стандартный ввод/вывод):** В этом режиме сервер общается через консоль. Идеально подходит для локальной разработки и интеграции с настольными клиентами, такими как Claude Desktop.
2.  **HTTP-сервер:** В этом режиме запускается полноценный веб-сервер с поддержкой протокола `streamable-http`, что позволяет подключаться к нему по сети с любого удаленного клиента, поддерживающего MCP.

## Запуск и использование

Есть два основных способа запуска сервера:

### 1. Локальный запуск (режим stdio)

Этот режим предназначен для быстрой проверки и отладки.

**Шаги для запуска:**

1.  **Установите API-ключ** (если требуется для вашего API). В нашем примере с Wildberries используется переменная `WB_API_TOKEN`.
    ```bash
    export WB_API_TOKEN="ВАШ_API_КЛЮЧ"
    ```

2.  **Запустите сервер с флагом `--stdio`**, указав путь к вашему файлу конфигурации.
    ```bash
    go run main.go --stdio wildberries-config.yaml
    ```
После этого сервер будет запущен и готов к подключению от локальных MCP-клиентов (например, Claude Desktop).

### 2. Запуск как Веб-сервер (Replit, VDS, Docker)

Этот режим запускает `streamable-http` сервер, который делает ваши MCP-инструменты доступными по сети.

**Шаги для запуска:**

1.  **Установите переменные окружения.**
    - `WB_API_TOKEN`: Ваш ключ доступа к API.
    - `PORT` (опционально): Порт, на котором будет работать сервер. Если не указан, по умолчанию используется `8081`. На платформах вроде Replit эта переменная обычно устанавливается автоматически.

    **На VDS или локально:**
    ```bash
    export WB_API_TOKEN="ВАШ_API_КЛЮЧ"
    export PORT="8080" # по желанию
    ```

2.  **Запустите главный сервер**, указав путь к конфигу.
    ```bash
    go run main.go wildberries-config.yaml
    ```

После запуска в консоли появится сообщение:
`Starting StreamableHTTP MCP server on http://localhost:8081`

Теперь ваш сервер доступен для подключения по адресу `http://<адрес_вашего_сервера>:<порт>`.

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
