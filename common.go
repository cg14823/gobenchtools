package gobenchtools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const DefaultFileMode = 0660

// ParsedBench is a map of package name to a Slice containing all the benchmarks for that slice.
type ParsedBench map[string][]Benchmark

type Benchmark struct {
	ID      int     `json:"id,omitempty"`
	Commit  string  `json:"commit,omitempty"`
	Name    string  `json:"name,omitempty"`
	N       uint64  `json:"n"`
	NSPerOp float64 `json:"ns_per_op"`
}

var PkgNameExp = regexp.MustCompile(`^pkg: (?P<pkgName>[0-9A-Za-z_\-/.]+)`)
var BenchMarkResultExp = regexp.MustCompile(`^(?P<benchName>[0-9A-Za-z_\-/.]+)\s+(?P<n>\d+)\s+(?P<nsPerOp>\d+(?:\.\d+)?) ns/op`)

func ParseFile(fileIn string) (ParsedBench, error) {
	f, err := os.OpenFile(fileIn, os.O_RDONLY, DefaultFileMode)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	parsed := ParsedBench(make(map[string][]Benchmark))

	var currentPkg string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		line = strings.TrimSpace(line)

		match := PkgNameExp.FindAllStringSubmatch(line, -1)
		// got a package line
		if len(match) == 1 && len(match[0]) == 2 {
			currentPkg = match[0][1]
			parsed[currentPkg] = make([]Benchmark, 0)
			continue
		}

		if currentPkg == "" {
			continue
		}

		match = BenchMarkResultExp.FindAllStringSubmatch(line, -1)
		// if not the expected match then skip
		if len(match) != 1 || len(match[0]) != 4 {
			continue
		}

		n, err := strconv.ParseUint(match[0][2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse loop count: %w", err)
		}

		nsPerOp, err := strconv.ParseFloat(match[0][3], 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse ns per op: %w", err)
		}

		parsed[currentPkg] = append(parsed[currentPkg], Benchmark{
			Name:    match[0][1],
			N:       n,
			NSPerOp: nsPerOp,
		})
	}

	return parsed, nil
}

// HistoryPkgBench map package to a map of benchmark name to benchmark results
type HistoricPkgBench map[string]map[string][]Benchmark
