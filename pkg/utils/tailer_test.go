package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
)

func TestTailer(t *testing.T) {
	tailer := utils.NewLogTailer(3)

	assert.Empty(t, tailer.GetLines())

	tailer.Append("a")
	assert.Equal(t, []string{"a"}, tailer.GetLines())

	tailer.Append("b")
	assert.Equal(t, []string{"a", "b"}, tailer.GetLines())

	tailer.Append("c")
	assert.Equal(t, []string{"a", "b", "c"}, tailer.GetLines())

	tailer.Append("d")
	assert.Equal(t, []string{"b", "c", "d"}, tailer.GetLines())

	tailer.Append("e")
	assert.Equal(t, []string{"c", "d", "e"}, tailer.GetLines())

	tailer.Append("f")
	assert.Equal(t, []string{"d", "e", "f"}, tailer.GetLines())

	tailer.Append("g")
	assert.Equal(t, []string{"e", "f", "g"}, tailer.GetLines())
}
