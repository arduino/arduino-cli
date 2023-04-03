package builder_test

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/stretchr/testify/require"
)

func TestParseCppString(t *testing.T) {
	_, _, ok := builder.ParseCppString(`foo`)
	require.Equal(t, false, ok)

	_, _, ok = builder.ParseCppString(`"foo`)
	require.Equal(t, false, ok)

	str, rest, ok := builder.ParseCppString(`"foo"`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = builder.ParseCppString(`"foo\\bar"`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo\bar`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = builder.ParseCppString(`"foo \"is\" quoted and \\\\bar\"\" escaped\\" and "then" some`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo "is" quoted and \\bar"" escaped\`, str)
	require.Equal(t, ` and "then" some`, rest)

	str, rest, ok = builder.ParseCppString(`" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~"`)
	require.Equal(t, true, ok)
	require.Equal(t, ` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_abcdefghijklmnopqrstuvwxyz{|}~`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = builder.ParseCppString(`"/home/ççç/"`)
	require.Equal(t, true, ok)
	require.Equal(t, `/home/ççç/`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = builder.ParseCppString(`"/home/ççç/ /$sdsdd\\"`)
	require.Equal(t, true, ok)
	require.Equal(t, `/home/ççç/ /$sdsdd\`, str)
	require.Equal(t, ``, rest)
}
