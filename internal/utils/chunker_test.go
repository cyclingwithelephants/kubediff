package utils

import (
	"testing"
)

func TestNewChunker(t *testing.T) {
	testCases := []struct {
		name           string
		maxChunkChars  int
		expectedOutput int
		expectError    bool
	}{
		{
			name:           "Positive Case: Valid Chunk Size",
			maxChunkChars:  5,
			expectedOutput: 5,
			expectError:    false,
		},
		{
			name:           "Negative Case: Invalid Chunk Size",
			maxChunkChars:  -1,
			expectedOutput: -1,
			expectError:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			chunker := NewChunker(testCase.maxChunkChars)
			if !testCase.expectError && chunker.MaxChunkChars != testCase.expectedOutput {
				t.Errorf("Expected MaxChunkChars to be %v, got %v", testCase.expectedOutput, chunker.MaxChunkChars)
			} else if testCase.expectError && chunker.MaxChunkChars >= 0 {
				t.Errorf("Expected error for MaxChunkChars < 0, but got a chunker with MaxChunkChars: %v", chunker.MaxChunkChars)
			}
		})
	}
}

func TestChunker_Chunk(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		maxChunkSize   int
		expectedOutput []string
	}{
		{
			name:         "Case 1: Single chunk",
			input:        "Hello",
			maxChunkSize: 10,
			expectedOutput: []string{
				"Hello\n",
			},
		},
		{
			name:         "Case 2: Multiple chunks",
			input:        "Hello\nWorld",
			maxChunkSize: 5,
			expectedOutput: []string{
				"Hello\n",
				"World\n",
			},
		},
		{
			name:         "Case 3: Empty string",
			input:        "",
			maxChunkSize: 10,
			expectedOutput: []string{
				"\n",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			chunker := NewChunker(testCase.maxChunkSize)
			result := chunker.Chunk(testCase.input)
			if len(result) != len(testCase.expectedOutput) {
				t.Errorf("Expected %v chunks, got %v", len(testCase.expectedOutput), len(result))
				return
			}
			for i, chunk := range result {
				if chunk != testCase.expectedOutput[i] {
					t.Errorf("Expected chunk %v to be %v, got %v", i, testCase.expectedOutput[i], chunk)
				}
			}
		})
	}
}
