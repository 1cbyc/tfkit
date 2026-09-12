package main

import (
	"flag"
	"fmt"
	"os"

	plancheck "github.com/1cbyc/tfkit/internal/plan"
)

func main() {
	approved := flag.Bool("approve-destructive", false, "acknowledge deletes and replacements")
	flag.Usage = func() { fmt.Fprintln(flag.CommandLine.Output(), "usage: tfplan [-approve-destructive] plan.json") }
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "tfplan: %v\n", err)
		os.Exit(2)
	}
	defer f.Close()
	report, err := plancheck.Analyze(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tfplan: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("create=%d update=%d delete=%d replace=%d no-op=%d\n", report.Create, report.Update, report.Delete, report.Replace, report.NoOp)
	for _, address := range report.Destructive {
		fmt.Printf("destructive: %s\n", address)
	}
	if len(report.Destructive) > 0 && !*approved {
		fmt.Fprintln(os.Stderr, "tfplan: destructive changes require -approve-destructive")
		os.Exit(1)
	}
}
