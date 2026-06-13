package main

import (
        "errors"
        "log"
        "os"
        "time"

        "github.com/gofiber/fiber/v2"
        "github.com/gofiber/fiber/v2/middleware/cors"
        fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
        "github.com/gofiber/fiber/v2/middleware/session"
        "github.com/joho/godotenv"
)

// main — точка входа: подключение к БД, миграции, регистрация маршрутов и запуск сервера.
func main() {
        // 0. Загружаем переменные окружения из файла .env (если он есть).
        // Ошибку игнорируем — на проде .env может отсутствовать, переменные задаются системой.
        if err := godotenv.Load(); err != nil {
                log.Println("Файл .env не найден — используются системные переменные окружения")
        } else {
                log.Println("Переменные окружения загружены из .env")
        }

        // 1. Подключение к PostgreSQL.
        // Поддерживаем два варианта: либо одна строка DATABASE_URL,
        // либо отдельные переменные DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME / DB_SSLMODE.
        dbURL := os.Getenv("DATABASE_URL")
        if dbURL == "" {
                dbURL = buildDBURLFromParts()
        }
        if dbURL == "" {
                log.Fatal("Не задано подключение к БД: укажите DATABASE_URL или DB_HOST/DB_USER/... в .env")
        }
        pool, err := connectDB(dbURL)
        if err != nil {
                log.Fatalf("Не удалось подключиться к базе данных: %v", err)
        }
        defer pool.Close()
        log.Println("Подключение к базе данных установлено")

        // 2. Миграции (создание таблиц при первом запуске) + начальные данные.
        runMigrations(pool)
        seedCourses(pool)

        // 3. Каталог для загруженных файлов.
        if err := os.MkdirAll("./static/uploads", 0755); err != nil {
                log.Printf("Не удалось создать каталог uploads: %v", err)
        }

        // 4. Создание Fiber-приложения.
        app := fiber.New(fiber.Config{
                ErrorHandler: func(c *fiber.Ctx, err error) error {
                        code := fiber.StatusInternalServerError
                        var e *fiber.Error
                        if errors.As(err, &e) {
                                code = e.Code
                        }
                        return c.Status(code).JSON(fiber.Map{"error": err.Error()})
                },
        })

        app.Use(fiberlogger.New())
        app.Use(cors.New(cors.Config{
                AllowOrigins:     "*",
                AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
                AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
                AllowCredentials: false,
        }))

        // 5. Сессии (хранятся в cookie mp_session, живут 7 дней).
        store := session.New(session.Config{
                Expiration:     7 * 24 * time.Hour,
                CookieSecure:   false,
                CookieHTTPOnly: true,
                KeyLookup:      "cookie:mp_session",
        })

        // 6. Регистрация всех маршрутов /api/*.
        api := app.Group("/api")

        api.Get("/health", func(c *fiber.Ctx) error {
                return c.JSON(fiber.Map{"status": "ok"})
        })

        RegisterAuthRoutes(api, pool, store)
        RegisterCourseRoutes(api, pool)
        RegisterDashboardRoutes(api, pool, store)
        RegisterProfileRoutes(api, pool, store)
        RegisterAdminRoutes(api, pool, store)

        // 7. Раздача статического фронтенда (vanilla JS SPA).
        app.Static("/", "./static")

        // SPA fallback — для всех не-API путей возвращаем index.html.
        app.Get("/*", func(c *fiber.Ctx) error {
                return c.SendFile("./static/index.html")
        })

        // 8. Запуск сервера.
        port := os.Getenv("PORT")
        if port == "" {
                port = "3000"
        }
        log.Printf("🚀 Сервер МаркетингПро запущен на порту %s", port)
        log.Fatal(app.Listen(":" + port))
}
