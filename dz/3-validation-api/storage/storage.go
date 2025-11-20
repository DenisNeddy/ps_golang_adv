package storage

import (
	"encoding/json"
	"os"
	"sync"
)

// VerificationData структура для хранения данных верификации

type VerificationData struct {
	Email string `json:"email"`
	Hash  string `json:"hash"`
}

// Storage управляет хранением данных верификации в JSON файле

type Storage struct {
	mu       sync.RWMutex
	data     map[string]VerificationData
	filename string
}

// NewStorage создает новый Storage и загружает данные из файла

func NewStorage(filename string) *Storage {
	storage := &Storage{
		data:     make(map[string]VerificationData),
		filename: filename,
	}
	storage.loadFromFile()
	return storage
}

// Save сохраняет email и hash в хранилище и JSON

func (s *Storage) Save(email, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[hash] = VerificationData{
		Email: email,
		Hash:  hash,
	}

	return s.saveToFile()
}

// Get возвращает  данные верификации по hash

func (s *Storage) Get(hash string) (VerificationData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.data[hash]
	return data, exists
}

// Delete удаляет запись по hash и сохраняет изменения в файл

func (s *Storage) Delete(hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, hash)
	return s.saveToFile()
}

// saveToFile сохраняет все данные в JSON файл

func (s *Storage) saveToFile() error {
	file, err := os.Create(s.filename)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s.data)
}

//loadFromfile Загружаем данные из JSON файл

func (s *Storage) loadFromFile() {
	file, err := os.Open(s.filename)
	if err != nil {
		// Если файл не существует, это номально при первом запуске
		return
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Decode(&s.data)
}
