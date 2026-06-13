package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

// adminSecretCode возвращает секрет для активации прав администратора.
// Берётся из переменной окружения ADMIN_SECRET, по умолчанию "admin123".
func adminSecretCode() string {
	s := os.Getenv("ADMIN_SECRET")
	if s == "" {
		return "admin123"
	}
	return s
}

// RegisterAdminRoutes регистрирует /api/admin/* маршруты.
func RegisterAdminRoutes(api fiber.Router, pool *pgxpool.Pool, store *session.Store) {
	secret := adminSecretCode()

	// POST /api/admin/become — активация прав по секретному коду.
	api.Post("/admin/become", func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}
		var body struct {
			Secret string `json:"secret"`
		}
		if err := c.BodyParser(&body); err != nil || body.Secret != secret {
			return c.Status(403).JSON(fiber.Map{"error": "Неверный секретный код"})
		}
		_, err := pool.Exec(context.Background(),
			"UPDATE users SET is_admin = true WHERE id = $1", userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка обновления"})
		}
		return c.JSON(fiber.Map{"message": "Права администратора активированы"})
	})

	// Middleware: проверяет, что пользователь — администратор.
	adminOnly := func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}
		var isAdmin bool
		err := pool.QueryRow(context.Background(),
			"SELECT is_admin FROM users WHERE id = $1", userID).Scan(&isAdmin)
		if err != nil || !isAdmin {
			return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
		}
		return c.Next()
	}

	admin := api.Group("/admin", adminOnly)

	// GET /api/admin/courses — список всех курсов для админ-панели.
	admin.Get("/courses", func(c *fiber.Ctx) error {
		rows, err := pool.Query(context.Background(),
			fmt.Sprintf("SELECT %s FROM courses ORDER BY created_at DESC", courseColumns))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения курсов"})
		}
		defer rows.Close()
		courses := []Course{}
		for rows.Next() {
			if course, err := scanCourse(rows); err == nil {
				courses = append(courses, course)
			}
		}
		return c.JSON(courses)
	})

	// POST /api/admin/courses — создать новый курс.
	admin.Post("/courses", func(c *fiber.Ctx) error {
		var body struct {
			Title        string   `json:"title"`
			Description  string   `json:"description"`
			Category     string   `json:"category"`
			IsPaid       bool     `json:"isPaid"`
			Price        *float64 `json:"price"`
			ImageURL     *string  `json:"imageUrl"`
			AuthorName   string   `json:"authorName"`
			Duration     int      `json:"duration"`
			LessonsCount int      `json:"lessonsCount"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
		}
		if body.Title == "" || body.AuthorName == "" || body.Category == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Название, категория и преподаватель обязательны"})
		}
		course, err := scanCourseRow(pool.QueryRow(context.Background(),
			fmt.Sprintf(`INSERT INTO courses (title, description, category, is_paid, price, image_url, author_name, duration, lessons_count)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING %s`, courseColumns),
			body.Title, body.Description, body.Category, body.IsPaid, body.Price,
			body.ImageURL, body.AuthorName, body.Duration, body.LessonsCount))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка создания курса"})
		}
		return c.Status(201).JSON(course)
	})

	// PUT /api/admin/courses/:id — обновить существующий курс.
	admin.Put("/courses/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный ID курса"})
		}
		var body struct {
			Title        *string  `json:"title"`
			Description  *string  `json:"description"`
			Category     *string  `json:"category"`
			IsPaid       *bool    `json:"isPaid"`
			Price        *float64 `json:"price"`
			ImageURL     *string  `json:"imageUrl"`
			AuthorName   *string  `json:"authorName"`
			Duration     *int     `json:"duration"`
			LessonsCount *int     `json:"lessonsCount"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
		}
		setClauses := []string{}
		args := []interface{}{}
		argIdx := 1
		add := func(col string, v interface{}) {
			setClauses = append(setClauses, fmt.Sprintf("%s=$%d", col, argIdx))
			args = append(args, v)
			argIdx++
		}
		if body.Title != nil {
			add("title", *body.Title)
		}
		if body.Description != nil {
			add("description", *body.Description)
		}
		if body.Category != nil {
			add("category", *body.Category)
		}
		if body.IsPaid != nil {
			add("is_paid", *body.IsPaid)
		}
		if body.Price != nil {
			add("price", *body.Price)
		}
		if body.ImageURL != nil {
			add("image_url", *body.ImageURL)
		}
		if body.AuthorName != nil {
			add("author_name", *body.AuthorName)
		}
		if body.Duration != nil {
			add("duration", *body.Duration)
		}
		if body.LessonsCount != nil {
			add("lessons_count", *body.LessonsCount)
		}
		if len(setClauses) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Нет данных для обновления"})
		}
		args = append(args, id)
		query := fmt.Sprintf("UPDATE courses SET %s WHERE id=$%d RETURNING %s",
			strings.Join(setClauses, ","), argIdx, courseColumns)
		course, err := scanCourseRow(pool.QueryRow(context.Background(), query, args...))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка обновления курса"})
		}
		return c.JSON(course)
	})

	// DELETE /api/admin/courses/:id — удалить курс.
	admin.Delete("/courses/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный ID"})
		}
		_, err = pool.Exec(context.Background(), "DELETE FROM courses WHERE id=$1", id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка удаления"})
		}
		return c.JSON(fiber.Map{"message": "Курс удалён"})
	})
}
