package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectDB подключается к PostgreSQL по строке dbURL и проверяет соединение.
func connectDB(dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// runMigrations создаёт все необходимые таблицы (если их ещё нет)
// и выполняет ALTER TABLE для добавления новых колонок.
func runMigrations(pool *pgxpool.Pool) {
	statements := []string{
		// users
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			avatar_url TEXT,
			date_of_birth TEXT,
			is_admin BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// courses
		`CREATE TABLE IF NOT EXISTS courses (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL,
			is_paid BOOLEAN NOT NULL DEFAULT false,
			price NUMERIC(10,2),
			discount_price NUMERIC(10,2),
			image_url TEXT,
			author_name TEXT NOT NULL,
			duration INTEGER NOT NULL DEFAULT 0,
			lessons_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// lessons
		`CREATE TABLE IF NOT EXISTS lessons (
			id SERIAL PRIMARY KEY,
			course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT,
			video_url TEXT,
			duration INTEGER NOT NULL DEFAULT 0,
			"order" INTEGER NOT NULL DEFAULT 0,
			is_free BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// enrollments
		`CREATE TABLE IF NOT EXISTS enrollments (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
			enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (user_id, course_id)
		)`,
		// lesson_progress
		`CREATE TABLE IF NOT EXISTS lesson_progress (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			lesson_id INTEGER NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
			course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
			completed BOOLEAN NOT NULL DEFAULT false,
			completed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (user_id, lesson_id)
		)`,
		// На случай старой схемы — добавим is_admin, если его нет.
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT false`,
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(context.Background(), stmt); err != nil {
			log.Printf("Migration warning: %v", err)
		}
	}
}

// seedCourses загружает стартовый набор курсов и уроков, если таблица courses пуста.
func seedCourses(pool *pgxpool.Pool) {
	var count int
	if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM courses").Scan(&count); err != nil {
		log.Printf("Seed check failed: %v", err)
		return
	}
	if count > 0 {
		return
	}
	log.Println("Загрузка начальных данных (курсы и уроки)...")

	type seedLesson struct {
		Title       string
		Description string
		Duration    int
		IsFree      bool
	}
	type seedCourse struct {
		Title        string
		Description  string
		Category     string
		IsPaid       bool
		Price        *float64
		AuthorName   string
		Duration     int
		LessonsCount int
		Lessons      []seedLesson
	}
	p := func(v float64) *float64 { return &v }

	data := []seedCourse{
		{
			Title: "Основы цифрового маркетинга",
			Description: "Полный курс по основам цифрового маркетинга: SEO, контекстная реклама, " +
				"социальные сети и email-маркетинг. Идеально для начинающих.",
			Category: "Цифровой маркетинг", IsPaid: false, Price: nil,
			AuthorName: "Анна Смирнова", Duration: 480, LessonsCount: 12,
			Lessons: []seedLesson{
				{"Что такое цифровой маркетинг?", "Обзор основных каналов и инструментов цифрового маркетинга.", 25, true},
				{"SEO: основы поисковой оптимизации", "Как работают поисковые системы и базовые принципы SEO.", 30, true},
				{"Контекстная реклама (Google Ads)", "Запуск первой рекламной кампании в Google Ads.", 45, false},
				{"Маркетинг в социальных сетях", "Стратегии продвижения в ВКонтакте, Instagram и Telegram.", 40, false},
			},
		},
		{
			Title: "SMM для бизнеса",
			Description: "Научитесь создавать эффективные стратегии в социальных сетях, создавать вирусный контент " +
				"и продвигать бренд в Instagram, ВКонтакте и Telegram.",
			Category: "Социальные сети", IsPaid: true, Price: p(2990),
			AuthorName: "Иван Петров", Duration: 360, LessonsCount: 10,
			Lessons: []seedLesson{
				{"Стратегия SMM: с чего начать?", "Как разработать SMM-стратегию для бизнеса.", 35, true},
				{"Контент-план: как его составить", "Планирование публикаций и создание редакционного календаря.", 30, false},
				{"Работа с Instagram и Reels", "Создание виральных коротких видео.", 40, false},
			},
		},
		{
			Title: "SEO-продвижение сайтов",
			Description: "Практический курс по поисковой оптимизации: технический SEO, работа с ключевыми словами, " +
				"наращивание ссылочной массы и аналитика.",
			Category: "SEO", IsPaid: true, Price: p(3990),
			AuthorName: "Мария Иванова", Duration: 540, LessonsCount: 15,
			Lessons: []seedLesson{
				{"Технический SEO", "Скорость сайта, структура URL и мобильная оптимизация.", 50, true},
				{"Ключевые слова и семантическое ядро", "Сбор и кластеризация ключевых слов.", 45, false},
			},
		},
		{
			Title: "Контент-маркетинг",
			Description: "Создание и распространение ценного контента для привлечения целевой аудитории. " +
				"Блоги, видео, подкасты и инфографика.",
			Category: "Контент", IsPaid: false, Price: nil,
			AuthorName: "Дмитрий Козлов", Duration: 300, LessonsCount: 8,
			Lessons: []seedLesson{
				{"Что такое контент-маркетинг?", "Принципы создания контента, который продаёт.", 20, true},
				{"Блог как инструмент привлечения", "SEO-статьи, которые приводят органический трафик.", 35, false},
			},
		},
		{
			Title: "Email-маркетинг и автоматизация",
			Description: "Построение эффективных email-кампаний, сегментация аудитории, A/B тестирование " +
				"и автоматизация рассылок.",
			Category: "Email-маркетинг", IsPaid: true, Price: p(2490),
			AuthorName: "Елена Новикова", Duration: 420, LessonsCount: 11,
			Lessons: []seedLesson{
				{"Основы email-маркетинга", "Сбор базы подписчиков и первая рассылка.", 30, true},
				{"Сегментация аудитории", "Как разделить подписчиков на группы для точных рассылок.", 35, false},
			},
		},
		{
			Title: "Аналитика и метрики в маркетинге",
			Description: "Google Analytics, Яндекс.Метрика, отслеживание конверсий и построение отчётов " +
				"для принятия маркетинговых решений.",
			Category: "Аналитика", IsPaid: true, Price: p(3490),
			AuthorName: "Алексей Морозов", Duration: 480, LessonsCount: 14,
			Lessons: []seedLesson{
				{"Google Analytics 4: введение", "Настройка и основные отчёты GA4.", 40, true},
				{"Яндекс.Метрика", "Тепловые карты, вебвизор и воронки.", 35, false},
			},
		},
	}

	for _, c := range data {
		var courseID int
		err := pool.QueryRow(context.Background(),
			`INSERT INTO courses (title, description, category, is_paid, price, author_name, duration, lessons_count)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
			c.Title, c.Description, c.Category, c.IsPaid, c.Price,
			c.AuthorName, c.Duration, c.LessonsCount,
		).Scan(&courseID)
		if err != nil {
			log.Printf("Seed course failed: %v", err)
			continue
		}
		for i, l := range c.Lessons {
			_, err := pool.Exec(context.Background(),
				`INSERT INTO lessons (course_id, title, description, duration, "order", is_free)
				 VALUES ($1,$2,$3,$4,$5,$6)`,
				courseID, l.Title, l.Description, l.Duration, i+1, l.IsFree)
			if err != nil {
				log.Printf("Seed lesson failed: %v", err)
			}
		}
	}
	log.Println("Начальные данные загружены успешно")
}
