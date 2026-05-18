# Plan: OpenAI-Compatible API Recognizer

**Date:** 2026-05-18
**Branch:** `crispasr-integration`
**Status:** ✅ DONE (2026-05-18)
**Commit:** 849f214

## Goal
Додати третю опцію розпізнавання — OpenAI-compatible API, яка приймає не тільки API-ключ, але й URL сервера (наприклад, для локальних Ollama, vLLM, LM Studio тощо).

---

## Крок 1: `pkg/recognizer/openai.go` — новий конструктор ✅

- [x] Створити `newOpenAIRecognizerWithBaseURL(apiKey, baseURL, model, callsign string) Recognizer`, який використовує `option.WithBaseURL(baseURL)` + `option.WithAPIKey(apiKey)`
- [x] Експортувати публічний конструктор `NewOpenAICompatibleRecognizer(apiKey, baseURL, model, callsign string) Recognizer`
- [x] Рефакторити існуючі конструктори — залишено як є, новий конструктор використовує спільну структуру з `WithBaseURL`

## Крок 2: `internal/conf/configuration.go` — новий тип + поле ✅

- [x] Додати константу `OpenAICompatible Recognizer = "openai-compatible"`
- [x] Додати поле `OpenAIAPIBaseURL string` до `Configuration` (порожнє для інших типів recognizer)

## Крок 3: `cmd/skyeye/main.go` — CLI flag + config wiring ✅

- [x] Додати змінну `openAIAPIBaseURL string` та флаг `--openai-api-base-url`
- [x] Додати `"openai-compatible"` до enum `recognizerFlag` (рядок 129)
- [x] `MarkFlagsOneRequired("whisper-model", "openai-api-key")` залишено без змін — openai-api-key потрібен для всіх API-базованих режимів
- [x] Додати поле `OpenAIAPIBaseURL: openAIAPIBaseURL` до конфігу при створенні

## Крок 4: `internal/application/app.go` — switch case ✅

- [x] Додати кейс `conf.OpenAICompatible` у switch, який викликає `recognizer.NewOpenAICompatibleRecognizer(config.OpenAIAPIKey, config.OpenAIAPIBaseURL, "whisper", config.Callsign)`
- [x] Модель для compatible режиму — hardcoded `"whisper"` (для більшості OpenAI-compatible серверів це правильний дефолт)

---

## Usage Example

```bash
# Using local Ollama with whisper model
skyeye --recognizer openai-compatible \
  --openai-api-key "anything" \
  --openai-api-base-url "http://localhost:11434/v1"

# Using vLLM or LM Studio
skyeye --recognizer openai-compatible \
  --openai-api-key "your-key" \
  --openai-api-base-url "http://localhost:8000/v1"
```

## Notes
- OpenAI Go SDK (`github.com/openai/openai-go`) вже підтримує `option.WithBaseURL()` — перевірено
- Для compatible режиму користувач вказує свій URL (наприклад `http://localhost:8000/v1`) та API ключ
- Модель hardcoded як `"whisper"` — достатньо для більшості use-case. Якщо потрібна гнучкість, можна додати окремий флаг пізніше
