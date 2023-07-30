package utils

import (
	"strings"
	"unicode/utf8"
)

type Chunker struct {
	MaxChunkChars int
}

func NewChunker(maxChunkChars int) Chunker {
	return Chunker{MaxChunkChars: maxChunkChars}
}

func (C Chunker) Chunk(
	toSplit string, // the string to split
	// RegexpDelim string, // the regexp delimiter to split on, accepts RE2 syntax
) []string {
	asList := strings.Split(toSplit, "\n")

	var chunks []string
	totalCharsForChunk := 0
	newChunk := ""
	for _, line := range asList {
		chars := utf8.RuneCountInString(line)
		// If adding this line would exceed the max chunk size,
		// flush the chunk and start a new one
		if totalCharsForChunk+chars > C.MaxChunkChars {
			totalCharsForChunk = 0
			chunks = append(chunks, newChunk)
			newChunk = ""
		}
		newChunk = newChunk + line + "\n"
		totalCharsForChunk += chars
	}

	// Flush the last chunk
	return append(chunks, newChunk)
}
