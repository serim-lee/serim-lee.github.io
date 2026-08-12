package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/citations"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/toolcli"
)

const usage = `Usage: citations [--root PATH] [--kci-only] [--kci-count DOI COUNT]...

Refresh OpenAlex citation counts while preserving verified KCI counts.

  --root PATH            Repository root (otherwise search upward from the current directory).
  --kci-only             Skip OpenAlex and update only manually verified KCI counts.
  --kci-count DOI COUNT  Set an existing, published KCI article's verified count; repeat as needed.
  -h, --help             Show this help.
`

type updater func(context.Context, citations.Options) (citations.Result, error)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	exit := run(ctx, os.Args[1:], os.Stdout, os.Stderr, citations.Update)
	stop()
	os.Exit(exit)
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, update updater) int {
	options, help, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "citations: %v\n\n%s", err, usage)
		return 2
	}
	if help {
		fmt.Fprint(stdout, usage)
		return 0
	}
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "citations: read current directory: %v\n", err)
		return 1
	}
	options.Root, err = toolcli.FindRoot(workingDir, options.Root)
	if err != nil {
		fmt.Fprintf(stderr, "citations: %v\n", err)
		return 1
	}
	options.Log = func(line string) {
		fmt.Fprint(stdout, line)
	}
	result, err := update(ctx, options)
	if err != nil {
		fmt.Fprintf(stderr, "citations: %v\n", err)
		return 1
	}
	if !result.Success {
		return 1
	}
	return 0
}

func parseArgs(args []string) (citations.Options, bool, error) {
	var options citations.Options
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "-h" || arg == "--help":
			return options, true, nil
		case arg == "--kci-only":
			options.KCIOnly = true
		case arg == "--root" || strings.HasPrefix(arg, "--root="):
			if arg == "--root" {
				if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
					return options, false, fmt.Errorf("--root requires a repository path")
				}
				index++
				options.Root = args[index]
			} else {
				options.Root = strings.TrimPrefix(arg, "--root=")
			}
			if strings.TrimSpace(options.Root) == "" {
				return options, false, fmt.Errorf("--root requires a repository path")
			}
		case arg == "--kci-count":
			if index+2 >= len(args) {
				return options, false, fmt.Errorf("--kci-count requires DOI and COUNT")
			}
			doi, countText := args[index+1], args[index+2]
			if strings.TrimSpace(doi) == "" || strings.HasPrefix(doi, "-") {
				return options, false, fmt.Errorf("--kci-count requires DOI and COUNT")
			}
			count, err := strconv.ParseInt(countText, 10, 64)
			if err != nil || count < 0 || count > 9007199254740991 || !decimalDigits(countText) {
				return options, false, fmt.Errorf("--kci-count COUNT must be a nonnegative integer at most 9007199254740991: %q", countText)
			}
			options.KCICounts = append(options.KCICounts, citations.KCICount{DOI: doi, Count: count})
			index += 2
		case arg == "--" && index == len(args)-1:
			// No positional arguments are accepted.
		default:
			return options, false, fmt.Errorf("unrecognized argument %q", arg)
		}
	}
	return options, false, nil
}

func decimalDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}
