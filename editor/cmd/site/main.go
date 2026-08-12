package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/sitebuild"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/sitecheck"
	"github.com/Gonie-Gonie/Gonie-Gonie.github.io/editor/internal/toolcli"
)

const usage = `Usage:
  site build [--root PATH] [--validate-only]
  site validate [--root PATH]
  site check [--root PATH]
  site serve [--root PATH] [--port 8000]
  site bibliography

Build exports, validate website data and static references, or serve the repository locally.
Without --root, search upward from the current directory for the repository.
The bibliography command reads [{"settings":{...},"items":[...]}] from stdin
and writes the resolved bibliography fields as one nested JSON array.
Use -h or --help to show this help.
`

type commandOptions struct {
	name         string
	root         string
	validateOnly bool
	port         int
}

type commands struct {
	build   func(string, bool) (string, error)
	check   func(string) (string, error)
	resolve func(map[string]any, map[string]any) map[string]any
	serve   func(context.Context, string, int, io.Writer) error
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	exit := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, commands{
		build: sitebuild.Build, check: sitecheck.Check,
		resolve: sitebuild.ResolvePublication, serve: serve,
	})
	stop()
	os.Exit(exit)
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, actions commands) int {
	options, help, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "site: %v\n\n%s", err, usage)
		return 2
	}
	if help {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if options.name == "bibliography" {
		if err := bibliography(stdin, stdout, actions.resolve); err != nil {
			fmt.Fprintf(stderr, "site bibliography: %v\n", err)
			return 1
		}
		return 0
	}
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "site: read current directory: %v\n", err)
		return 1
	}
	root, err := toolcli.FindRoot(workingDir, options.root)
	if err != nil {
		fmt.Fprintf(stderr, "site: %v\n", err)
		return 1
	}
	var output string
	switch options.name {
	case "build":
		output, err = actions.build(root, options.validateOnly)
	case "validate", "check":
		output, err = actions.check(root)
	case "serve":
		err = actions.serve(ctx, root, options.port, stdout)
	}
	if output != "" {
		fmt.Fprint(stdout, output)
		if !strings.HasSuffix(output, "\n") {
			fmt.Fprintln(stdout)
		}
	}
	if err != nil {
		fmt.Fprintf(stderr, "site %s: %v\n", options.name, err)
		return 1
	}
	return 0
}

func parseArgs(args []string) (commandOptions, bool, error) {
	options := commandOptions{port: 8000}
	if len(args) == 0 {
		return options, false, errors.New("a command is required")
	}
	if args[0] == "-h" || args[0] == "--help" {
		return options, true, nil
	}
	options.name = args[0]
	switch options.name {
	case "build", "validate", "check", "serve", "bibliography":
	default:
		return options, false, fmt.Errorf("unknown command %q", options.name)
	}
	for index := 1; index < len(args); index++ {
		arg := args[index]
		if arg == "-h" || arg == "--help" {
			return options, true, nil
		}
		switch {
		case options.name != "bibliography" && (arg == "--root" || strings.HasPrefix(arg, "--root=")):
			if arg == "--root" {
				if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
					return options, false, errors.New("--root requires a repository path")
				}
				index++
				options.root = args[index]
			} else {
				options.root = strings.TrimPrefix(arg, "--root=")
			}
			if strings.TrimSpace(options.root) == "" {
				return options, false, errors.New("--root requires a repository path")
			}
		case options.name == "build" && arg == "--validate-only":
			options.validateOnly = true
		case options.name == "serve" && (arg == "--port" || strings.HasPrefix(arg, "--port=")):
			var value string
			if arg == "--port" {
				if index+1 >= len(args) {
					return options, false, errors.New("--port requires an integer from 1 to 65535")
				}
				index++
				value = args[index]
			} else {
				value = strings.TrimPrefix(arg, "--port=")
			}
			port, err := strconv.Atoi(value)
			if err != nil || port < 1 || port > 65535 {
				return options, false, fmt.Errorf("--port requires an integer from 1 to 65535: %q", value)
			}
			options.port = port
		default:
			return options, false, fmt.Errorf("unrecognized argument for %s: %q", options.name, arg)
		}
	}
	return options, false, nil
}

func bibliography(input io.Reader, output io.Writer, resolve func(map[string]any, map[string]any) map[string]any) error {
	var cases []struct {
		Settings map[string]any   `json:"settings"`
		Items    []map[string]any `json:"items"`
	}
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&cases); err != nil {
		return fmt.Errorf("read JSON array: %w", err)
	}
	if cases == nil {
		return errors.New("input must be a JSON array of {settings, items} objects")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("input must contain exactly one JSON array")
		}
		return fmt.Errorf("unexpected input after JSON array: %w", err)
	}
	for caseIndex, itemCase := range cases {
		if itemCase.Settings == nil || itemCase.Items == nil {
			return fmt.Errorf("case %d requires a settings object and an items array", caseIndex+1)
		}
		for itemIndex, item := range itemCase.Items {
			if item == nil {
				return fmt.Errorf("case %d item %d must be an object", caseIndex+1, itemIndex+1)
			}
		}
	}
	keys := []string{"date", "venue_en", "venue_kr", "reference_en", "reference_ko", "conference_info_en", "conference_info_ko", "indexing"}
	result := make([][]map[string]any, len(cases))
	for caseIndex, itemCase := range cases {
		result[caseIndex] = make([]map[string]any, len(itemCase.Items))
		for itemIndex, item := range itemCase.Items {
			resolved := resolve(item, itemCase.Settings)
			fields := make(map[string]any, len(keys))
			for _, key := range keys {
				if value, exists := resolved[key]; exists {
					fields[key] = value
				}
			}
			result[caseIndex][itemIndex] = fields
		}
	}
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}

func staticHandler(root string) (http.Handler, func() error, error) {
	files, err := os.OpenRoot(root)
	if err != nil {
		return nil, nil, err
	}
	// os.Root prevents path traversal and symlinks from exposing files outside root.
	return http.FileServerFS(files.FS()), files.Close, nil
}

func serve(ctx context.Context, root string, port int, output io.Writer) error {
	handler, closeFiles, err := staticHandler(root)
	if err != nil {
		return err
	}
	defer closeFiles()
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	finished := make(chan error, 1)
	go func() { finished <- server.Serve(listener) }()
	fmt.Fprintf(output, "Serving %s at http://%s/\n", root, listener.Addr())
	select {
	case err := <-finished:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			_ = server.Close()
			return err
		}
		if err := <-finished; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
