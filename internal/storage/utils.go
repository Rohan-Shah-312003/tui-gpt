package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StorageStats represents storage usage statistics
type StorageStats struct {
	TotalChats    int   `json:"total_chats"`
	TotalMessages int   `json:"total_messages"`
	TotalSize     int64 `json:"total_size"`
}

// CleanupStorage removes old chat files if storage gets too large
func (s *Storage) CleanupStorage(maxFiles int) error {
	summaries, err := s.GetChatSummaries()
	if err != nil {
		return err
	}

	if len(summaries) <= maxFiles {
		return nil // No cleanup needed
	}

	// Remove oldest files (summaries are sorted by UpdatedAt desc)
	filesToRemove := summaries[maxFiles:]

	for _, summary := range filesToRemove {
		if err := s.DeleteChat(summary.ID); err != nil {
			return fmt.Errorf("failed to delete old chat %s: %v", summary.ID, err)
		}
	}

	return nil
}

// GetStorageStats returns information about storage usage
func (s *Storage) GetStorageStats() (StorageStats, error) {
	stats := StorageStats{}

	summaries, err := s.GetChatSummaries()
	if err != nil {
		return stats, err
	}

	stats.TotalChats = len(summaries)

	// Calculate total storage size
	err = filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			stats.TotalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return stats, err
	}

	// Calculate total messages
	for _, summary := range summaries {
		stats.TotalMessages += summary.MessageCount
	}

	return stats, nil
}

// ExportChat exports a chat session to a readable text format
func (s *Storage) ExportChat(chatID string, outputPath string) error {
	session, err := s.LoadChat(chatID)
	if err != nil {
		return fmt.Errorf("failed to load chat: %v", err)
	}

	var content strings.Builder

	// Header
	content.WriteString("=====================================\n")
	content.WriteString(fmt.Sprintf("Chat Export: %s\n", session.Title))
	content.WriteString(fmt.Sprintf("Created: %s\n", session.CreatedAt.Format("January 2, 2006 at 15:04:05")))
	content.WriteString(fmt.Sprintf("Updated: %s\n", session.UpdatedAt.Format("January 2, 2006 at 15:04:05")))
	content.WriteString(fmt.Sprintf("Total Messages: %d\n", len(session.Messages)))
	content.WriteString("=====================================\n\n")

	// Messages
	for i, msg := range session.Messages {
		content.WriteString(fmt.Sprintf("[%s] %s:\n",
			msg.Timestamp.Format("15:04:05"),
			strings.Title(msg.Role)))
		content.WriteString(fmt.Sprintf("%s\n", msg.Content))

		// Add separator between messages
		if i < len(session.Messages)-1 {
			content.WriteString("\n" + strings.Repeat("-", 50) + "\n\n")
		}
	}

	// Footer
	content.WriteString("\n\n=====================================\n")
	content.WriteString("Exported from TUI-GPT\n")
	content.WriteString(fmt.Sprintf("Export Time: %s\n", time.Now().Format("January 2, 2006 at 15:04:05")))
	content.WriteString("=====================================\n")

	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

// ImportChatFromText imports a chat from a text file (simple format)
func (s *Storage) ImportChatFromText(filePath string, title string) (*ChatSession, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	session := &ChatSession{
		ID:        generateChatID(),
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []ChatMessage{},
	}

	var currentMessage strings.Builder
	var currentRole string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check if line starts with a role indicator
		if strings.HasPrefix(line, "User:") || strings.HasPrefix(line, "YOU:") {
			// Save previous message if exists
			if currentRole != "" && currentMessage.Len() > 0 {
				session.Messages = append(session.Messages, ChatMessage{
					Role:      currentRole,
					Content:   strings.TrimSpace(currentMessage.String()),
					Timestamp: time.Now(),
				})
			}
			currentRole = "user"
			currentMessage.Reset()
			currentMessage.WriteString(strings.TrimSpace(line[5:]))
		} else if strings.HasPrefix(line, "AI:") || strings.HasPrefix(line, "Assistant:") {
			// Save previous message if exists
			if currentRole != "" && currentMessage.Len() > 0 {
				session.Messages = append(session.Messages, ChatMessage{
					Role:      currentRole,
					Content:   strings.TrimSpace(currentMessage.String()),
					Timestamp: time.Now(),
				})
			}
			currentRole = "assistant"
			currentMessage.Reset()
			prefixLen := 3
			if strings.HasPrefix(line, "Assistant:") {
				prefixLen = 10
			}
			currentMessage.WriteString(strings.TrimSpace(line[prefixLen:]))
		} else if line != "" && currentRole != "" {
			// Continue current message
			if currentMessage.Len() > 0 {
				currentMessage.WriteString("\n")
			}
			currentMessage.WriteString(line)
		}
	}

	// Save the last message
	if currentRole != "" && currentMessage.Len() > 0 {
		session.Messages = append(session.Messages, ChatMessage{
			Role:      currentRole,
			Content:   strings.TrimSpace(currentMessage.String()),
			Timestamp: time.Now(),
		})
	}

	// Save the imported session
	if err := s.SaveChat(session); err != nil {
		return nil, fmt.Errorf("failed to save imported chat: %v", err)
	}

	return session, nil
}

// BackupChats creates a backup of all chats in a specified directory
func (s *Storage) BackupChats(backupDir string) error {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	summaries, err := s.GetChatSummaries()
	if err != nil {
		return fmt.Errorf("failed to get chat summaries: %v", err)
	}

	for _, summary := range summaries {
		// Create safe filename from chat title
		safeTitle := strings.ReplaceAll(summary.Title, "/", "_")
		safeTitle = strings.ReplaceAll(safeTitle, "\\", "_")
		safeTitle = strings.ReplaceAll(safeTitle, ":", "_")

		filename := fmt.Sprintf("%s_%s.txt", summary.ID, safeTitle)
		outputPath := filepath.Join(backupDir, filename)

		if err := s.ExportChat(summary.ID, outputPath); err != nil {
			return fmt.Errorf("failed to backup chat %s: %v", summary.ID, err)
		}
	}

	return nil
}

// ValidateStorage checks the integrity of stored chat files
func (s *Storage) ValidateStorage() ([]string, error) {
	var issues []string

	// Check if storage directory exists
	if _, err := os.Stat(s.baseDir); os.IsNotExist(err) {
		issues = append(issues, "Storage directory does not exist")
		return issues, nil
	}

	// Get all JSON files in storage directory
	files, err := filepath.Glob(filepath.Join(s.baseDir, "*.json"))
	if err != nil {
		return issues, fmt.Errorf("failed to list chat files: %v", err)
	}

	for _, file := range files {
		filename := filepath.Base(file)
		chatID := strings.TrimSuffix(filename, ".json")

		// Try to load each chat
		_, err := s.LoadChat(chatID)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Corrupted chat file: %s (%v)", filename, err))
		}
	}

	return issues, nil
}

// GetChatsByDateRange returns chats within a specific date range
func (s *Storage) GetChatsByDateRange(startDate, endDate time.Time) ([]ChatSession, error) {
	allSessions, err := s.ListChats()
	if err != nil {
		return nil, err
	}

	var filteredSessions []ChatSession
	for _, session := range allSessions {
		if (session.UpdatedAt.After(startDate) || session.UpdatedAt.Equal(startDate)) &&
			(session.UpdatedAt.Before(endDate) || session.UpdatedAt.Equal(endDate)) {
			filteredSessions = append(filteredSessions, session)
		}
	}

	return filteredSessions, nil
}

// SearchChats searches for chats containing specific text in messages
func (s *Storage) SearchChats(query string) ([]ChatSession, error) {
	allSessions, err := s.ListChats()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matchingSessions []ChatSession

	for _, session := range allSessions {
		// Check title
		if strings.Contains(strings.ToLower(session.Title), query) {
			matchingSessions = append(matchingSessions, session)
			continue
		}

		// Check messages
		for _, message := range session.Messages {
			if strings.Contains(strings.ToLower(message.Content), query) {
				matchingSessions = append(matchingSessions, session)
				break
			}
		}
	}

	return matchingSessions, nil
}

// GetStorageSize returns the total size of the storage directory in bytes
func (s *Storage) GetStorageSize() (int64, error) {
	var size int64

	err := filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// FormatStorageSize returns a human-readable storage size string
func FormatStorageSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
