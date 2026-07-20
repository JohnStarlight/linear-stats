// Command linear-stats reads a list of numbers from a file and prints two
// statistics computed from them: the Linear Regression Line and the
// Pearson Correlation Coefficient.
//
// The numbers in the file are treated as the y-values of a data set, one
// point per line. The corresponding x-value of each point is simply its
// line index, starting at 0 (line 1 -> x=0, line 2 -> x=1, and so on).
//
// Usage:
//
//	go run main.go <path-to-data-file>
package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// readData opens the file at path and parses it into a slice of float64
// values, one per non-empty line. Blank lines are skipped so that trailing
// newlines in the input file don't produce spurious data points.
func readData(path string) ([]float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var values []float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		v, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid data %q: %w", line, err)
		}
		values = append(values, v)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

// linearRegression computes the slope and intercept of the least-squares
// regression line y = slope*x + intercept for the given set of points.
//
// It uses the standard closed-form (normal equations) solution:
//
//	slope     = (nΣxy - ΣxΣy) / (nΣx² - (Σx)²)
//	intercept = (Σy - slope*Σx) / n
//
// x and y must have the same length and contain at least 2 points, since
// the denominator (nΣx² - (Σx)²) is zero when there is only one point (or
// when all x values are identical), which would produce a division by
// zero.
func linearRegression(x, y []float64) (slope, intercept float64) {
	n := float64(len(x))

	// Accumulate the sums needed by the normal equations in a single pass
	// over the data.
	var sumX, sumY, sumXY, sumX2 float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept = (sumY - slope*sumX) / n
	return
}

// pearsonCorrelation computes Pearson's correlation coefficient r for the
// given set of points, using the same running-sums approach as
// linearRegression:
//
//	r = (nΣxy - ΣxΣy) / sqrt((nΣx² - (Σx)²) * (nΣy² - (Σy)²))
//
// The result is a value in [-1, 1] describing how well the points fit a
// straight line: values near 1 or -1 indicate a strong linear relationship
// (positive or negative), while values near 0 indicate little to no linear
// relationship.
func pearsonCorrelation(x, y []float64) float64 {
	n := float64(len(x))

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
	return numerator / denominator
}

func main() {
	// The path to the data file is expected as the first command-line
	// argument, e.g. `go run main.go data.txt`.
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run main.go <data-file>")
		os.Exit(1)
	}

	// The file only contains y-values; read them in as-is.
	y, err := readData(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// Both statistics are undefined (division by zero) with fewer than 2
	// points, so fail fast with a clear error instead of printing NaN/Inf.
	if len(y) < 2 {
		fmt.Fprintln(os.Stderr, "error: need at least 2 data points")
		os.Exit(1)
	}

	// Reconstruct the x-values as the 0-based line index of each y-value,
	// per the assignment's definition of the data set (0, 1, 2, 3, ...).
	x := make([]float64, len(y))
	for i := range y {
		x[i] = float64(i)
	}

	slope, intercept := linearRegression(x, y)
	r := pearsonCorrelation(x, y)

	// Required output format: regression coefficients with 6 decimal
	// places, Pearson coefficient with 10 decimal places.
	fmt.Printf("Linear Regression Line: y = %.6fx + %.6f\n", slope, intercept)
	fmt.Printf("Pearson Correlation Coefficient: %.10f\n", r)
}
