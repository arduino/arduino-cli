package paths

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	testPath := New("_testdata")
	require.Equal(t, "_testdata", testPath.String())
	isDir, err := testPath.IsDir()
	require.True(t, isDir)
	require.NoError(t, err)
	exist, err := testPath.Exist()
	require.True(t, exist)
	require.NoError(t, err)

	folderPath := testPath.Join("folder")
	require.Equal(t, "_testdata/folder", folderPath.String())
	isDir, err = folderPath.IsDir()
	require.True(t, isDir)
	require.NoError(t, err)
	exist, err = folderPath.Exist()
	require.True(t, exist)
	require.NoError(t, err)

	filePath := testPath.Join("file")
	require.Equal(t, "_testdata/file", filePath.String())
	isDir, err = filePath.IsDir()
	require.False(t, isDir)
	require.NoError(t, err)
	exist, err = filePath.Exist()
	require.True(t, exist)
	require.NoError(t, err)

	anotherFilePath := filePath.Join("notexistent")
	require.Equal(t, "_testdata/file/notexistent", anotherFilePath.String())
	isDir, err = anotherFilePath.IsDir()
	require.False(t, isDir)
	require.Error(t, err)
	exist, err = anotherFilePath.Exist()
	require.False(t, exist)
	require.Error(t, err)

	list, err := folderPath.ReadDir()
	require.NoError(t, err)
	require.Len(t, list, 4)
	require.Equal(t, "_testdata/folder/.hidden", list[0].String())
	require.Equal(t, "_testdata/folder/file2", list[1].String())
	require.Equal(t, "_testdata/folder/file3", list[2].String())
	require.Equal(t, "_testdata/folder/subfolder", list[3].String())

	list2 := list.Clone()
	list2.FilterDirs()
	require.Len(t, list2, 1)
	require.Equal(t, "_testdata/folder/subfolder", list2[0].String())

	list2 = list.Clone()
	list2.FilterOutHiddenFiles()
	require.Len(t, list2, 3)
	require.Equal(t, "_testdata/folder/file2", list2[0].String())
	require.Equal(t, "_testdata/folder/file3", list2[1].String())
	require.Equal(t, "_testdata/folder/subfolder", list2[2].String())

	list2 = list.Clone()
	list2.FilterOutPrefix("file")
	require.Len(t, list2, 2)
	require.Equal(t, "_testdata/folder/.hidden", list2[0].String())
	require.Equal(t, "_testdata/folder/subfolder", list2[1].String())
}

func TestResetStatCacheWhenFollowingSymlink(t *testing.T) {
	testdata := New("_testdata")
	files, err := testdata.ReadDir()
	require.NoError(t, err)
	for _, file := range files {
		if file.Base() == "symlinktofolder" {
			err = file.FollowSymLink()
			require.NoError(t, err)
			isDir, err := file.IsDir()
			require.NoError(t, err)
			require.True(t, isDir)
			break
		}
	}
}
