package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/1cbyc/tfkit/internal/policy"
)

func main() {
	flag.Usage = func() { fmt.Fprintln(flag.CommandLine.Output(), "usage: tfcheck [module-directory]") }
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}
	dir := "."
	if flag.NArg() == 1 {
		dir = flag.Arg(0)
	}
	findings, err := policy.CheckTree(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tfcheck: %v\n", err)
		os.Exit(2)
	}
	for _, finding := range findings {
		fmt.Println(finding)
	}
	if len(findings) != 0 {
		fmt.Printf("tfcheck: %d finding(s)\n", len(findings))
		os.Exit(1)
	}
	fmt.Println("tfcheck: policy passed")
}
