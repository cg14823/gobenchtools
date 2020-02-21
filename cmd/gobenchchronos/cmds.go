package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func checkoutCommit(location, commit string) error {
	cmd := exec.Command("git", "checkout", commit)
	cmd.Dir = location

	OutBuff := bytes.NewBuffer([]byte{})
	ErrBuff := bytes.NewBuffer([]byte{})

	cmd.Stdout = OutBuff
	cmd.Stderr = ErrBuff

	err := cmd.Run()
	if err == nil {
		return nil
	}

	return fmt.Errorf("error checking out commit: %s : %s : %s", err.Error(), OutBuff.String(), ErrBuff.String())
}

func getCommit(location string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = location

	OutBuff := bytes.NewBuffer([]byte{})
	ErrBuff := bytes.NewBuffer([]byte{})

	cmd.Stdout = OutBuff
	cmd.Stderr = ErrBuff
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not get current commit hash: %s :%s :%s", err.Error(), OutBuff.String(),
			ErrBuff.String())
	}

	return strings.TrimSpace(OutBuff.String()), nil
}

func runBenchmark(location, timeout, runExp, benchExp, pkg string, count int) ([]byte, error) {
	cmd := exec.Command("go", "test", "-timeout="+timeout, "-count="+strconv.Itoa(count), "-run="+runExp,
		"-bench="+benchExp, pkg)
	cmd.Dir = location
	cmd.Env = os.Environ()

	OutBuff := bytes.NewBuffer([]byte{})
	ErrBuff := bytes.NewBuffer([]byte{})

	cmd.Stdout = OutBuff
	cmd.Stderr = ErrBuff
	err := cmd.Run()
	if err != nil {
		return OutBuff.Bytes(), fmt.Errorf("could not run benchmarks `%q`: %s :%s",
			cmd.Args, err.Error(), ErrBuff.String())
	}

	return OutBuff.Bytes(), nil
}
