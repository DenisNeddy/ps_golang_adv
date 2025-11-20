package config

// Объявление пакета config

// Определение структуры Config для хранения конфигурации
type Config struct {
	// Поле Email для хранения email адреса отправителя
	Email string
	// Поле Password для хранения пароля от email аккаунта
	Password string
	// Поле Address для хранения SMTP адреса сервера
	Address string
}
