# Plan: OpenAI-Compatible API Recognizer

**Date:** 2026-05-18
**Branch:** `crispasr-integration`
**Status:** 🔄 IN PROGRESS (2026-05-19)
**Commit:** 1f3f07f

## Goal
Додати підтримку OpenAI-compatible API для STT та TTS, що дозволяє використовувати будь-які сумісні сервери (Ollama, vLLM, LM Studio тощо).

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

## Крок 5: Тестування на сервері 🔄

- [ ] Зібрати бінарник на сервері з Go (`make skyeye`)
- [ ] Запустити інтеграційне тестування з реальним OpenAI-compatible сервером (наприклад, Ollama + whisper модель)
  - Перевірити `--recognizer openai-compatible --openai-api-base-url "http://localhost:11434/v1"`
- [ ] Перевірити fallback поведінку при недоступності сервера
- [ ] Запустити `make test` та `make lint vet fix format`

## Крок 6: OpenAI-compatible TTS Speaker ✅

Додати підтримку OpenAI-compatible API для синтезу мовлення (TTS), аналогічно до STT.

- [x] Створити `pkg/synthesizer/speakers/openai.go` — новий `openAITTSSpeaker` struct, що реалізує `Speaker`
  - Використано raw HTTP запити до `/v1/audio/speech` (openai-go alpha.41 має нестабільне API Speech)
  - `response_format: "pcm"` — повертає raw 16-bit LE PCM (24kHz), без MP3 декодера
  - Конвертувати PCM → downsample до 16kHz → `[]float32` через `downsample()` + `pcm.S16LEBytesToF32LE()`
- [x] Додати тип `TTSEngine string` та константи (`"piper"`, `"macos"`, `"openai-compatible"`) до `internal/conf/configuration.go`
  - Поля `OpenAITTSModel` та `OpenAITTSVoice` до `Configuration`
  - Для macOS/Linux залишено автодетекцію через `runtime.GOOS` як fallback, якщо явно не вказано engine
- [x] Додати CLI флаги:
  - `--tts-engine openai-compatible` — вибір TTS движка (default: платформозалежний — piper/macos)
  - `--openai-tts-model` — модель для TTS (default: `"tts-1"`)
  - `--openai-tts-voice` — голос для TTS (default: `"alloy"`)
  - `--openai-api-base-url` вже існує, використовується спільно для STT та TTS
- [x] Додати кейс у switch `internal/application/app.go` для створення TTS speaker на основі `config.TTSEngine`

---

## Usage Example

```bash
# Using local Ollama with whisper model (STT only)
skyeye --recognizer openai-compatible \
  --openai-api-key "anything" \
  --openai-api-base-url "http://localhost:11434/v1"

# Using vLLM or LM Studio for both STT and TTS
skyeye --recognizer openai-compatible \
  --tts-engine openai-compatible \
  --openai-api-key "your-key" \
  --openai-api-base-url "http://localhost:8000/v1" \
  --openai-tts-model tts-1 \
  --openai-tts-voice alloy

# Using OpenAI Platform for TTS with local Piper (default)
skyeye --recognizer openai-whisper-local \
  --whisper-model ggml-small.en.bin \
  --tts-engine openai-compatible \
  --openai-api-key "sk-..." \
  --openai-tts-model tts-1-hd \
  --openai-tts-voice nova
```

## Notes
- OpenAI Go SDK (`github.com/openai/openai-go`) вже підтримує `option.WithBaseURL()` — перевірено
- Для compatible режиму користувач вказує свій URL (наприклад `http://localhost:8000/v1`) та API ключ
- Модель hardcoded як `"whisper"` — достатньо для більшості use-case. Якщо потрібна гнучкість, можна додати окремий флаг пізніше
