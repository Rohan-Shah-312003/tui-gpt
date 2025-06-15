// internal/storage/utils.go
package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

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
