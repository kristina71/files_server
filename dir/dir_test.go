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

func initTestEnv(t *testing.T) (string, *httptest.Server) {
	testServer := httptest.NewServer(dir.New())

	dir, err := os.MkdirTemp(os.TempDir(), "example")
	require.NoError(t, err)

	_, err = testServer.Client().Get(testServer.URL + "/cd?dir=" + dir)
	require.NoError(t, err)
	return dir, testServer
}

type expectation struct {
	path         string
	expected_dir string
}

type testCase struct {
	name            string
	path            string
	testServer      *httptest.Server
	dirname         string
	expected_result int
	expectations    []expectation
	prepare         func(t *testing.T, tc *testCase)
}

func (tc *testCase) init(t *testing.T) {
	tc.path, tc.testServer = initTestEnv(t)
	if tc.prepare != nil {
		tc.prepare(t, tc)
	}
}

func (tc *testCase) close(t *testing.T) {
	tc.testServer.Close()
	os.RemoveAll(tc.path)
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

func TestLs(t *testing.T) {
	testServer := httptest.NewServer(dir.New())
	defer testServer.Close()

	dir, err := os.MkdirTemp(os.TempDir(), "example")
	defer os.RemoveAll(dir)
	require.NoError(t, err)

	filePath := filepath.Join(dir, "dir.txt")
	err = os.WriteFile(filePath, []byte("content"), 0666)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(dir, "dir"), 0666)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, ".htaccess"), []byte("content"), 0666)
	require.NoError(t, err)

	testServer.Client().Get(testServer.URL + "/cd?dir=" + dir)

	innerDir := filepath.Join(dir, "dir")

	testCases := []struct {
		name            string
		hidden          string
		expected_result int
		expected_files  string
		prepare         func(t *testing.T)
	}{
		{
			name:            "Get directory",
			hidden:          "false",
			expected_result: http.StatusOK,
			expected_files:  "[\"dir\",\"dir.txt\"]",
			prepare:         func(t *testing.T) {},
		},
		{
			name:            "Get directory with hidden",
			hidden:          "true",
			expected_result: http.StatusOK,
			expected_files:  "[\"dir\",\".htaccess\",\"dir.txt\"]",
			prepare:         func(t *testing.T) {},
		},
		{
			name:            "Empty dir",
			hidden:          "true",
			expected_result: http.StatusOK,
			expected_files:  "[]",
			prepare: func(t *testing.T) {
				_, err := testServer.Client().Get(testServer.URL + "/cd?dir=" + innerDir)
				require.NoError(t, err)
			},
		},
		{
			name:            "Error dir",
			hidden:          "true",
			expected_result: http.StatusInternalServerError,
			expected_files:  "open " + innerDir + ": no such file or directory\n",
			prepare: func(t *testing.T) {
				_, err := testServer.Client().Get(testServer.URL + "/cd?dir=" + innerDir)
				require.NoError(t, err)
				os.RemoveAll(innerDir)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				testCase.prepare(t)
				resp, err := testServer.Client().Get(testServer.URL + "/ls?hide=" + testCase.hidden)
				require.NoError(t, err)
				b, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				require.NoError(t, err)
				require.Equal(t, testCase.expected_result, resp.StatusCode)
				require.Equal(t, testCase.expected_files, string(b))
			},
		)
	}
}

func TestMkDir(t *testing.T) {
	testCases := []testCase{
		{
			name:    "Create directory",
			dirname: "dir",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"dir\"]",
					},
				}
			},
		},
		{
			name:    "Create recursive directory",
			dirname: "dir/dir1/dir2",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"dir\"]",
					},
					{
						path:         filepath.Join(tc.path, "dir"),
						expected_dir: "[\"dir1\"]",
					},
					{
						path:         filepath.Join(filepath.Join(tc.path, "dir"), "dir1"),
						expected_dir: "[\"dir2\"]",
					},
				}
			},
		},
		{
			name:    "Empty directory",
			dirname: "",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusBadRequest
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[]",
					},
				}
			},
		},
		{
			name:    "Create file-directory",
			dirname: "dir.txt",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"dir.txt\"]",
					},
				}
			},
		},
		{
			name:            "Absolute directory",
			expected_result: http.StatusOK,
			prepare: func(t *testing.T, tc *testCase) {
				tc.dirname = filepath.Join(tc.path, "abs_dir")
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"abs_dir\"]",
					},
				}
			},
		},
		{
			name:    ". directory",
			dirname: "./dir",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"dir\"]",
					},
				}
			},
		},
		{
			name:    ".. directory",
			dirname: "../dir5",
			prepare: func(t *testing.T, tc *testCase) {
				tc.testServer.Client().Get(tc.testServer.URL + "/mkdir?dirname=dir1")
				tc.testServer.Client().Get(tc.testServer.URL + "/cd?dir=dir1")
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"dir1\",\"dir5\"]",
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				testCase.init(t)
				defer testCase.close(t)
				resp, err := testCase.testServer.Client().Get(testCase.testServer.URL + "/mkdir?dirname=" + testCase.dirname)
				require.NoError(t, err)
				resp.Body.Close()
				require.Equal(t, testCase.expected_result, resp.StatusCode)

				for _, expectation := range testCase.expectations {
					resp, err = testCase.testServer.Client().Get(testCase.testServer.URL + "/cd?dir=" + expectation.path)
					require.NoError(t, err)
					resp.Body.Close()
					require.Equal(t, http.StatusOK, resp.StatusCode)

					resp1, err := testCase.testServer.Client().Get(testCase.testServer.URL + "/ls")
					require.NoError(t, err)
					b1, err := ioutil.ReadAll(resp1.Body)
					resp1.Body.Close()
					require.NoError(t, err)

					require.Equal(t, expectation.expected_dir, string(b1))
				}
			},
		)
	}
}

func TestTouch(t *testing.T) {
	testCases := []testCase{
		{
			name:    "Create file",
			dirname: "dir.txt",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"dir.txt\"]",
					},
				}
			},
		},
		{
			name:    "Empty file",
			dirname: "",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusBadRequest
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[]",
					},
				}
			},
		},
		{
			name: "Absolute file",
			prepare: func(t *testing.T, tc *testCase) {
				tc.dirname = filepath.Join(tc.path, "abs_file.txt")
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"abs_file.txt\"]",
					},
				}
			},
		},
		{
			name:    ". file",
			dirname: "./file.txt",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[\"file.txt\"]",
					},
				}
			},
		},
		{
			name:    "../file",
			dirname: "../file.txt",
			prepare: func(t *testing.T, tc *testCase) {
				tc.expected_result = http.StatusOK
				tc.expectations = []expectation{
					{
						path:         tc.path,
						expected_dir: "[]",
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				testCase.init(t)
				defer testCase.close(t)
				for _, expectation := range testCase.expectations {
					resp, err := testCase.testServer.Client().Get(testCase.testServer.URL + "/touch?filename=" + testCase.dirname)
					require.NoError(t, err)
					resp.Body.Close()

					require.Equal(t, testCase.expected_result, resp.StatusCode)

					resp, err = testCase.testServer.Client().Get(testCase.testServer.URL + "/ls")
					require.NoError(t, err)
					b, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					require.NoError(t, err)
					require.Equal(t, expectation.expected_dir, string(b))
				}
			},
		)
	}
}
