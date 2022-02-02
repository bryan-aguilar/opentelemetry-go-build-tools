package crosslink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

var (
	testDataDir, _ = filepath.Abs("../test_data")
	mockDataDir, _ = filepath.Abs("../mock_test_data")
)

func TestCrosslink(t *testing.T) {
	tests := []struct {
		testName string
		config   runConfig
		expected map[string][]byte
	}{
		{
			testName: "testSimple",
			config:   DefaultRunConfig(),
			expected: map[string][]byte{
				filepath.Join("go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.17\n\n"),
			},
		},
		{
			testName: "testCyclic",
			config:   DefaultRunConfig(),
			expected: map[string][]byte{
				filepath.Join("go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot => ../"),
				// b has req on root but not neccessary to write out with current comparison logic
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.17\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot => ../\n\n"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tmpRootDir, err := os.MkdirTemp(testDataDir, test.testName)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			mockDataDir := filepath.Join(mockDataDir, test.testName)
			cp.Copy(mockDataDir, tmpRootDir)
			test.config.RootPath = tmpRootDir
			assert.NotPanics(t, func() { Crosslink(test.config) })

			if assert.NoError(t, err, "error message on execution %s") {

				for modFilePath, modFilesExpected := range test.expected {
					modFileActual, err := os.ReadFile(filepath.Join(tmpRootDir, modFilePath))

					if err != nil {
						t.Fatalf("error reading actual mod files: %v", err)
					}

					actual, err := modfile.Parse("go.mod", modFileActual, nil)
					if err != nil {
						t.Fatalf("error decoding original mod files: %v", err)
					}
					actual.Cleanup()

					expected, err := modfile.Parse("go.mod", modFilesExpected, nil)
					if err != nil {
						t.Fatalf("error decoding expected mod file: %v", err)
					}
					expected.Cleanup()

					// replace structs need to be assorted to avoid flaky fails in test
					replaceSortFunc := func(x, y *modfile.Replace) bool {
						return x.Old.Path < y.Old.Path
					}

					if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
						cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax"),
						cmpopts.SortSlices(replaceSortFunc),
					); diff != "" {
						t.Errorf("Replace{} mismatch (-want +got):\n%s", diff)
					}
				}
			}
			os.RemoveAll(tmpRootDir)
		})
	}
}

func TestOverwrite(t *testing.T) {
	lg, _ := zap.NewProduction()
	defer lg.Sync()

	tests := []struct {
		testName string
		config   runConfig
		expected map[string][]byte
	}{
		{
			testName: "testOverwrite",
			config: runConfig{
				Verbose:       true,
				Overwrite:     true,
				ExcludedPaths: map[string]struct{}{},
				logger:        lg,
			},
			expected: map[string][]byte{
				filepath.Join("go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.17\n\n"),
			},
		},
		{
			testName: "testNoOverwrite",
			config: runConfig{
				ExcludedPaths: map[string]struct{}{},
				Verbose:       true,
				logger:        lg,
			},
			expected: map[string][]byte{
				filepath.Join("go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.17\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.17\n\n"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tmpRootDir, err := os.MkdirTemp(testDataDir, test.testName)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			mockDataDir := filepath.Join(mockDataDir, test.testName)
			cp.Copy(mockDataDir, tmpRootDir)
			test.config.RootPath = tmpRootDir

			assert.NotPanics(t, func() { Crosslink(test.config) })

			if assert.NoError(t, err, "error message on execution %s") {
				// a mock_test_data_expected folder could be built instead of building expected files by hand.

				for modFilePath, modFilesExpected := range test.expected {
					modFileActual, err := os.ReadFile(filepath.Join(tmpRootDir, modFilePath))

					if err != nil {
						t.Fatalf("error reading actual mod files: %v", err)
					}

					actual, err := modfile.Parse("go.mod", modFileActual, nil)
					if err != nil {
						t.Fatalf("error decoding original mod files: %v", err)
					}
					actual.Cleanup()

					expected, err := modfile.Parse("go.mod", modFilesExpected, nil)
					if err != nil {
						t.Fatalf("error decoding expected mod file: %v", err)
					}
					expected.Cleanup()

					// replace structs need to be assorted to avoid flaky fails in test
					replaceSortFunc := func(x, y *modfile.Replace) bool {
						return x.Old.Path < y.Old.Path
					}

					if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
						cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax"),
						cmpopts.SortSlices(replaceSortFunc),
					); diff != "" {
						t.Errorf("Replace{} mismatch (-want +got):\n%s", diff)
					}
				}
			}

			os.RemoveAll(tmpRootDir)
		})
	}

}

// Testing exclude functionality for prune, overwrite, and no overwrite.
func TestExclude(t *testing.T) {
	testName := "testExclude"
	lg, _ := zap.NewProduction()
	tests := []struct {
		testCase string
		config   runConfig
	}{
		{
			testCase: "Overwrite off",
			config: runConfig{
				Prune: true,
				ExcludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				Verbose: true,
				logger:  lg,
			},
		},
		{
			testCase: "Overwrite on",
			config: runConfig{
				Overwrite: true,
				Prune:     true,
				ExcludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				logger:  lg,
				Verbose: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCase, func(t *testing.T) {
			tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			mockDataDir := filepath.Join(mockDataDir, testName)
			cp.Copy(mockDataDir, tmpRootDir)

			assert.NotPanics(t, func() { Crosslink(tc.config) })
			if assert.NoError(t, err, "error message on execution %s") {
				// a mock_test_data_expected folder could be built instead of building expected files by hand.
				modFilesExpected := map[string][]byte{
					filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
						"go 1.17\n\n" +
						"require (\n\t" +
						"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
						")\n" +
						"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
						"replace go.opentelemetry.io/build-tools/excludeme => ../excludeme\n\n"),
					filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
						"go 1.17\n\n" +
						"require (\n\t" +
						"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
						")\n" +
						"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
					filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
						"go 1.17\n\n"),
				}

				for modFilePath, modFilesExpected := range modFilesExpected {
					modFileActual, err := os.ReadFile(modFilePath)

					if err != nil {
						t.Fatalf("TestCase: %s, error reading actual mod files: %v", tc.testCase, err)
					}

					actual, err := modfile.Parse("go.mod", modFileActual, nil)
					if err != nil {
						t.Fatalf("error decoding original mod files: %v", err)
					}
					actual.Cleanup()

					expected, err := modfile.Parse("go.mod", modFilesExpected, nil)
					if err != nil {
						t.Fatalf("TestCase: %s ,error decoding expected mod file: %v", tc.testCase, err)
					}
					expected.Cleanup()

					// replace structs need to be assorted to avoid flaky fails in test
					replaceSortFunc := func(x, y *modfile.Replace) bool {
						return x.Old.Path < y.Old.Path
					}

					if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
						cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax"),
						cmpopts.SortSlices(replaceSortFunc),
					); diff != "" {
						t.Errorf("TestCase: %s \n Replace{} mismatch (-want +got):\n%s", tc.testCase, diff)
					}
				}
			}
			os.RemoveAll(tmpRootDir)
		})
	}
}
