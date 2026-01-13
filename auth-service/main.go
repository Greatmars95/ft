package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Конфигурация
var (
	jwtSecret []byte
	db        *sql.DB
)

// initJWTSecret инициализация JWT секрета
func initJWTSecret() {
	secretStr := os.Getenv("JWT_SECRET")

	if secretStr == "" {
		log.Println("⚠️  JWT_SECRET не задан!")
		log.Println("⚠️  Генерируется случайный секрет (токены не будут работать после перезапуска)")

		// Генерируем случайный секрет
		randomBytes := make([]byte, 32)
		_, err := rand.Read(randomBytes)
		if err != nil {
			log.Fatal("❌ Не удалось сгенерировать случайный секрет:", err)
		}
		secretStr = base64.StdEncoding.EncodeToString(randomBytes)
		log.Printf("💡 Сгенерирован случайный JWT_SECRET. Для production установите переменную окружения JWT_SECRET")
	} else if secretStr == "your-super-secret-jwt-key-change-in-production" {
		log.Println("🚨 ВНИМАНИЕ: Используется дефолтный JWT_SECRET!")
		log.Println("🚨 Это небезопасно для production! Установите уникальный JWT_SECRET")
	}

	jwtSecret = []byte(secretStr)
	log.Println("✅ JWT секрет инициализирован")
}

// User модель пользователя
type User struct {
	ID           int        `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Не возвращаем хеш в JSON
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"` // По умолчанию "user"
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse ответ с токеном
type TokenResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}

// JWTClaims кастомные claims для JWT
type JWTClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	// Инициализация JWT секрета
	initJWTSecret()

	// Подключение к БД
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "admin")
	dbPassword := getEnv("DB_PASSWORD", "secret123")
	dbName := getEnv("DB_NAME", "quotopia")

	connStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser +
		" password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	// Проверка подключения
	if err = db.Ping(); err != nil {
		log.Fatalf("❌ БД не отвечает: %v", err)
	}
	log.Println("✅ Подключено к PostgreSQL")

	// Создание роутера
	r := gin.Default()

	// CORS middleware
	r.Use(corsMiddleware())

	// Публичные endpoints
	public := r.Group("/auth")
	{
		public.POST("/register", register)
		public.POST("/login", login)
		public.POST("/refresh", refreshToken)
	}

	// Защищённые endpoints
	protected := r.Group("/auth")
	protected.Use(authMiddleware())
	{
		protected.GET("/me", getCurrentUser)
		protected.POST("/logout", logout)
		protected.PUT("/password", changePassword)
	}

	// Admin endpoints
	admin := r.Group("/auth/admin")
	admin.Use(authMiddleware(), adminMiddleware())
	{
		admin.GET("/users", listUsers)
		admin.PUT("/users/:id/role", changeUserRole)
		admin.DELETE("/users/:id", deleteUser)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auth"})
	})

	port := getEnv("PORT", "8090")
	log.Printf("🚀 Auth Service запущен на порту %s", port)
	log.Printf("📚 Endpoints:")
	log.Printf("   POST   /auth/register    - Регистрация")
	log.Printf("   POST   /auth/login       - Вход")
	log.Printf("   GET    /auth/me          - Текущий пользователь")
	log.Printf("   POST   /auth/refresh     - Обновить токен")
	log.Printf("   GET    /health           - Health check")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

// register регистрация нового пользователя
func register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка существования пользователя
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Роль по умолчанию
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Только admin может создавать admin или trader
	if role == "admin" || role == "trader" {
		// Проверяем наличие токена авторизации
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required to create admin or trader"})
			return
		}

		// Убираем "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Парсим токен
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required to create admin or trader"})
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || claims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required to create admin or trader"})
			return
		}
	}

	// Создание пользователя
	var user User
	err = db.QueryRow(`
		INSERT INTO users (email, password_hash, role, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id, email, role, is_active, created_at
	`, req.Email, string(hashedPassword), role).Scan(
		&user.ID, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		log.Printf("❌ Ошибка создания пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Генерация токена
	token, expiresAt, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	log.Printf("✅ Зарегистрирован новый пользователь: %s (role: %s)", user.Email, user.Role)

	c.JSON(http.StatusCreated, TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         user,
	})
}

// login вход пользователя
func login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Поиск пользователя
	var user User
	err := db.QueryRow(`
		SELECT id, email, password_hash, role, is_active, created_at, last_login
		FROM users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role,
		&user.IsActive, &user.CreatedAt, &user.LastLogin,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		log.Printf("❌ Ошибка поиска пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Проверка активности
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is inactive"})
		return
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Обновление last_login
	_, err = db.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", user.ID)
	if err != nil {
		log.Printf("⚠️ Не удалось обновить last_login: %v", err)
	}

	// Генерация токена
	token, expiresAt, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	log.Printf("✅ Вход: %s (role: %s)", user.Email, user.Role)

	c.JSON(http.StatusOK, TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         user,
	})
}

// getCurrentUser получить текущего пользователя
func getCurrentUser(c *gin.Context) {
	userID := c.GetInt("user_id")

	var user User
	err := db.QueryRow(`
		SELECT id, email, role, is_active, created_at, last_login
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Role,
		&user.IsActive, &user.CreatedAt, &user.LastLogin,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// RefreshTokenRequest запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// refreshToken обновить токен
func refreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Парсинг refresh токена
	token, err := jwt.ParseWithClaims(req.RefreshToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Проверка что это именно refresh token
	if claims.Issuer != "quotopia-auth-refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not a refresh token"})
		return
	}

	// Проверка что refresh токен существует в БД
	var tokenID int
	var expiresAt time.Time
	err = db.QueryRow(`
		SELECT id, expires_at FROM refresh_tokens
		WHERE token = $1 AND user_id = $2
	`, req.RefreshToken, claims.UserID).Scan(&tokenID, &expiresAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Ошибка проверки refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Проверка срока действия
	if time.Now().After(expiresAt) {
		// Удаляем истекший токен
		db.Exec("DELETE FROM refresh_tokens WHERE id = $1", tokenID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// Получаем данные пользователя
	var user User
	err = db.QueryRow(`
		SELECT id, email, role, is_active FROM users WHERE id = $1
	`, claims.UserID).Scan(&user.ID, &user.Email, &user.Role, &user.IsActive)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is inactive"})
		return
	}

	// Генерация нового access токена
	newToken, newExpiresAt, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("✅ Токен обновлён для пользователя: %s", user.Email)

	c.JSON(http.StatusOK, gin.H{
		"token":      newToken,
		"expires_at": newExpiresAt,
	})
}

// logout выход
func logout(c *gin.Context) {
	// Получаем токен из заголовка
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	userID := c.GetInt("user_id")

	// Парсим токен чтобы получить время истечения
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(*JWTClaims); ok {
			// Добавляем токен в blacklist
			_, err := db.Exec(`
				INSERT INTO token_blacklist (token, user_id, expires_at)
				VALUES ($1, $2, $3)
				ON CONFLICT (token) DO NOTHING
			`, tokenString, userID, claims.ExpiresAt.Time)

			if err != nil {
				log.Printf("⚠️ Не удалось добавить токен в blacklist: %v", err)
			}
		}
	}

	// Удаляем все refresh токены пользователя
	_, err = db.Exec("DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("⚠️ Не удалось удалить refresh токены: %v", err)
	}

	log.Printf("✅ Logout для user_id: %d", userID)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ChangePasswordRequest запрос на смену пароля
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// changePassword смена пароля
func changePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("user_id")

	// Получаем текущий хеш пароля
	var currentHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Проверяем старый пароль
	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPassword))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// Проверяем что новый пароль отличается от старого
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different from old password"})
		return
	}

	// Хешируем новый пароль
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Обновляем пароль
	_, err = db.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", string(newHash), userID)
	if err != nil {
		log.Printf("❌ Ошибка обновления пароля: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Удаляем все refresh токены пользователя (требуем повторного входа)
	db.Exec("DELETE FROM refresh_tokens WHERE user_id = $1", userID)

	log.Printf("✅ Пароль изменён для user_id: %d", userID)
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// listUsers список всех пользователей (только admin)
func listUsers(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, email, role, is_active, created_at, last_login
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt, &user.LastLogin)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// ChangeRoleRequest запрос на изменение роли
type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// changeUserRole изменить роль пользователя (только admin)
func changeUserRole(c *gin.Context) {
	var req ChangeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация роли
	validRoles := map[string]bool{"admin": true, "trader": true, "user": true, "viewer": true}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Allowed: admin, trader, user, viewer"})
		return
	}

	userID := c.Param("id")

	// Проверяем что пользователь существует
	var email string
	var currentRole string
	err := db.QueryRow("SELECT email, role FROM users WHERE id = $1", userID).Scan(&email, &currentRole)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Обновляем роль
	_, err = db.Exec("UPDATE users SET role = $1 WHERE id = $2", req.Role, userID)
	if err != nil {
		log.Printf("❌ Ошибка изменения роли: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	// Удаляем все refresh токены (требуем повторного входа с новой ролью)
	db.Exec("DELETE FROM refresh_tokens WHERE user_id = $1", userID)

	log.Printf("✅ Роль изменена: %s (%s -> %s)", email, currentRole, req.Role)
	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully", "email": email, "new_role": req.Role})
}

// deleteUser удалить пользователя (только admin)
func deleteUser(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetInt("user_id")

	// Проверяем что админ не удаляет сам себя
	if userID == string(rune(adminID)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Проверяем что пользователь существует
	var email string
	err := db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Мягкое удаление - деактивируем пользователя
	_, err = db.Exec("UPDATE users SET is_active = false WHERE id = $1", userID)
	if err != nil {
		log.Printf("❌ Ошибка удаления пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Удаляем все refresh токены пользователя
	db.Exec("DELETE FROM refresh_tokens WHERE user_id = $1", userID)

	log.Printf("✅ Пользователь деактивирован: %s (id: %s)", email, userID)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "email": email})
}

// generateToken генерация JWT токена
func generateToken(user User) (string, time.Time, error) {
	expiresAt := time.Now().Add(1 * time.Hour) // 1 час

	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "quotopia-auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// generateRefreshToken генерация refresh токена
func generateRefreshToken(user User) (string, error) {
	expiresAt := time.Now().Add(24 * 7 * time.Hour) // 7 дней

	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "quotopia-auth-refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	// Сохраняем refresh токен в БД
	_, err = db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, tokenString, expiresAt)

	if err != nil {
		log.Printf("⚠️ Не удалось сохранить refresh token в БД: %v", err)
		// Не возвращаем ошибку, токен все равно валидный
	}

	return tokenString, nil
}

// authMiddleware middleware для проверки JWT
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Убрать "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Проверка blacklist
		var isBlacklisted bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token = $1)", tokenString).Scan(&isBlacklisted)
		if err == nil && isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// Парсинг токена
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// adminMiddleware middleware для проверки роли admin
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// corsMiddleware CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// getEnv получить переменную окружения с дефолтным значением
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
