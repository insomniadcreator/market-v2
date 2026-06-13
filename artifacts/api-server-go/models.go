package main

// User — модель пользователя платформы.
type User struct {
	ID          int     `json:"id"`
	Email       string  `json:"email"`
	Name        string  `json:"name"`
	AvatarURL   *string `json:"avatarUrl"`
	DateOfBirth *string `json:"dateOfBirth"`
	IsAdmin     bool    `json:"isAdmin"`
	CreatedAt   string  `json:"createdAt"`
}

// Course — модель курса.
type Course struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	IsPaid        bool     `json:"isPaid"`
	Price         *float64 `json:"price"`
	DiscountPrice *float64 `json:"discountPrice"`
	ImageURL      *string  `json:"imageUrl"`
	AuthorName    string   `json:"authorName"`
	Duration      int      `json:"duration"`
	LessonsCount  int      `json:"lessonsCount"`
	CreatedAt     string   `json:"createdAt"`
}

// CourseWithLessons — курс с включёнными уроками (для страницы курса).
type CourseWithLessons struct {
	Course
	Lessons []Lesson `json:"lessons"`
}

// Lesson — модель урока.
type Lesson struct {
	ID          int     `json:"id"`
	CourseID    int     `json:"courseId"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	VideoURL    *string `json:"videoUrl"`
	Duration    int     `json:"duration"`
	Order       int     `json:"order"`
	IsFree      bool    `json:"isFree"`
}

// Enrollment — запись пользователя на курс.
type Enrollment struct {
	ID               int    `json:"id"`
	UserID           int    `json:"userId"`
	CourseID         int    `json:"courseId"`
	Course           Course `json:"course"`
	CompletedLessons int    `json:"completedLessons"`
	TotalLessons     int    `json:"totalLessons"`
	EnrolledAt       string `json:"enrolledAt"`
}

// DayProgress — прогресс за один день недели.
type DayProgress struct {
	Day              string `json:"day"`
	LessonsCompleted int    `json:"lessonsCompleted"`
}

// WeeklyProgress — недельная статистика.
type WeeklyProgress struct {
	WeekStart      string        `json:"weekStart"`
	WeekEnd        string        `json:"weekEnd"`
	Days           []DayProgress `json:"days"`
	TotalCompleted int           `json:"totalCompleted"`
}

// Dashboard — данные для главной страницы пользователя.
type Dashboard struct {
	User             User           `json:"user"`
	Enrollments      []Enrollment   `json:"enrollments"`
	ContinueWatching []Enrollment   `json:"continueWatching"`
	WeeklyProgress   WeeklyProgress `json:"weeklyProgress"`
	FeaturedCourses  []Course       `json:"featuredCourses"`
}

// LessonProgress — прогресс прохождения отдельного урока.
type LessonProgress struct {
	ID          int     `json:"id"`
	UserID      int     `json:"userId"`
	LessonID    int     `json:"lessonId"`
	CourseID    int     `json:"courseId"`
	Completed   bool    `json:"completed"`
	CompletedAt *string `json:"completedAt"`
}
