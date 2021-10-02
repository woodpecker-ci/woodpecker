package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/woodpecker-ci/woodpecker/shared/envsubst"
)

func main() {
	stdin := bufio.NewScanner(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)

	for stdin.Scan() {
		line, err := envsubst.EvalEnv(stdin.Text())
		if err != nil {
			log.Fatalf("Error while envsubst: %v", err)
		}
		_, err = fmt.Fprintln(stdout, line)
		if err != nil {
			log.Fatalf("Error while writing to stdout: %v", err)
		}
		stdout.Flush()
	}
}

