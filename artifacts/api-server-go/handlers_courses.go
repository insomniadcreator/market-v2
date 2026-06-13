package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterCourseRoutes регистрирует публичные маршруты для каталога курсов и уроков.
func RegisterCourseRoutes(api fiber.Router, pool *pgxpool.Pool) {

	// GET /api/courses — список курсов с фильтрами по категории, типу (платный/бесплатный) и поиску.
	api.Get("/courses", func(c *fiber.Ctx) error {
		category := c.Query("category")
		isPaidStr := c.Query("isPaid")
		search := c.Query("search")

		query := fmt.Sprintf("SELECT %s FROM courses", courseColumns)
		args := []interface{}{}
		conditions := []string{}
		argIdx := 1

		if category != "" {
			conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
			args = append(args, category)
			argIdx++
		}
		if isPaidStr == "true" {
			conditions = append(conditions, fmt.Sprintf("is_paid = $%d", argIdx))
			args = append(args, true)
			argIdx++
		} else if isPaidStr == "false" {
			conditions = append(conditions, fmt.Sprintf("is_paid = $%d", argIdx))
			args = append(args, false)
			argIdx++
		}
		if search != "" {
			conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIdx))
			args = append(args, "%"+search+"%")
			argIdx++
		}

		if len(conditions) > 0 {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}
		query += " ORDER BY created_at DESC"

		rows, err := pool.Query(context.Background(), query, args...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения курсов"})
		}
		defer rows.Close()

		courses := []Course{}
		for rows.Next() {
			course, err := scanCourse(rows)
			if err != nil {
				continue
			}
			courses = append(courses, course)
		}
		return c.JSON(courses)
	})

	// GET /api/courses/:id — детальная информация о курсе со всеми уроками.
	api.Get("/courses/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный ID курса"})
		}

		course, err := scanCourseRow(pool.QueryRow(context.Background(),
			fmt.Sprintf("SELECT %s FROM courses WHERE id = $1", courseColumns), id))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Курс не найден"})
		}

		rows, err := pool.Query(context.Background(),
			`SELECT id, course_id, title, description, video_url, duration, "order", is_free
			 FROM lessons WHERE course_id = $1 ORDER BY "order"`, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения уроков"})
		}
		defer rows.Close()

		lessons := []Lesson{}
		for rows.Next() {
			var l Lesson
			if err := rows.Scan(&l.ID, &l.CourseID, &l.Title, &l.Description,
				&l.VideoURL, &l.Duration, &l.Order, &l.IsFree); err != nil {
				continue
			}
			lessons = append(lessons, l)
		}

		return c.JSON(CourseWithLessons{Course: course, Lessons: lessons})
	})

	// GET /api/courses/:id/lessons — отдельный список уроков для курса.
	api.Get("/courses/:id/lessons", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный ID курса"})
		}

		rows, err := pool.Query(context.Background(),
			`SELECT id, course_id, title, description, video_url, duration, "order", is_free
			 FROM lessons WHERE course_id = $1 ORDER BY "order"`, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения уроков"})
		}
		defer rows.Close()

		lessons := []Lesson{}
		for rows.Next() {
			var l Lesson
			if err := rows.Scan(&l.ID, &l.CourseID, &l.Title, &l.Description,
				&l.VideoURL, &l.Duration, &l.Order, &l.IsFree); err != nil {
				continue
			}
			lessons = append(lessons, l)
		}
		return c.JSON(lessons)
	})
}
