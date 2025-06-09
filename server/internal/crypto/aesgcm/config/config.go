package config

type Config struct {
	// Мастер-ключ
	MasterSK string
	// Интревал создания нового промежуточного ключа
	NewSKIntervalD int
}
