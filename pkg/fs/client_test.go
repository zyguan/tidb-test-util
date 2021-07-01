package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteStatDelete(t *testing.T) {
	cli := Default()
	require.NoError(t, cli.Auth())

	testFile := "/pingcap/qa/draft/" + t.Name()

	// write
	hash, err := cli.Write(testFile, String("hello"))
	require.NoError(t, err)
	require.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hash)

	// stat
	info, err := cli.Stat(testFile, true)
	require.NoError(t, err)
	require.Equal(t, hash, info.Checksum.SHA256)

	// delete
	require.NoError(t, cli.Delete(testFile))

}

func TestPutGetDelFile(t *testing.T) {
	cli := Default()
	require.NoError(t, cli.Auth())

	tempDir := t.TempDir()
	testFile := "/pingcap/qa/draft/" + t.Name()

	f, err := os.Open("client_test.go")
	require.NoError(t, err)
	defer f.Close()

	require.NoError(t, cli.PutFile(testFile+ExtPartial, File(f), true))

	// partial file exists
	_, err = f.Seek(0, 0)
	require.NoError(t, err)
	require.Error(t, cli.PutFile(testFile, File(f), false))

	// partial file exists but force write it
	_, err = f.Seek(0, 0)
	require.NoError(t, err)
	require.NoError(t, cli.PutFile(testFile, File(f), true))

	// partial file has been deleted by force write
	_, err = f.Seek(0, 0)
	require.NoError(t, err)
	require.NoError(t, cli.PutFile(testFile, File(f), false))

	require.NoError(t, cli.GetFile(testFile, filepath.Join(tempDir, t.Name())))

	require.NoError(t, cli.DelFile(testFile, false))
	require.NoError(t, cli.Delete(testFile+ExtPartial+ExtChecksum))
}
