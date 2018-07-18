package librariesresolver

import (
	"testing"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/stretchr/testify/require"
)

func TestCppHeaderPriority(t *testing.T) {
	l1 := &libraries.Library{Name: "Calculus Lib", Location: libraries.Sketchbook}
	l2 := &libraries.Library{Name: "Calculus Lib-master", Location: libraries.Sketchbook}
	l3 := &libraries.Library{Name: "Calculus Lib Improved", Location: libraries.Sketchbook}
	l4 := &libraries.Library{Name: "Another Calculus Lib", Location: libraries.Sketchbook}
	l5 := &libraries.Library{Name: "Yet Another Calculus Lib Improved", Location: libraries.Sketchbook}
	l6 := &libraries.Library{Name: "AnotherLib", Location: libraries.Sketchbook}

	r1 := computePriority(l1, "calculus_lib.h", "avr")
	r2 := computePriority(l2, "calculus_lib.h", "avr")
	r3 := computePriority(l3, "calculus_lib.h", "avr")
	r4 := computePriority(l4, "calculus_lib.h", "avr")
	r5 := computePriority(l5, "calculus_lib.h", "avr")
	r6 := computePriority(l6, "calculus_lib.h", "avr")
	require.True(t, r1 > r2)
	require.True(t, r2 > r3)
	require.True(t, r3 > r4)
	require.True(t, r4 > r5)
	require.True(t, r5 > r6)
}
