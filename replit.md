# МаркетингПро — Онлайн курсы по маркетингу

## Overview

Educational web platform for marketing courses. Single binary Go + Fiber backend that serves both the REST API and the vanilla JS SPA frontend.

## Stack

- **Backend**: Go + Fiber v2 (HTTP server, REST API, session auth, static file serving)
- **Database**: PostgreSQL (direct pgx/pgxpool queries, no ORM)
- **Frontend**: Vanilla JS (ES6+), HTML5, CSS3 — no frameworks, no build step
- **Auth**: Session cookies (in-memory Fiber session store)

## Project Structure

```
artifacts/api-server-go/
├── main.go          # Single Go file — all routes + DB logic
├── static/
│   ├── index.html   # SPA entry point
│   ├── css/style.css
│   └── js/app.js    # Full SPA with hash-based routing
├── go.mod
└── go.sum
```

## Features

- Russian language interface
- Session-based auth (register / login / logout)
- Dashboard with weekly progress chart and enrollment overview
- Course catalog with search and filters (category, free/paid)
- Course detail page with lesson list and progress tracking
- Profile editing (name, date of birth, avatar URL)
- Hash-based client-side routing (#/, #/courses, #/courses/:id, #/profile)

## API Endpoints (all under /api)

- `GET  /health`             — Health check
- `POST /auth/register`      — Register
- `POST /auth/login`         — Login
- `POST /auth/logout`        — Logout
- `GET  /auth/me`            — Current user
- `GET  /courses`            — List courses (search, category, isPaid filters)
- `GET  /courses/:id`        — Course detail with lessons
- `GET  /enrollments`        — User enrollments
- `POST /enrollments`        — Enroll in course
- `GET  /progress`           — Weekly progress
- `POST /progress`           — Update lesson progress
- `GET  /dashboard`          — Dashboard aggregate data
- `PATCH /users/profile`     — Update user profile

## Database Schema

- `users`          — id, email, password_hash, name, avatar_url, date_of_birth, created_at
- `courses`        — id, title, description, category, is_paid, price, discount_price, image_url, author_name, duration, lessons_count
- `lessons`        — id, course_id, title, description, video_url, duration, order, is_free
- `enrollments`    — id, user_id, course_id, enrolled_at
- `lesson_progress`— id, user_id, lesson_id, course_id, completed, completed_at

## Key Commands

- `cd artifacts/api-server-go && PORT=3000 go run main.go` — run server
- `cd artifacts/api-server-go && go build -o server .` — build binary
