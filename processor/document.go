package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unidoc/unipdf/v4/extractor"
	"github.com/unidoc/unipdf/v4/model"
)

type DocumentProcessor struct {
	chunkSize    int
	chunkOverlap int
}

func NewDocumentProcessor(chunkSize, chunkOverlap int) *DocumentProcessor {
	return &DocumentProcessor{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

func (dp *DocumentProcessor) ProcessFile(ctx context.Context, filePath string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return dp.processPDF(ctx, filePath)
	case ".txt":
		return dp.processText(ctx, filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func (dp *DocumentProcessor) processPDF(ctx context.Context, filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	reader, err := model.NewPdfReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var chunks []string
	numPages, err := reader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get number of pages: %w", err)
	}

	for i := 0; i < numPages; i++ {
		page, err := reader.GetPage(i + 1)
		if err != nil {
			return nil, fmt.Errorf("failed to get page %d: %w", i+1, err)
		}

		ext, err := extractor.New(page)
		if err != nil {
			return nil, fmt.Errorf("failed to create extractor for page %d: %w", i+1, err)
		}

		text, err := ext.ExtractText()
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}

		pageChunks := dp.chunkText(text)
		chunks = append(chunks, pageChunks...)
	}

	return chunks, nil
}

func (dp *DocumentProcessor) processText(ctx context.Context, filepath string) ([]string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	return dp.chunkText(string(content)), nil
}

func (dp *DocumentProcessor) chunkText(text string) []string {
	words := strings.Fields(text)
	var chunks []string
	var currentChunk strings.Builder

	for i, word := range words {
		currentChunk.WriteString(word)
		currentChunk.WriteString(" ")

		if (i+1)%dp.chunkSize == 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()

			// Add overlap
			if i+dp.chunkOverlap < len(words) {
				for j := i - dp.chunkOverlap + 1; j <= i; j++ {
					if j >= 0 {
						currentChunk.WriteString(words[j])
						currentChunk.WriteString(" ")
					}
				}
			}
		}
	}

	// Add remaining text
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}
