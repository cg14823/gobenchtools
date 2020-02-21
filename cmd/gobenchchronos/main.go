package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cg14823/gobenchtools"
)

type chronosOptions struct {
	repo string

	packageRelative string
	benchRegex      string
	runRegex        string
	benchTimeout    string

	commitList []string
	step       int
	count      int
	numOfSteps int

	tempOutputDir string
}

func argsError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	var repoLocation, outputFile, benchTimeout, packageRelative, benchRegex, runRegex, commitList string
	var numOfSteps, step, count int
	var writeAsCsv bool
	flag.StringVar(&repoLocation, "repo", "", "Local location of the repo.")
	flag.StringVar(&outputFile, "out", "", "Location to write results to")
	flag.StringVar(&packageRelative, "package", "./...", "The package to benchmark, by default it does all")
	flag.StringVar(&benchRegex, "bench-regex", "Benchmark", "The value to pass to go test -bench")
	flag.StringVar(&runRegex, "bench-run-regex", "^$", "The value to pass to go test -run")
	flag.StringVar(&benchTimeout, "bench-timeout", "5m", "Value to pass to go test -timeout")
	flag.StringVar(&commitList, "commits", "", "A comma separated list of commits to checkout if "+
		"provided the step, count and go-back values are ignored")
	flag.IntVar(&count, "bench-count", 1, "Value to pass to go test -count")
	flag.IntVar(&step, "gco-step", 1, "How many commits to move back by in every step")
	flag.IntVar(&numOfSteps, "num-of-steps", 10, "How many times to step back")
	flag.BoolVar(&writeAsCsv, "output-in-csv", false, "The result will be writen to the desired "+
		"out location in csv format.")
	flag.Parse()

	if repoLocation == "" {
		argsError("-repo is required")
	}

	if count == 0 {
		argsError("invalid value `0` for -count")
	}

	if step == 0 {
		argsError("invalid value `0` for -gco-step")
	}

	if numOfSteps == 0 {
		argsError("invalid value `0` for -num-of-steps")
	}

	if outputFile == "" && writeAsCsv {
		argsError("-out is required with -output-in-csv")
	}

	stat, err := os.Stat(repoLocation)
	if err != nil {
		argsError("could not verify repo location:" + err.Error())
	}

	if !stat.IsDir() {
		argsError("location is not a directory")
	}

	dir, err := ioutil.TempDir("", "gobenchchronos")
	if err != nil {
		argsError("Could not create temporary location")
	}

	var commits []string
	if commitList != "" {
		commits = strings.Split(commitList, ",")
	}

	options := &chronosOptions{
		repo:            repoLocation,
		packageRelative: packageRelative,
		benchRegex:      benchRegex,
		runRegex:        runRegex,
		benchTimeout:    benchTimeout,
		commitList:      commits,
		step:            step,
		count:           count,
		numOfSteps:      numOfSteps,
		tempOutputDir:   dir,
	}

	out, err := runChronos(options)
	if err != nil {
		argsError("Could not run chronos: " + err.Error())
	}

	err = outputResults(outputFile, writeAsCsv, out)
	if err != nil {
		argsError("Failed to write output: " + err.Error())
	}
}

func outputResults(outputFile string, writeAsCSV bool, parsed gobenchtools.HistoricPkgBench) error {
	if writeAsCSV {
		return writeOutputAsCSV(outputFile, parsed)
	}

	bOut, err := json.Marshal(&parsed)
	if err != nil {
		return err
	}

	if outputFile == "" {
		fmt.Println(string(bOut))
		return nil
	}

	err = ioutil.WriteFile(outputFile, bOut, gobenchtools.DefaultFileMode)
	if err != nil {
		return err
	}

	return nil
}

func writeOutputAsCSV(outFile string, parsed gobenchtools.HistoricPkgBench) error {
	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, gobenchtools.DefaultFileMode)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(f)
	headers := []string{"ID", "pkg", "commit", "name", "n", "ns_per_op"}
	err = writer.Write(headers)
	if err != nil {
		return err
	}

	for pkgName, benchs := range parsed {
		for name, b := range benchs {
			for _, bb := range b {
				err = writer.Write([]string{
					strconv.Itoa(bb.ID),
					pkgName,
					bb.Commit,
					name,
					strconv.FormatUint(bb.N, 64),
					strconv.FormatFloat(bb.NSPerOp, 'f', -1, 64),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	writer.Flush()
	return nil
}

func runChronos(options *chronosOptions) (gobenchtools.HistoricPkgBench, error) {
	// get original commit so at the end we can go back to it
	commit, err := getCommit(options.repo)
	if err != nil {
		return nil, fmt.Errorf("could not get commit: %w", err)
	}

	defer checkoutCommit(options.repo, commit)

	if len(options.commitList) == 0 {
		getBenchCommitList(options)
	}

	getBenchStepped(options)
	return parseBenchmarkOutputs(options.tempOutputDir)
}

func getBenchCommitList(options *chronosOptions) {
	err := checkoutAndBench("", options, true, 0)
	if err != nil {
		fmt.Println("Failed to benchmark on step: ", 0)
	}

	for i, commit := range options.commitList {
		err = checkoutAndBench(commit, options, false, i+1)
		if err != nil {
			fmt.Println("Failed to benchmark on step: ", i+1)
		}
	}
}

func getBenchStepped(options *chronosOptions) {
	err := checkoutAndBench("", options, true, 0)
	if err != nil {
		fmt.Println("Failed to benchmark on step: ", 0, err)
	}

	commitToCheckout := "HEAD^" + strconv.Itoa(options.step)
	for i := 1; i < options.numOfSteps; i++ {
		err = checkoutAndBench(commitToCheckout, options, false, i)
		if err != nil {
			fmt.Println("Failed to benchmark on step: ", i, err)
		}
	}
}

func checkoutAndBench(commit string, options *chronosOptions, noCheckout bool, id int) error {
	if !noCheckout {
		err := checkoutCommit(options.repo, commit)
		if err != nil {
			return err
		}
	}

	commitHash, err := getCommit(options.repo)
	if err != nil {
		return err
	}

	fmt.Println("Running benchmarks for: ", commitHash)
	out, err := runBenchmark(options.repo, options.benchTimeout, options.runRegex, options.benchRegex,
		options.packageRelative, options.count)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(options.tempOutputDir, strconv.Itoa(id)+"-bench-out-"+commitHash), out,
		gobenchtools.DefaultFileMode)
}

func parseBenchmarkOutputs(dir string) (gobenchtools.HistoricPkgBench, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	historic := gobenchtools.HistoricPkgBench{}
	fileRegex := regexp.MustCompile(`\d+-bench-out-.+`)
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if !fileRegex.MatchString(f.Name()) {
			continue
		}

		var count int
		var commitHash string
		n, err := fmt.Sscanf(f.Name(), "%d-bench-out-%s", &count, &commitHash)
		if err != nil || n != 2 {
			continue
		}

		fmt.Println("Parsing file: ", filepath.Join(dir, f.Name()))
		parsed, err := gobenchtools.ParseFile(filepath.Join(dir, f.Name()))
		if err != nil {
			fmt.Printf("Could not parse output for file %s: %s\n", f.Name(), err.Error())
		}

		// add file parsed to historic data
		for pkgName, benchs := range parsed {
			_, ok := historic[pkgName]
			if !ok {
				historic[pkgName] = make(map[string][]gobenchtools.Benchmark)
			}

			for _, b := range benchs {
				_, ok := historic[pkgName][b.Name]
				if !ok {
					historic[pkgName][b.Name] = make([]gobenchtools.Benchmark, 0)
				}

				historic[pkgName][b.Name] = append(historic[pkgName][b.Name], gobenchtools.Benchmark{
					ID:      n,
					Commit:  commitHash,
					N:       b.N,
					NSPerOp: b.NSPerOp,
				})
			}
		}
	}

	return historic, nil
}
