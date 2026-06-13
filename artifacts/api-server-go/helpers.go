package main

import (
        "crypto/sha256"
        "encoding/hex"
        "fmt"
        "os"
        "strconv"
        "time"

        "github.com/gofiber/fiber/v2"
        "github.com/gofiber/fiber/v2/middleware/session"
        "github.com/jackc/pgx/v5"
)

// buildDBURLFromParts собирает DATABASE_URL из отдельных переменных окружения.
// Возвращает пустую строку, если не задан хотя бы DB_USER или DB_NAME.
func buildDBURLFromParts() string {
        user := os.Getenv("DB_USER")
        name := os.Getenv("DB_NAME")
        if user == "" || name == "" {
                return ""
        }
        host := os.Getenv("DB_HOST")
        if host == "" {
                host = "localhost"
        }
        port := os.Getenv("DB_PORT")
        if port == "" {
                port = "5432"
        }
        password := os.Getenv("DB_PASSWORD")
        sslmode := os.Getenv("DB_SSLMODE")
        if sslmode == "" {
                sslmode = "disable"
        }
        return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
                user, password, host, port, name, sslmode)
}

// Соль для хеширования паролей. В production вынесите в переменную окружения.
const saltedPrefix = "rosstat_salt_2024"

// hashPassword — детерминированный SHA256-хеш пароля с солью.
func hashPassword(password string) string {
        h := sha256.Sum256([]byte(password + saltedPrefix))
        return hex.EncodeToString(h[:])
}

// weekBounds возвращает начало (понедельник 00:00) и конец (воскресенье 23:59) текущей недели.
func weekBounds() (start, end time.Time) {
        now := time.Now()
        weekday := int(now.Weekday())
        if weekday == 0 {
                weekday = 7
        }
        start = now.AddDate(0, 0, -(weekday - 1))
        start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
        end = start.AddDate(0, 0, 6)
        end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999000000, end.Location())
        return
}

// courseColumns — список колонок для SELECT-запросов по курсам.
const courseColumns = `id, title, description, category, is_paid, price, discount_price, image_url, author_name, duration, lessons_count, created_at`

// scanCourse сканирует строку из результата pgx.Rows в структуру Course.
func scanCourse(rows pgx.Rows) (Course, error) {
        var c Course
        var priceStr, discountPriceStr *string
        var createdAt time.Time
        err := rows.Scan(
                &c.ID, &c.Title, &c.Description, &c.Category,
                &c.IsPaid, &priceStr, &discountPriceStr,
                &c.ImageURL, &c.AuthorName, &c.Duration, &c.LessonsCount, &createdAt,
        )
        if err != nil {
                return c, err
        }
        if priceStr != nil {
                v, _ := strconv.ParseFloat(*priceStr, 64)
                c.Price = &v
        }
        if discountPriceStr != nil {
                v, _ := strconv.ParseFloat(*discountPriceStr, 64)
                c.DiscountPrice = &v
        }
        c.CreatedAt = createdAt.UTC().Format(time.RFC3339)
        return c, nil
}

// scanCourseRow сканирует одну строку pgx.Row в структуру Course.
func scanCourseRow(row pgx.Row) (Course, error) {
        var c Course
        var priceStr, discountPriceStr *string
        var createdAt time.Time
        err := row.Scan(
                &c.ID, &c.Title, &c.Description, &c.Category,
                &c.IsPaid, &priceStr, &discountPriceStr,
                &c.ImageURL, &c.AuthorName, &c.Duration, &c.LessonsCount, &createdAt,
        )
        if err != nil {
                return c, err
        }
        if priceStr != nil {
                v, _ := strconv.ParseFloat(*priceStr, 64)
                c.Price = &v
        }
        if discountPriceStr != nil {
                v, _ := strconv.ParseFloat(*discountPriceStr, 64)
                c.DiscountPrice = &v
        }
        c.CreatedAt = createdAt.UTC().Format(time.RFC3339)
        return c, nil
}

// getUserID извлекает ID авторизованного пользователя из сессии.
func getUserID(store *session.Store, c *fiber.Ctx) (int, bool) {
        sess, err := store.Get(c)
        if err != nil {
                return 0, false
        }
        uid := sess.Get("userId")
        if uid == nil {
                return 0, false
        }
        id, ok := uid.(int)
        return id, ok
}
