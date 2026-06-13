package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

// allowedImageExt — расширения файлов, разрешённые при загрузке.
var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// RegisterProfileRoutes регистрирует маршруты обновления профиля и загрузки файлов.
func RegisterProfileRoutes(api fiber.Router, pool *pgxpool.Pool, store *session.Store) {

	// PATCH /api/users/profile — обновление имени, аватара, даты рождения.
	api.Patch("/users/profile", func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}

		var body struct {
			Name        *string `json:"name"`
			AvatarURL   *string `json:"avatarUrl"`
			DateOfBirth *string `json:"dateOfBirth"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
		}

		setClauses := []string{}
		args := []interface{}{}
		argIdx := 1

		if body.Name != nil && *body.Name != "" {
			setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
			args = append(args, *body.Name)
			argIdx++
		}
		if body.AvatarURL != nil {
			setClauses = append(setClauses, fmt.Sprintf("avatar_url = $%d", argIdx))
			args = append(args, *body.AvatarURL)
			argIdx++
		}
		if body.DateOfBirth != nil {
			setClauses = append(setClauses, fmt.Sprintf("date_of_birth = $%d", argIdx))
			if *body.DateOfBirth == "" {
				args = append(args, nil)
			} else {
				args = append(args, *body.DateOfBirth)
			}
			argIdx++
		}

		if len(setClauses) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Нет данных для обновления"})
		}

		args = append(args, userID)
		query := fmt.Sprintf(
			`UPDATE users SET %s WHERE id = $%d
			 RETURNING id, email, name, avatar_url, date_of_birth, is_admin, created_at`,
			strings.Join(setClauses, ", "), argIdx,
		)

		var user User
		var createdAt time.Time
		err := pool.QueryRow(context.Background(), query, args...).
			Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.DateOfBirth, &user.IsAdmin, &createdAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка обновления профиля"})
		}
		user.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		return c.JSON(user)
	})

	// POST /api/upload — загрузка картинки (аватар, обложка курса).
	api.Post("/upload", func(c *fiber.Ctx) error {
		_, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Файл не найден"})
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedImageExt[ext] {
			return c.Status(400).JSON(fiber.Map{"error": "Допустимые форматы: jpg, png, gif, webp"})
		}
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := "./static/uploads/" + filename
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка сохранения файла"})
		}
		return c.JSON(fiber.Map{"url": "/uploads/" + filename})
	})
}
