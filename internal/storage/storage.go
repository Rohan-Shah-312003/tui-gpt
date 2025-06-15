// internal/storage/storage.go
package storage

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	storageDir     = "chat_history"
	maxChatName    = 50
	dateTimeFormat = "2006-01-02_15-04-05"
)

type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatSession struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Messages  []ChatMessage `json:"messages"`
}

type Storage struct {
	baseDir string
}

func NewStorage() *Storage {
	return &Storage{
		baseDir: storageDir,
	}
}

func (s *Storage) Initialize() error {
	return os.MkdirAll(s.baseDir, 0755)
}

func (s *Storage) SaveChat(session *ChatSession) error {
	if session.ID == "" {
		session.ID = generateChatID()
	}

	session.UpdatedAt = time.Now()

	// Generate title from first user message if not set
	if session.Title == "" && len(session.Messages) > 0 {
		for _, msg := range session.Messages {
			if msg.Role == "user" {
				session.Title = generateTitle(msg.Content)
				break
			}
		}
		if session.Title == "" {
			session.Title = "New Chat"
		}
	}

	filename := fmt.Sprintf("%s.json", session.ID)
	filepath := filepath.Join(s.baseDir, filename)

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chat session: %v", err)
	}

	return os.WriteFile(filepath, data, 0644)
}

func (s *Storage) LoadChat(chatID string) (*ChatSession, error) {
	filename := fmt.Sprintf("%s.json", chatID)
	filepath := filepath.Join(s.baseDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chat file: %v", err)
	}

	var session ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chat session: %v", err)
	}

	return &session, nil
}

func (s *Storage) ListChats() ([]ChatSession, error) {
	var sessions []ChatSession

	err := filepath.WalkDir(s.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		chatID := strings.TrimSuffix(d.Name(), ".json")
		session, loadErr := s.LoadChat(chatID)
		if loadErr != nil {
			// Skip corrupted files
			return nil
		}

		sessions = append(sessions, *session)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list chat files: %v", err)
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

func (s *Storage) DeleteChat(chatID string) error {
	filename := fmt.Sprintf("%s.json", chatID)
	filepath := filepath.Join(s.baseDir, filename)
	return os.Remove(filepath)
}

func (s *Storage) GetChatSummaries() ([]ChatSummary, error) {
	sessions, err := s.ListChats()
	if err != nil {
		return nil, err
	}

	var summaries []ChatSummary
	for _, session := range sessions {
		summary := ChatSummary{
			ID:           session.ID,
			Title:        session.Title,
			CreatedAt:    session.CreatedAt,
			UpdatedAt:    session.UpdatedAt,
			MessageCount: len(session.Messages),
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

type ChatSummary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
}

func generateChatID() string {
	return fmt.Sprintf("chat_%s", time.Now().Format(dateTimeFormat))
}

func generateTitle(content string) string {
	// Clean and truncate content for title
	title := strings.TrimSpace(content)
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(title, "  ") {
		title = strings.ReplaceAll(title, "  ", " ")
	}

	if len(title) > maxChatName {
		title = title[:maxChatName] + "..."
	}

	if title == "" {
		title = "New Chat"
	}

	return title
}
