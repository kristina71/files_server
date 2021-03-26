package commands_test

import (
	"files_server/commands"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLs(t *testing.T) {

	testCases := []struct {
		name            string
		expected_result []string
		error_checker   func(t require.TestingT, err error, msgAndArgs ...interface{})
		prepare         func(t *testing.T) string
		flag            bool
	}{
		{
			name:            "Empty Dir",
			expected_result: []string{},
			error_checker:   require.NoError,
			prepare: func(t *testing.T) string {
				dir, err := os.MkdirTemp(os.TempDir(), "example")
				require.NoError(t, err)
				return dir
			},
			flag: false,
		},
		{
			name:            "Ls no Dir",
			expected_result: nil,
			error_checker:   require.Error,
			prepare: func(t *testing.T) string {
				return filepath.Join(os.TempDir(), "example_not_exists")
			},
			flag: false,
		},
		{
			name:            "Ls test1",
			expected_result: []string{"dir", "temp_dir", "dir.txt"},
			error_checker:   require.NoError,
			prepare: func(t *testing.T) string {
				dir, err := os.MkdirTemp(os.TempDir(), "example")
				require.NoError(t, err)
		
				err = os.Mkdir(filepath.Join(dir, "temp_dir"), 0666)
				require.NoError(t, err)

				err = os.Mkdir(filepath.Join(dir, "dir"), 0666)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(dir, "dir.txt"), []byte("content"), 0666)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(dir, ".htaccess"), []byte("content"), 0666)
				require.NoError(t, err)

				return dir
			},
			flag: false,
		},
		{
			name:            "Ls Show Hidden",
			expected_result: []string{"dir", "temp_dir", ".htaccess", "dir.txt"},
			error_checker:   require.NoError,
			prepare: func(t *testing.T) string {
				dir, err := os.MkdirTemp(os.TempDir(), "example")
				require.NoError(t, err)
			
				err = os.Mkdir(filepath.Join(dir, "temp_dir"), 0666)
				require.NoError(t, err)

				err = os.Mkdir(filepath.Join(dir, "dir"), 0666)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(dir, "dir.txt"), []byte("content"), 0666)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(dir, ".htaccess"), []byte("content"), 0666)
				require.NoError(t, err)

				return dir
			},
			flag: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				dir := testCase.prepare(t)

				defer os.RemoveAll(dir)
				ls_res, err := commands.Ls(dir, testCase.flag)
				testCase.error_checker(t, err)
				require.Equal(t, testCase.expected_result, ls_res)
			},
		)
	}

}
