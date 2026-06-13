# 🚀 МаркетингПро — запуск проекта на вашем компьютере

Это пошаговая инструкция, как поднять платформу **МаркетингПро** локально после того, как вы скачали и распаковали архив с проектом.

---

## 📦 Что вам понадобится

1. **Go** версии **1.21** или новее — <https://go.dev/dl/>
2. **PostgreSQL** версии **14** или новее — <https://www.postgresql.org/download/>
3. Любой текстовый редактор (VS Code, Sublime Text, Notepad++)
4. Терминал (Командная строка / PowerShell на Windows, Terminal на macOS/Linux)

> ❗ **Никаких Node.js, npm, React, TypeScript ставить не нужно.** Фронтенд — это обычный HTML + CSS + JavaScript, его раздаёт сам Go-сервер.

---

## 🗂️ Структура проекта

```
api-server-go/
├── main.go                 ← точка входа
├── models.go               ← структуры данных
├── helpers.go              ← вспомогательные функции
├── db.go                   ← подключение к БД, миграции, начальные данные
├── handlers_auth.go        ← регистрация / вход / выход
├── handlers_courses.go     ← список курсов и уроков
├── handlers_dashboard.go   ← главная панель, прогресс, записи на курсы
├── handlers_profile.go     ← профиль пользователя, загрузка файлов
├── handlers_admin.go       ← админ-панель (CRUD курсов)
├── go.mod / go.sum         ← зависимости Go
└── static/                 ← фронтенд (HTML/CSS/JS) + папка uploads
```

---

## 🐘 Шаг 1. Установка PostgreSQL

### Windows
1. Скачайте установщик с <https://www.postgresql.org/download/windows/>
2. Во время установки **запомните пароль** для пользователя `postgres`
3. Оставьте порт по умолчанию — `5432`

### macOS
```bash
brew install postgresql@16
brew services start postgresql@16
```

### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

---

## 🛠️ Шаг 2. Создание базы данных

Откройте терминал и подключитесь к PostgreSQL:

**Windows / macOS / Linux:**
```bash
psql -U postgres
```
(введите пароль, который задали при установке)

Внутри `psql` выполните:

```sql
CREATE DATABASE marketingpro;
CREATE USER mpuser WITH PASSWORD 'mppass123';
GRANT ALL PRIVILEGES ON DATABASE marketingpro TO mpuser;
\c marketingpro
GRANT ALL ON SCHEMA public TO mpuser;
\q
```

> ✅ Готово — база `marketingpro` создана, пользователь `mpuser` имеет к ней полный доступ.

**Таблицы создавать вручную не нужно** — сервер при первом запуске сам создаст все таблицы (users, courses, lessons, enrollments, lesson_progress) и наполнит каталог 6 курсами с 15 уроками.

---

## 🔑 Шаг 3. Файл `.env` с настройками (рекомендуемый способ)

Все настройки задаются через файл `.env` — больше не нужно каждый раз вводить переменные в терминале.

### Что делать

1. Перейдите в папку `api-server-go` — там уже лежит готовый шаблон `.env.example`.
2. Скопируйте его в файл с именем `.env`:

   **Windows (PowerShell):**
   ```powershell
   copy .env.example .env
   ```
   **macOS / Linux:**
   ```bash
   cp .env.example .env
   ```

3. Откройте `.env` в любом редакторе и впишите свои данные. Самая простая форма — одна строка подключения:

   ```env
   DATABASE_URL=postgres://mpuser:Your3@localhost:Your/Your?sslmode=disable
   ADMIN_SECRET=Your
   PORT=3000
   ```

   Или, если удобнее задавать настройки по отдельности (например, при использовании pgAdmin), уберите `DATABASE_URL` и заполните каждую переменную отдельно:

   ```env
   DB_HOST=localhost
   DB_PORT=Your
   DB_USER=Your
   DB_PASSWORD=Your
   DB_NAME=Your
   DB_SSLMODE=disable
   ADMIN_SECRET=Your
   PORT=3000
   ```

| Переменная | Что это |
|-----------|---------|
| `DATABASE_URL` | полная строка подключения (если задана — используется первой) |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` | альтернатива `DATABASE_URL` — сервер сам соберёт строку |
| `ADMIN_SECRET` | секретный код для активации прав администратора (по умолчанию `admin123`) |
| `PORT` | порт HTTP-сервера (по умолчанию `3000`) |

> 🔒 Файл `.env` уже добавлен в `.gitignore` — он не попадёт в Git, ваши пароли останутся только у вас на ПК.

### Альтернатива: системные переменные окружения

Можно по-старому, без `.env` — через переменные окружения. Они имеют приоритет над `.env`:

**Windows (PowerShell):**
```powershell
$env:DATABASE_URL="postgres://mpuser:Your3@localhost:Your/Your?sslmode=disable"
```

**Windows (cmd.exe):**
```cmd
set DATABASE_URL=postgres://mpuser:Your3@localhost:Your/Your?sslmode=disable
```

**macOS / Linux:**
```bash
export DATABASE_URL="postgres://mpuser:Your3@localhost:Your/Your?sslmode=disable"
```

---

## ▶️ Шаг 4. Запуск сервера

В терминале перейдите в папку `api-server-go` и запустите:

```bash
cd api-server-go
go mod download
go run .
```

> ⚠️ Важно: используйте именно `go run .` (с точкой), а **не** `go run main.go`. Точка означает «скомпилировать все файлы в этой папке» — это нужно, потому что код разбит на несколько файлов.

После запуска вы увидите:
```
2026/04/22 12:32:42 Подключение к базе данных установлено
2026/04/22 12:32:42 🚀 Сервер МаркетингПро запущен на порту 3000
```

Откройте браузер: <http://localhost:3000>

---

## 👨‍💼 Шаг 5. Получение прав администратора

1. Зарегистрируйтесь в приложении (любой email и пароль).
2. Зайдите в раздел **Профиль** → кнопка **«Стать администратором»**.
3. Введите секретный код — по умолчанию **`Your** (или то, что задали в `ADMIN_SECRET`).
4. После активации в меню появится раздел **Админ-панель**, где можно создавать, редактировать и удалять курсы.

---

## 🧪 Полезные команды

**Проверить, что код компилируется без ошибок:**
```bash
cd api-server-go
go vet ./...
```

**Собрать готовый исполняемый файл:**
```bash
cd api-server-go
go build -o marketingpro .
./marketingpro
```

**Посмотреть, какие таблицы создались:**
```bash
psql -U mpuser -d marketingpro -c "\dt"
```

**Очистить базу (удалить все данные) и начать сначала:**
```sql
psql -U postgres
DROP DATABASE marketingpro;
CREATE DATABASE marketingpro;
GRANT ALL PRIVILEGES ON DATABASE marketingpro TO mpuser;
\c marketingpro
GRANT ALL ON SCHEMA public TO mpuser;
\q
```
При следующем запуске сервер снова всё создаст и наполнит.

---

## ❓ Частые проблемы

| Ошибка | Причина и решение |
|--------|-------------------|
| `DATABASE_URL is not set` | Не задана переменная окружения. Вернитесь к шагу 3. |
| `connection refused` | PostgreSQL не запущен. Запустите службу. |
| `password authentication failed` | Неверный пароль в `DATABASE_URL`. Проверьте логин/пароль. |
| `permission denied for schema public` | Не выполнили `GRANT ALL ON SCHEMA public TO mpuser;` — повторите шаг 2. |
| `port already in use` | Порт 3000 занят. Поменяйте `PORT=3001`. |
| `go: command not found` | Не установлен Go. Установите его — шаг «Что вам понадобится». |

---

## 📞 Готово!

Всё, проект работает. Открывайте <http://localhost:3000> и пользуйтесь платформой **МаркетингПро**.
