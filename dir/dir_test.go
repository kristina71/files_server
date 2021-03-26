package dir_test

import (
	"files_server/dir"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLs(t *testing.T) {

}

func TestPwd(t *testing.T) {
	testServer := httptest.NewServer(dir.New())
	defer testServer.Close()

	checkPwd(testServer, t, "/Users")
}

func checkPwd(testServer *httptest.Server, t *testing.T, dir string) {
	resp, err := testServer.Client().Get(testServer.URL + "/pwd")
	require.NoError(t, err)

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, dir, string(b))
}

func TestCd(t *testing.T) {
	testServer := httptest.NewServer(dir.New())
	defer testServer.Close()

	dir, err := os.MkdirTemp(os.TempDir(), "example")
	defer os.RemoveAll(dir)
	require.NoError(t, err)

	filePath := filepath.Join(dir, "dir.txt")
	err = os.WriteFile(filePath, []byte("content"), 0666)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		dirName         string
		expected_result int
	}{
		{
			name:            "Change directory",
			dirName:         dir,
			expected_result: http.StatusOK,
		},
		{
			name:            "Not directory",
			dirName:         filePath,
			expected_result: http.StatusBadRequest,
		},
		{
			name:            "No directory",
			dirName:         dir + "/fsdf",
			expected_result: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				resp, err := testServer.Client().Get(testServer.URL + "/cd?dir=" + testCase.dirName)
				require.NoError(t, err)
				resp.Body.Close()
				require.Equal(t, testCase.expected_result, resp.StatusCode)
			},
		)
	}
	checkPwd(testServer, t, dir)
}

//написать тест на GET ls