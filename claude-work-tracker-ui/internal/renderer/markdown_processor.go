package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MarkdownProcessor handles pre-processing of markdown content for Glamour rendering
type MarkdownProcessor struct {
	baseDir string
}

// NewMarkdownProcessor creates a new markdown processor
func NewMarkdownProcessor(baseDir string) *MarkdownProcessor {
	return &MarkdownProcessor{
		baseDir: baseDir,
	}
}

// ProcessForRendering prepares markdown content for Glamour rendering with full embedding resolution
func (mp *MarkdownProcessor) ProcessForRendering(content string) string {
	// Step 1: Convert extended todo syntax to standard + emoji
	content = mp.convertExtendedTodos(content)
	
	// Step 2: Resolve artifact embeddings
	content = mp.resolveEmbeddings(content)
	
	return content
}

// ProcessForLightRendering prepares markdown content for fast rendering without embedding resolution
func (mp *MarkdownProcessor) ProcessForLightRendering(content string) string {
	// Step 1: Convert extended todo syntax to standard + emoji
	content = mp.convertExtendedTodos(content)
	
	// Step 2: Convert embeddings to clickable placeholders instead of resolving them
	content = mp.convertEmbeddingsToPlaceholders(content)
	
	return content
}

// ProcessWithAsyncEmbeddings prepares content with loaded embeddings and loading states
func (mp *MarkdownProcessor) ProcessWithAsyncEmbeddings(content string, loadedEmbeddings map[string]string, loadingStates map[string]string) string {
	// Step 1: Convert extended todo syntax to standard + emoji
	content = mp.convertExtendedTodos(content)
	
	// Step 2: Replace embeddings with loaded content, spinners, or placeholders
	content = mp.replaceEmbeddingsWithSpinners(content, loadedEmbeddings, loadingStates)
	
	return content
}

// replaceEmbeddingsWithSpinners replaces embeddings with loaded content, spinners, or placeholders
func (mp *MarkdownProcessor) replaceEmbeddingsWithSpinners(content string, loadedEmbeddings map[string]string, loadingStates map[string]string) string {
	// Pattern for embedding syntax: ![[filename.md]] or ![[artifact-id]]
	embeddingPattern := regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
	
	return embeddingPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the reference
		submatches := embeddingPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match // Return original if can't parse
		}
		
		reference := submatches[1]
		filename := filepath.Base(reference)
		
		// Check if we have loaded content for this embedding
		if loadedContent, exists := loadedEmbeddings[reference]; exists {
			if loadedContent == "" {
				// Content loaded but empty - show error message
				return fmt.Sprintf("\n\n### 📄 %s\n\n⚠️ **Content not found or empty**\n", filename)
			}
			// Process the loaded content first (remove frontmatter, etc.)
			processedContent := mp.extractContentFromMarkdown(loadedContent)
			if processedContent == "" {
				// Processed content is empty - show original content
				processedContent = loadedContent
			}
			// Return the loaded content in a styled container
			return fmt.Sprintf("\n\n---\n\n### 📄 %s\n\n%s\n\n---\n", filename, processedContent)
		}
		
		// Check if this embedding is currently loading
		if spinnerView, exists := loadingStates[reference]; exists {
			return fmt.Sprintf("📎 **[%s]** %s **loading...**", filename, spinnerView)
		}
		
		// Return default placeholder
		return fmt.Sprintf("📎 **[%s]** _(auto-loading...)_", filename)
	})
}

// convertExtendedTodos converts our extended checkbox syntax to Glamour-compatible format
func (mp *MarkdownProcessor) convertExtendedTodos(content string) string {
	// Define patterns for extended todo syntax
	patterns := map[string]string{
		`\[…\]`:  `[🔄]`,  // In progress -> spinning arrow
		`\[!\]`:  `[⚠️]`,  // Blocked -> warning
		`\[-\]`:  `[❌]`,  // Cancelled -> X
		`\[x\]`:  `[✅]`,  // Completed -> checkmark
		`\[ \]`:  `[⭕]`,  // Todo -> circle
	}
	
	for pattern, replacement := range patterns {
		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, replacement)
	}
	
	return content
}

// convertEmbeddingsToPlaceholders converts ![[artifact-id]] to clickable placeholders for lightweight rendering
func (mp *MarkdownProcessor) convertEmbeddingsToPlaceholders(content string) string {
	// Pattern for embedding syntax: ![[filename.md]] or ![[artifact-id]]
	embeddingPattern := regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
	
	return embeddingPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the reference
		submatches := embeddingPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match // Return original if can't parse
		}
		
		reference := submatches[1]
		filename := filepath.Base(reference)
		
		// Create a clickable placeholder link
		return fmt.Sprintf("📎 **[%s]** _(click to view)_", filename)
	})
}


// ExtractEmbeddingReferences extracts all embedding references from content
func (mp *MarkdownProcessor) ExtractEmbeddingReferences(content string) []string {
	embeddingPattern := regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
	matches := embeddingPattern.FindAllStringSubmatch(content, -1)
	
	var references []string
	for _, match := range matches {
		if len(match) >= 2 {
			references = append(references, match[1])
		}
	}
	
	return references
}

// resolveEmbeddings resolves ![[artifact-id]] references to actual content
func (mp *MarkdownProcessor) resolveEmbeddings(content string) string {
	// Pattern for embedding syntax: ![[filename.md]] or ![[artifact-id]]
	embeddingPattern := regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
	
	return embeddingPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the reference
		submatches := embeddingPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match // Return original if can't parse
		}
		
		reference := submatches[1]
		
		// Try to resolve the reference
		resolvedContent := mp.resolveReference(reference)
		if resolvedContent == "" {
			// Fallback to a link if can't resolve
			return fmt.Sprintf("🔗 [%s](artifact:%s)", reference, reference)
		}
		
		// Return resolved content with border and indentation
		filename := filepath.Base(reference)
		
		// Use horizontal rules for visual separation
		return fmt.Sprintf("\n\n---\n\n### 📄 %s\n\n%s\n\n---\n", filename, resolvedContent)
	})
}

// ResolveReference attempts to resolve an artifact reference to content (public method)
func (mp *MarkdownProcessor) ResolveReference(reference string) string {
	return mp.resolveReference(reference)
}

// resolveReference attempts to resolve an artifact reference to content
func (mp *MarkdownProcessor) resolveReference(reference string) string {
	// Try different resolution strategies
	
	// Strategy 1: Direct file path (e.g., updates/work-123.md)
	if strings.Contains(reference, "/") {
		fullPath := filepath.Join(mp.baseDir, reference)
		if content, err := os.ReadFile(fullPath); err == nil {
			return mp.extractContentFromMarkdown(string(content))
		}
	}
	
	// Strategy 2: Look in artifacts directory by ID
	if !strings.HasSuffix(reference, ".md") {
		// Try to find artifact by ID
		artifactContent := mp.findArtifactByID(reference)
		if artifactContent != "" {
			return artifactContent
		}
	}
	
	// Strategy 3: Look for exact filename in various directories
	if strings.HasSuffix(reference, ".md") {
		dirs := []string{
			"artifacts/plans",
			"artifacts/proposals", 
			"artifacts/analyses",
			"artifacts/updates",
			"artifacts/decisions",
			"updates",
		}
		
		for _, dir := range dirs {
			fullPath := filepath.Join(mp.baseDir, dir, reference)
			if content, err := os.ReadFile(fullPath); err == nil {
				return mp.extractContentFromMarkdown(string(content))
			}
		}
	}
	
	return "" // Could not resolve
}

// findArtifactByID searches for an artifact by ID across all artifact directories
func (mp *MarkdownProcessor) findArtifactByID(artifactID string) string {
	dirs := []string{
		"artifacts/plans",
		"artifacts/proposals", 
		"artifacts/analyses",
		"artifacts/updates",
		"artifacts/decisions",
	}
	
	for _, dir := range dirs {
		dirPath := filepath.Join(mp.baseDir, dir)
		if files, err := os.ReadDir(dirPath); err == nil {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
					filePath := filepath.Join(dirPath, file.Name())
					if content, err := os.ReadFile(filePath); err == nil {
						// Check if this file contains the artifact ID
						if strings.Contains(string(content), fmt.Sprintf("id: %s", artifactID)) {
							return mp.extractContentFromMarkdown(string(content))
						}
					}
				}
			}
		}
	}
	
	return ""
}

// extractContentFromMarkdown extracts the main content from a markdown file, skipping frontmatter
func (mp *MarkdownProcessor) extractContentFromMarkdown(content string) string {
	// Split by frontmatter delimiter
	parts := strings.Split(content, "---")
	
	// If we have frontmatter (starts with ---), content is after the second ---
	if len(parts) >= 3 && strings.TrimSpace(parts[0]) == "" {
		// Join everything after frontmatter
		mainContent := strings.Join(parts[2:], "---")
		return strings.TrimSpace(mainContent)
	}
	
	// No frontmatter, return as-is
	return strings.TrimSpace(content)
}

// GetTaskSummary returns a brief summary of tasks for quick display
func (mp *MarkdownProcessor) GetTaskSummary(content string) string {
	lines := strings.Split(content, "\n")
	var summary strings.Builder
	
	todoCount := 0
	inProgressCount := 0
	completedCount := 0
	blockedCount := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [") {
			if strings.Contains(line, "[ ]") || strings.Contains(line, "[⭕]") {
				todoCount++
			} else if strings.Contains(line, "[x]") || strings.Contains(line, "[✅]") {
				completedCount++
			} else if strings.Contains(line, "[…]") || strings.Contains(line, "[🔄]") {
				inProgressCount++
			} else if strings.Contains(line, "[!]") || strings.Contains(line, "[⚠️]") {
				blockedCount++
			}
		}
	}
	
	total := todoCount + inProgressCount + completedCount + blockedCount
	if total == 0 {
		return ""
	}
	
	summary.WriteString(fmt.Sprintf("**Tasks**: %d total", total))
	if completedCount > 0 {
		summary.WriteString(fmt.Sprintf(" • ✅ %d completed", completedCount))
	}
	if inProgressCount > 0 {
		summary.WriteString(fmt.Sprintf(" • 🔄 %d in progress", inProgressCount))
	}
	if blockedCount > 0 {
		summary.WriteString(fmt.Sprintf(" • ⚠️ %d blocked", blockedCount))
	}
	if todoCount > 0 {
		summary.WriteString(fmt.Sprintf(" • ⭕ %d todo", todoCount))
	}
	
	return summary.String()
}

// ProcessForDisplay creates a lightweight version for list display
func (mp *MarkdownProcessor) ProcessForDisplay(content string, maxLength int) string {
	// Get task summary
	taskSummary := mp.GetTaskSummary(content)
	
	// Extract first paragraph of actual content (skip headers and tasks)
	lines := strings.Split(content, "\n")
	var contentLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, headers, tasks, and frontmatter
		if line == "" || strings.HasPrefix(line, "#") || 
		   strings.HasPrefix(line, "- [") || strings.HasPrefix(line, "---") ||
		   strings.HasPrefix(line, "*Last updated:") {
			continue
		}
		
		// Skip embeddings in display mode
		if strings.Contains(line, "![[") {
			continue
		}
		
		contentLines = append(contentLines, line)
		
		// Stop after first substantial paragraph
		if len(strings.Join(contentLines, " ")) > maxLength/2 {
			break
		}
	}
	
	result := strings.Join(contentLines, " ")
	
	// Truncate if too long
	if len(result) > maxLength {
		result = result[:maxLength-3] + "..."
	}
	
	// Add task summary if we have one
	if taskSummary != "" {
		if result != "" {
			result = result + "\n\n" + taskSummary
		} else {
			result = taskSummary
		}
	}
	
	return result
}

