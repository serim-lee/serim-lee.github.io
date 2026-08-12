package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/release"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/toolcli"
)

func run(args []string) error {
	return runWithWriters(args, os.Stdout, os.Stderr)
}

func runWithWriters(args []string, output, errorOutput io.Writer) error {
	command := "check"
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}
	flags := flag.NewFlagSet("releases "+command, flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	requested := flags.String("root", "", "repository root")
	preserveData := flags.Bool("preserve-data", false, "skip checks against personal data files")
	if err := flags.Parse(args); err != nil {
		return err
	}
	start, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := toolcli.FindRoot(start, *requested)
	if err != nil {
		return err
	}
	switch command {
	case "check":
		if flags.NArg() != 0 {
			return fmt.Errorf("check accepts --root and --preserve-data")
		}
		manifest, err := release.CheckWithOptions(root, release.CheckOptions{PreserveData: *preserveData})
		if err != nil {
			return err
		}
		if *preserveData {
			fmt.Fprintf(output, "Release %s: release notes and update boundaries are valid; personal data checks skipped.\n", manifest.Version)
		} else {
			fmt.Fprintf(output, "Release %s: data versions, release notes and update boundaries are valid.\n", manifest.Version)
		}
	case "classify":
		policy, err := release.LoadPolicy(root)
		if err != nil {
			return err
		}
		if flags.NArg() == 0 {
			return fmt.Errorf("classify requires repository-relative paths")
		}
		for _, name := range flags.Args() {
			action, err := policy.Classify(name)
			if err != nil {
				return err
			}
			fmt.Fprintf(output, "%s\t%s\n", action, name)
		}
	default:
		return fmt.Errorf("usage: releases check [--root PATH] [--preserve-data] | releases classify [--root PATH] FILE...")
	}
	return nil
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
