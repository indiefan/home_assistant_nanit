package utils

import (
	"bufio"
	"io"
	"strings"
)

// LogTailer - implements circular buffer for tailing last n-lines
type LogTailer struct {
	lines       []string
	numLines    int
	maxNumLines int
	nextIdx     int
}

// NewLogTailer - constructor
func NewLogTailer(maxNumLines int) *LogTailer {
	return &LogTailer{
		lines:       make([]string, maxNumLines),
		nextIdx:     0,
		numLines:    0,
		maxNumLines: maxNumLines,
	}
}

// GetLines - return tailed lines
func (tailer *LogTailer) GetLines() []string {
	result := make([]string, tailer.numLines)
	for i := tailer.numLines - 1; i >= 0; i-- {
		result[tailer.numLines-i-1] = tailer.lines[(tailer.nextIdx-1-i+tailer.maxNumLines)%tailer.maxNumLines]
	}
	return result
}

// String - returns tailed lines as string
func (tailer *LogTailer) String() string {
	return strings.Join(tailer.GetLines(), "\n")
}

// Append - appends line to the buffer
func (tailer *LogTailer) Append(line string) {
	tailer.lines[tailer.nextIdx] = line
	tailer.nextIdx = (tailer.nextIdx + 1) % tailer.maxNumLines
	if tailer.numLines < tailer.maxNumLines {
		tailer.numLines++
	}
}

// Tail - tails line of given reader
func (tailer *LogTailer) Tail(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		tailer.Append(scanner.Text())
	}
}
