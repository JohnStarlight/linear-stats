package main

import (
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// floatEqual reports whether two float64 values are equal within a small
// tolerance. Direct equality (==) is unreliable for floating-point results
// because rounding differences can appear in the least significant bits, so
// tests compare against an epsilon instead.
func floatEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// makeXY builds the x/y slices the way main() does: x is the 0-based index
// of each y value (0, 1, 2, ...). This helper keeps the test cases focused
// on just the y values.
func makeXY(y []float64) (x, yOut []float64) {
	x = make([]float64, len(y))
	for i := range y {
		x[i] = float64(i)
	}
	return x, y
}

// TestLinearRegression checks the slope and intercept against values that
// were computed independently. The tolerance is looser than the 6-decimal
// output format because we only need to confirm the math is correct, not
// the exact rounding of the printed string.
func TestLinearRegression(t *testing.T) {
	tests := []struct {
		name          string
		y             []float64
		wantSlope     float64
		wantIntercept float64
	}{
		{
			// Spec's example data set.
			name:          "spec example",
			y:             []float64{189, 113, 121, 114, 145, 110},
			wantSlope:     -8.742857,
			wantIntercept: 153.857143,
		},
		{
			// Perfectly increasing line y = 2x + 1: slope and intercept
			// must be recovered exactly.
			name:          "perfect positive line",
			y:             []float64{1, 3, 5, 7, 9},
			wantSlope:     2.0,
			wantIntercept: 1.0,
		},
		{
			// Flat data: the best-fit line is horizontal, so the slope is 0
			// and the intercept equals the constant value.
			name:          "flat line",
			y:             []float64{5, 5, 5, 5},
			wantSlope:     0.0,
			wantIntercept: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := makeXY(tt.y)
			slope, intercept := linearRegression(x, y)
			if !floatEqual(slope, tt.wantSlope, 1e-6) {
				t.Errorf("slope = %.6f, want %.6f", slope, tt.wantSlope)
			}
			if !floatEqual(intercept, tt.wantIntercept, 1e-6) {
				t.Errorf("intercept = %.6f, want %.6f", intercept, tt.wantIntercept)
			}
		})
	}
}

// TestPearsonCorrelation checks the correlation coefficient for cases where
// the expected value is known exactly: a perfect positive line (+1) and a
// perfect negative line (-1), plus the spec example.
func TestPearsonCorrelation(t *testing.T) {
	tests := []struct {
		name  string
		y     []float64
		wantR float64
	}{
		{
			name:  "spec example",
			y:     []float64{189, 113, 121, 114, 145, 110},
			wantR: -0.5330331012,
		},
		{
			// y increases perfectly with x, so correlation is exactly +1.
			name:  "perfect positive correlation",
			y:     []float64{0, 1, 2, 3, 4},
			wantR: 1.0,
		},
		{
			// y decreases perfectly as x increases, so correlation is -1.
			name:  "perfect negative correlation",
			y:     []float64{4, 3, 2, 1, 0},
			wantR: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := makeXY(tt.y)
			r := pearsonCorrelation(x, y)
			if !floatEqual(r, tt.wantR, 1e-10) {
				t.Errorf("r = %.10f, want %.10f", r, tt.wantR)
			}
		})
	}
}

// TestReadData verifies that readData parses one value per line and ignores
// blank lines (including trailing newlines and surrounding whitespace).
func TestReadData(t *testing.T) {
	// Write a temporary data file that includes a trailing newline and some
	// surrounding whitespace to exercise the trimming/blank-line handling.
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	content := "189\n 113 \n121\n\n114\n145\n110\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	got, err := readData(path)
	if err != nil {
		t.Fatalf("readData returned error: %v", err)
	}

	want := []float64{189, 113, 121, 114, 145, 110}
	if len(got) != len(want) {
		t.Fatalf("got %d values, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("value[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestReadDataErrors confirms that readData reports errors for a missing
// file and for a file containing a non-numeric line, rather than silently
// succeeding.
func TestReadDataErrors(t *testing.T) {
	// Missing file.
	if _, err := readData(filepath.Join(t.TempDir(), "does-not-exist.txt")); err == nil {
		t.Error("expected error for missing file, got nil")
	}

	// File with an invalid (non-numeric) line.
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(path, []byte("1\n2\nnot-a-number\n"), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := readData(path); err == nil {
		t.Error("expected error for non-numeric line, got nil")
	}
}

// TestEndToEnd builds and runs the program as a user would, then checks the
// full formatted output. This guards the exact output strings (6 and 10
// decimal places) required by the assignment, which the unit tests above do
// not cover.
func TestEndToEnd(t *testing.T) {
	dir := t.TempDir()

	// Write the spec's example data to a file.
	dataPath := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(dataPath, []byte("189\n113\n121\n114\n145\n110\n"), 0o644); err != nil {
		t.Fatalf("failed to write data file: %v", err)
	}

	// Compile the program into a temporary binary so `go test` doesn't need
	// the source path at runtime.
	binPath := filepath.Join(dir, "linear-stats")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Run the binary against the data file and capture its stdout.
	out, err := exec.Command(binPath, dataPath).Output()
	if err != nil {
		t.Fatalf("running program failed: %v", err)
	}

	want := "Linear Regression Line: y = -8.742857x + 153.857143\n" +
		"Pearson Correlation Coefficient: -0.5330331012\n"
	if got := string(out); got != want {
		t.Errorf("output mismatch:\n got: %q\nwant: %q", got, want)
	}

	// Sanity-check the argument-count guard: running with no file should
	// exit non-zero and mention usage.
	noArg := exec.Command(binPath)
	noArgOut, err := noArg.CombinedOutput()
	if err == nil {
		t.Error("expected non-zero exit when run without a data file")
	}
	if !strings.Contains(string(noArgOut), "usage") {
		t.Errorf("expected usage message, got %q", noArgOut)
	}
}
