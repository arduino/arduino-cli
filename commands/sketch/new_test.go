package sketch

import (
	"context"
	"fmt"
	"testing"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

func Test_SketchNameWrongPattern(t *testing.T) {
	invalidNames := []string{
		"&",
		".hello",
		"-hello",
		"hello*",
		"||||||||||||||",
		",`hack[}attempt{];",
	}
	for _, name := range invalidNames {
		_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
			SketchName: name,
			SketchDir:  t.TempDir(),
		})

		require.EqualError(t, err, fmt.Sprintf(`Can't create sketch: invalid sketch name "%s": the first character must be alphanumeric or "_", the following ones can also contain "-" and ".".`,
			name))
	}
}

func Test_SketchNameEmpty(t *testing.T) {
	emptyName := ""
	_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
		SketchName: emptyName,
		SketchDir:  t.TempDir(),
	})

	require.EqualError(t, err, `Can't create sketch: sketch name cannot be empty`)
}

func Test_SketchNameTooLong(t *testing.T) {
	tooLongName := make([]byte, sketchNameMaxLength+1)
	for i := range tooLongName {
		tooLongName[i] = 'a'
	}
	_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
		SketchName: string(tooLongName),
		SketchDir:  t.TempDir(),
	})

	require.EqualError(t, err, fmt.Sprintf(`Can't create sketch: sketch name too long (%d characters). Maximum allowed length is %d`,
		len(tooLongName),
		sketchNameMaxLength))
}

func Test_SketchNameOk(t *testing.T) {
	lengthLimitName := make([]byte, sketchNameMaxLength)
	for i := range lengthLimitName {
		lengthLimitName[i] = 'a'
	}
	validNames := []string{
		"h",
		"h.ello",
		"h..ello-world",
		"h..ello-world.",
		"hello_world__",
		"_hello_world",
		string(lengthLimitName),
	}
	for _, name := range validNames {
		_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
			SketchName: name,
			SketchDir:  t.TempDir(),
		})
		require.Nil(t, err)
	}
}
