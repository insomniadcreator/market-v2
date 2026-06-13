package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterDashboardRoutes регистрирует маршруты записей на курс, прогресса и сводки на главной.
func RegisterDashboardRoutes(api fiber.Router, pool *pgxpool.Pool, store *session.Store) {

	// GET /api/enrollments — все курсы пользователя.
	api.Get("/enrollments", func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}

		rows, err := pool.Query(context.Background(),
			fmt.Sprintf(`SELECT e.id, e.user_id, e.course_id, e.enrolled_at, c.%s
				FROM enrollments e
				JOIN courses c ON e.course_id = c.id
				WHERE e.user_id = $1`, courseColumns),
			userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения записей"})
		}
		defer rows.Close()

		type enrollRow struct {
			id, userID, courseID int
			enrolledAt           time.Time
			course               Course
		}

		var eRows []enrollRow
		for rows.Next() {
			var er enrollRow
			var priceStr, discountStr *string
			var courseCreatedAt time.Time
			err := rows.Scan(
				&er.id, &er.userID, &er.courseID, &er.enrolledAt,
				&er.course.ID, &er.course.Title, &er.course.Description, &er.course.Category,
				&er.course.IsPaid, &priceStr, &discountStr,
				&er.course.ImageURL, &er.course.AuthorName, &er.course.Duration, &er.course.LessonsCount, &courseCreatedAt,
			)
			if err != nil {
				continue
			}
			if priceStr != nil {
				v, _ := strconv.ParseFloat(*priceStr, 64)
				er.course.Price = &v
			}
			if discountStr != nil {
				v, _ := strconv.ParseFloat(*discountStr, 64)
				er.course.DiscountPrice = &v
			}
			er.course.CreatedAt = courseCreatedAt.UTC().Format(time.RFC3339)
			eRows = append(eRows, er)
		}

		enrollments := []Enrollment{}
		for _, er := range eRows {
			var cnt int
			_ = pool.QueryRow(context.Background(),
				`SELECT COUNT(*) FROM lesson_progress
				 WHERE user_id = $1 AND course_id = $2 AND completed = true`,
				userID, er.courseID).Scan(&cnt)

			enrollments = append(enrollments, Enrollment{
				ID:               er.id,
				UserID:           er.userID,
				CourseID:         er.courseID,
				Course:           er.course,
				CompletedLessons: cnt,
				TotalLessons:     er.course.LessonsCount,
				EnrolledAt:       er.enrolledAt.UTC().Format(time.RFC3339),
			})
		}
		return c.JSON(enrollments)
	})

	// POST /api/enrollments — записать пользователя на курс.
	api.Post("/enrollments", func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}

		var body struct {
			CourseID int `json:"courseId"`
		}
		if err := c.BodyParser(&body); err != nil || body.CourseID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
		}

		var existingID int
		err := pool.QueryRow(context.Background(),
			"SELECT id FROM enrollments WHERE user_id = $1 AND course_id = $2",
			userID, body.CourseID).Scan(&existingID)
		if err == nil {
			return c.Status(409).JSON(fiber.Map{"error": "Вы уже записаны на этот курс"})
		}

		course, err := scanCourseRow(pool.QueryRow(context.Background(),
			fmt.Sprintf("SELECT %s FROM courses WHERE id = $1", courseColumns), body.CourseID))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Курс не найден"})
		}

		var enrollID int
		var enrolledAt time.Time
		err = pool.QueryRow(context.Background(),
			"INSERT INTO enrollments (user_id, course_id) VALUES ($1, $2) RETURNING id, enrolled_at",
			userID, body.CourseID).Scan(&enrollID, &enrolledAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка записи на курс"})
		}

		return c.Status(201).JSON(Enrollment{
			ID:               enrollID,
			UserID:           userID,
			CourseID:         body.CourseID,
			Course:           course,
			CompletedLessons: 0,
			TotalLessons:     course.LessonsCount,
			EnrolledAt:       enrolledAt.UTC().Format(time.RFC3339),
		})
	})

	// GET /api/progress — недельная статистика по дням.
	api.Get("/progress", func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}

		start, end := weekBounds()
		rows, err := pool.Query(context.Background(),
			`SELECT completed_at FROM lesson_progress
			 WHERE user_id = $1 AND completed = true
			   AND completed_at >= $2 AND completed_at <= $3`,
			userID, start, end)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения прогресса"})
		}
		defer rows.Close()

		dayCounts := map[int]int{0: 0, 1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0}
		total := 0
		for rows.Next() {
			var completedAt time.Time
			if err := rows.Scan(&completedAt); err != nil {
				continue
			}
			dayCounts[int(completedAt.Weekday())]++
			total++
		}

		dayNames := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
		orderedDays := []int{1, 2, 3, 4, 5, 6, 0}
		days := make([]DayProgress, 0, 7)
		for _, d := range orderedDays {
			days = append(days, DayProgress{
				Day:              dayNames[d],
				LessonsCompleted: dayCounts[d],
			})
		}

		return c.JSON(fiber.Map{
			"weekStart":      start.UTC().Format(time.RFC3339),
			"weekEnd":        end.UTC().Format(time.RFC3339),
			"days":           days,
			"totalCompleted": total,
		})
	})

	// POST /api/progress — пометить урок как пройденный / не пройденный.
	api.Post("/progress", func(c *fiber.Ctx) error {
		userID, ok := getUserID(store, c)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Не авторизован"})
		}

		var body struct {
			LessonID  int  `json:"lessonId"`
			CourseID  int  `json:"courseId"`
			Completed bool `json:"completed"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
		}

		var existingID int
		err := pool.QueryRow(context.Background(),
			"SELECT id FROM lesson_progress WHERE user_id = $1 AND lesson_id = $2",
			userID, body.LessonID).Scan(&existingID)

		var rec LessonProgress
		var completedAt *time.Time
		now := time.Now()
		var ca *time.Time
		if body.Completed {
			ca = &now
		}

		if err == nil {
			err = pool.QueryRow(context.Background(),
				`UPDATE lesson_progress SET completed = $1, completed_at = $2
				 WHERE user_id = $3 AND lesson_id = $4
				 RETURNING id, user_id, lesson_id, course_id, completed, completed_at`,
				body.Completed, ca, userID, body.LessonID,
			).Scan(&rec.ID, &rec.UserID, &rec.LessonID, &rec.CourseID, &rec.Completed, &completedAt)
		} else {
			err = pool.QueryRow(context.Background(),
				`INSERT INTO lesson_progress (user_id, lesson_id, course_id, completed, completed_at)
				 VALUES ($1, $2, $3, $4, $5)
				 RETURNING id, user_id, lesson_id, course_id, completed, completed_at`,
				userID, body.LessonID, body.CourseID, body.Completed, ca,
			).Scan(&rec.ID, &rec.UserID, &rec.LessonID, &rec.CourseID, &rec.Completed, &completedAt)
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка обновления прогресса"})
		}

		if completedAt != nil {
			s := completedAt.UTC().Format(time.RFC3339)
			rec.CompletedAt = &s
		}
		return c.JSON(rec)
	})

	// GET /api/dashboard — сводка для главной страницы.
	api.Get("/dashboard", func(c *fiber.Ctx) error {
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

		// Курсы пользователя.
		rows, err := pool.Query(context.Background(),
			fmt.Sprintf(`SELECT e.id, e.user_id, e.course_id, e.enrolled_at, c.%s
				FROM enrollments e JOIN courses c ON e.course_id = c.id
				WHERE e.user_id = $1`, courseColumns),
			userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения записей"})
		}

		type enrollRow struct {
			id, courseID int
			enrolledAt   time.Time
			course       Course
		}
		var eRows []enrollRow
		for rows.Next() {
			var er enrollRow
			var priceStr, discountStr *string
			var cc time.Time
			if err := rows.Scan(
				&er.id, new(int), &er.courseID, &er.enrolledAt,
				&er.course.ID, &er.course.Title, &er.course.Description, &er.course.Category,
				&er.course.IsPaid, &priceStr, &discountStr,
				&er.course.ImageURL, &er.course.AuthorName, &er.course.Duration, &er.course.LessonsCount, &cc,
			); err != nil {
				continue
			}
			if priceStr != nil {
				v, _ := strconv.ParseFloat(*priceStr, 64)
				er.course.Price = &v
			}
			if discountStr != nil {
				v, _ := strconv.ParseFloat(*discountStr, 64)
				er.course.DiscountPrice = &v
			}
			er.course.CreatedAt = cc.UTC().Format(time.RFC3339)
			eRows = append(eRows, er)
		}
		rows.Close()

		enrollments := []Enrollment{}
		for _, er := range eRows {
			var cnt int
			_ = pool.QueryRow(context.Background(),
				`SELECT COUNT(*) FROM lesson_progress WHERE user_id=$1 AND course_id=$2 AND completed=true`,
				userID, er.courseID).Scan(&cnt)
			enrollments = append(enrollments, Enrollment{
				ID: er.id, UserID: userID, CourseID: er.courseID,
				Course: er.course, CompletedLessons: cnt,
				TotalLessons: er.course.LessonsCount,
				EnrolledAt:   er.enrolledAt.UTC().Format(time.RFC3339),
			})
		}

		// Недельный прогресс.
		start, end := weekBounds()
		progRows, err := pool.Query(context.Background(),
			`SELECT completed_at FROM lesson_progress
			 WHERE user_id=$1 AND completed=true AND completed_at>=$2 AND completed_at<=$3`,
			userID, start, end)

		dayCounts := map[int]int{0: 0, 1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0}
		totalCompleted := 0
		if err == nil {
			for progRows.Next() {
				var ca time.Time
				if err := progRows.Scan(&ca); err == nil {
					dayCounts[int(ca.Weekday())]++
					totalCompleted++
				}
			}
			progRows.Close()
		}

		dayNames := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
		orderedDays := []int{1, 2, 3, 4, 5, 6, 0}
		days := make([]DayProgress, 0, 7)
		for _, d := range orderedDays {
			days = append(days, DayProgress{Day: dayNames[d], LessonsCompleted: dayCounts[d]})
		}

		// Рекомендуемые курсы.
		featRows, _ := pool.Query(context.Background(),
			fmt.Sprintf("SELECT %s FROM courses ORDER BY created_at DESC LIMIT 6", courseColumns))
		featuredCourses := []Course{}
		if featRows != nil {
			for featRows.Next() {
				if fc, err := scanCourse(featRows); err == nil {
					featuredCourses = append(featuredCourses, fc)
				}
			}
			featRows.Close()
		}

		// "Продолжить обучение" — курсы с незавершёнными уроками (макс. 3).
		continueWatching := []Enrollment{}
		for _, e := range enrollments {
			if e.CompletedLessons < e.TotalLessons {
				continueWatching = append(continueWatching, e)
				if len(continueWatching) == 3 {
					break
				}
			}
		}

		return c.JSON(Dashboard{
			User:             user,
			Enrollments:      enrollments,
			ContinueWatching: continueWatching,
			WeeklyProgress: WeeklyProgress{
				WeekStart:      start.UTC().Format(time.RFC3339),
				WeekEnd:        end.UTC().Format(time.RFC3339),
				Days:           days,
				TotalCompleted: totalCompleted,
			},
			FeaturedCourses: featuredCourses,
		})
	})
}
