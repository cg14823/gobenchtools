package gobenchtools

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func getTestDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "test-dir")
	if err != nil {
		t.Fatalf("Could not aquite test dir: %s", err.Error())
	}

	return dir
}

func cleanUpTestDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Could not clean up test dir: %s", err.Error())
	}
}

func TestParseFile(t *testing.T) {
	type testCase struct {
		name          string
		input         string
		expectedOut   ParsedBench
		expectedError bool
	}

	tests := []testCase{
		{
			name: "one-package",
			input: "goos: darwin\ngoarch: amd64\npkg: github.com/user/repo/package\nBenchmarkUnpackMetaData-12      35928573                28.1 ns/op\nBenchmark_Set/sqlite-50b-12        83124             15066 ns/op\nPASS\nok      github.com/user/repo/package     230.051s",
			expectedOut: ParsedBench{
				"github.com/user/repo/package": []Benchmark{
					{
						Name: "BenchmarkUnpackMetaData-12",
						N:35928573,
						NSPerOp: 28.1,
					},
					{
						Name: "Benchmark_Set/sqlite-50b-12",
						N: 83124,
						NSPerOp: 15066,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := getTestDir(t)
			defer cleanUpTestDir(t, dir)

			// write the file to be read
			filePath := filepath.Join(dir, tc.name)
			err := ioutil.WriteFile(filePath, []byte(tc.input), DefaultFileMode)
			if err != nil {
				t.Fatalf("Could not write test data: %s", err.Error())
			}

			out, err := ParseFile(filePath)
			if err != nil {
				if tc.expectedError {
					return
				}

				t.Fatalf("Unexpected error: %s", err.Error())
			}

			if tc.expectedError {
				t.Fatalf("Expected an error but got nil")
			}

			if !reflect.DeepEqual(tc.expectedOut, out) {
				t.Fatalf("Expected %+v got %+v", tc.expectedOut, out)
			}
		})
	}
}