package main

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestPasswordHashing проверяет хеширование паролей
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// Хешируем
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Проверяем корректный пароль
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		t.Error("Valid password should match")
	}

	// Проверяем неправильный пароль
	err = bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword"))
	if err == nil {
		t.Error("Invalid password should not match")
	}
}

// TestGetEnv проверяет функцию получения env переменных
func TestGetEnv(t *testing.T) {
	// Без установки переменной - должен вернуть default
	value := getEnv("NONEXISTENT_VAR", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}

	// Установим переменную
	t.Setenv("TEST_VAR", "test_value")
	value = getEnv("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
}

// TestJWTSecret проверяет что JWT secret не пустой
func TestJWTSecret(t *testing.T) {
	if len(jwtSecret) == 0 {
		t.Error("JWT secret should not be empty")
	}
}
