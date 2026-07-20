# linear-stats

A small command-line program that computes two statistics from a column of
numbers: the **Linear Regression Line** and the **Pearson Correlation
Coefficient**.

The input file is treated as a series of `(x, y)` points, where `x` is the
line number (0, 1, 2, 3, ...) and `y` is the numeric value on that line.

## Purpose

This program answers two related questions about a data set:

- **Linear Regression Line** — What straight line `y = ax + b` best fits the
  data? This is the least-squares regression line: the line that minimises
  the total squared vertical distance between the points and the line. It
  summarises the overall trend (rising, falling, or flat) and can be used to
  predict `y` for a given `x`.
- **Pearson Correlation Coefficient** — How strong and in which direction is
  the linear relationship between `x` and `y`? The result `r` is always
  between `-1` and `1`:
  - `r` near `+1` — strong positive linear relationship (y rises with x).
  - `r` near `-1` — strong negative linear relationship (y falls as x rises).
  - `r` near `0` — little to no linear relationship.

Together they describe *what* the trend line is and *how well* the data
actually follows it.

## Requirements

- [Go](https://go.dev/) 1.21 or later

## Usage

```sh
go run main.go <path-to-data-file>
```

The data file should contain one number per line, for example (`data.txt`):

```
189
113
121
114
145
110
```

### Example

```sh
$ go run main.go data.txt
Linear Regression Line: y = -8.742857x + 153.857143
Pearson Correlation Coefficient: -0.5330331012
```

### Output format

- `Linear Regression Line: y = <a>x + <b>` — `a` (slope) and `b` (intercept)
  are printed with **6** decimal places.
- `Pearson Correlation Coefficient: <r>` — printed with **10** decimal
  places.

### Errors

The program exits with a non-zero status and a message on standard error
when:

- no data file argument is provided (prints a usage message);
- the file cannot be opened, or a line is not a valid number;
- the file contains fewer than 2 data points (both statistics are undefined
  for a single point, since the formulas would divide by zero).

## Building a binary

```sh
go build -o linear-stats main.go
./linear-stats data.txt
```

## Testing

The test suite lives in [main_test.go](main_test.go). Run it with:

```sh
go test -v ./...
```

It covers the following cases:

| Test | What it verifies |
| --- | --- |
| `TestLinearRegression` | Slope and intercept for the spec example, a perfect `y = 2x + 1` line (exact recovery), and a flat line (slope `0`). |
| `TestPearsonCorrelation` | Correlation for the spec example, a perfect positive line (`r = +1`), and a perfect negative line (`r = -1`). |
| `TestReadData` | Parsing one value per line, ignoring blank lines and trimming surrounding whitespace. |
| `TestReadDataErrors` | Errors are reported for a missing file and for a non-numeric line. |
| `TestEndToEnd` | Builds and runs the real binary, checking the exact formatted output (6 / 10 decimal places) and the no-argument usage guard. |

> **Note on floating-point comparisons:** the numeric tests compare results
> within a small epsilon rather than with `==`, because rounding differences
> in the least significant bits make exact float equality unreliable. The
> tolerance is still tight enough to guarantee the required precision.

## How it works

Given `n` points `(x_i, y_i)`, both statistics are derived from the same set
of running sums (Σx, Σy, Σxy, Σx², Σy²):

- **Slope**: `a = (nΣxy − ΣxΣy) / (nΣx² − (Σx)²)`
- **Intercept**: `b = (Σy − aΣx) / n`
- **Pearson r**: `(nΣxy − ΣxΣy) / √((nΣx² − (Σx)²)(nΣy² − (Σy)²))`

See the comments in [main.go](main.go) for the implementation details.
