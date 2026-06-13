package main

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterAuthRoutes регистрирует маршруты регистрации, входа, выхода и /auth/me.
func RegisterAuthRoutes(api fiber.Router, pool *pgxpool.Pool, store *session.Store) {

	// POST /api/auth/register — создание нового пользователя.
	api.Post("/auth/register", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
		}
		if body.Email == "" || body.Password == "" || body.Name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Все поля обязательны"})
		}

		var existingID int
		err := pool.QueryRow(context.Background(),
			"SELECT id FROM users WHERE email = $1", body.Email).Scan(&existingID)
		if err == nil {
			return c.Status(409).JSON(fiber.Map{"error": "Email уже зарегистрирован"})
		}

		hash := hashPassword(body.Password)
		var user User
		var createdAt time.Time
		err = pool.QueryRow(context.Background(),
			`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
			 RETURNING id, email, name, avatar_url, date_of_birth, is_admin, created_at`,
			body.Email, hash, body.Name,
		).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.DateOfBirth, &user.IsAdmin, &createdAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка создания пользователя"})
		}
		user.CreatedAt = createdAt.UTC().Format(time.RFC3339)

		sess, _ := store.Get(c)
		sess.Set("userId", user.ID)
		_ = sess.Save()

		return c.Status(201).JSON(fiber.Map{
			"user":    user,
			"message": "Регистрация успешна",
		})
	})

	// POST /api/auth/login — вход по email + пароль.
	api.Post("/auth/login", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
		}

		hash := hashPassword(body.Password)
		var user User
		var storedHash string
		var createdAt time.Time
		err := pool.QueryRow(context.Background(),
			`SELECT id, email, name, avatar_url, date_of_birth, is_admin, created_at, password_hash
			 FROM users WHERE email = $1`, body.Email,
		).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.DateOfBirth, &user.IsAdmin, &createdAt, &storedHash)
		if err != nil || storedHash != hash {
			return c.Status(401).JSON(fiber.Map{"error": "Неверный email или пароль"})
		}
		user.CreatedAt = createdAt.UTC().Format(time.RFC3339)

		sess, _ := store.Get(c)
		sess.Set("userId", user.ID)
		_ = sess.Save()

		return c.JSON(fiber.Map{
			"user":    user,
			"message": "Вход выполнен успешно",
		})
	})

	// POST /api/auth/logout — выход (уничтожение сессии).
	api.Post("/auth/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err == nil {
			_ = sess.Destroy()
		}
		return c.JSON(fiber.Map{"message": "Выход выполнен"})
	})

	// GET /api/auth/me — данные текущего пользователя.
	api.Get("/auth/me", func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}
		var user User
		var createdAt time.Time
		err := pool.QueryRow(context.Background(),
			`SELECT id, email, name, avatar_url, date_of_birth, is_admin, created_at FROM users WHERE id = $1`, userID,
		).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.DateOfBirth, &user.IsAdmin, &createdAt)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Пользователь не найден"})
		}
		user.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		return c.JSON(user)
	})
}
