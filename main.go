package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eannchen/leetsolv/command"
	"github.com/eannchen/leetsolv/config"
	"github.com/eannchen/leetsolv/core"
	"github.com/eannchen/leetsolv/handler"
	"github.com/eannchen/leetsolv/internal/clock"
	"github.com/eannchen/leetsolv/internal/errs"
	"github.com/eannchen/leetsolv/internal/fileutil"
	"github.com/eannchen/leetsolv/internal/logger"
	"github.com/eannchen/leetsolv/storage"
	"github.com/eannchen/leetsolv/usecase"
)

// Version information - will be set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Phase 1: Early --json detection (before any DI initialization)
	wantJSON := hasJSONFlag(os.Args)

	// Setup dependencies once
	clock := clock.NewClock()
	fileutil := fileutil.NewJSONFileUtil()
	cfg, err := config.NewConfig(fileutil)
	if err != nil {
		exitWithError(wantJSON, fmt.Errorf("Failed to load configuration: %w", err), 2)
	}
	if err := logger.Init(cfg.InfoLogFile, cfg.ErrorLogFile); err != nil {
		exitWithError(wantJSON, fmt.Errorf("Failed to initialize logger: %w", err), 2)
	}
	storage := storage.NewFileStorage(cfg.QuestionsFile, cfg.DeltasFile, fileutil)
	scheduler := core.NewSM2Scheduler(cfg, clock)
	questionUseCase := usecase.NewQuestionUseCase(cfg, storage, scheduler, clock)
	ioHandler := handler.NewIOHandler(clock)
	h := handler.NewHandler(cfg, questionUseCase, ioHandler, Version)

	commandRegistry := command.NewCommandRegistry(h.HandleUnknown)

	listCommand := &command.ListCommand{Handler: h}
	commandRegistry.Register("list", listCommand)
	commandRegistry.Register("ls", listCommand)

	searchCommand := &command.SearchCommand{Handler: h}
	commandRegistry.Register("search", searchCommand)
	commandRegistry.Register("s", searchCommand)

	getCommand := &command.GetCommand{Handler: h}
	commandRegistry.Register("detail", getCommand)
	commandRegistry.Register("get", getCommand)

	statusCommand := &command.StatusCommand{Handler: h}
	commandRegistry.Register("status", statusCommand)
	commandRegistry.Register("stat", statusCommand)

	upsertCommand := &command.UpsertCommand{Handler: h}
	commandRegistry.Register("upsert", upsertCommand)
	commandRegistry.Register("add", upsertCommand)

	deleteCommand := &command.DeleteCommand{Handler: h}
	commandRegistry.Register("remove", deleteCommand)
	commandRegistry.Register("rm", deleteCommand)
	commandRegistry.Register("delete", deleteCommand)
	commandRegistry.Register("del", deleteCommand)

	undoCommand := &command.UndoCommand{Handler: h}
	commandRegistry.Register("undo", undoCommand)
	commandRegistry.Register("back", undoCommand)

	historyCommand := &command.HistoryCommand{Handler: h}
	commandRegistry.Register("history", historyCommand)
	commandRegistry.Register("hist", historyCommand)
	commandRegistry.Register("log", historyCommand)

	settingCommand := &command.SettingCommand{Handler: h}
	commandRegistry.Register("setting", settingCommand)
	commandRegistry.Register("config", settingCommand)
	commandRegistry.Register("cfg", settingCommand)

	helpCommand := &command.HelpCommand{Handler: h}
	commandRegistry.Register("help", helpCommand)
	commandRegistry.Register("h", helpCommand)

	versionCommand := &command.VersionCommand{Handler: h}
	commandRegistry.Register("version", versionCommand)
	commandRegistry.Register("ver", versionCommand)
	commandRegistry.Register("v", versionCommand)

	migrateCommand := &command.MigrateCommand{Handler: h}
	commandRegistry.Register("migrate", migrateCommand)

	resetCommand := &command.ResetCommand{Handler: h}
	commandRegistry.Register("reset", resetCommand)

	clearCommand := &command.ClearCommand{Handler: h}
	commandRegistry.Register("clear", clearCommand)
	commandRegistry.Register("cls", clearCommand)

	quitCommand := &command.QuitCommand{Handler: h}
	commandRegistry.Register("quit", quitCommand)
	commandRegistry.Register("q", quitCommand)
	commandRegistry.Register("exit", quitCommand)

	scanner := bufio.NewScanner(os.Stdin)

	// --- CLI argument mode ---
	if len(os.Args) > 1 {
		// Phase 2: Full global flag parsing (supports flags before/after command)
		cmd, filtered, opt, parseErr := parseGlobalFlags(os.Args[1:])
		ioHandler.SetCLIOptions(opt)

		if parseErr != nil {
			exitWithError(wantJSON || opt.JSON, parseErr, 1)
		}
		if cmd == "" {
			exitWithError(opt.JSON, fmt.Errorf("no command specified. Run 'leetsolv help' for usage"), 1)
		}

		commandRegistry.Execute(scanner, cmd, filtered)

		// Exit with proper code based on last error
		os.Exit(mapExitCode(ioHandler.LastError()))
	}

	// --- Interactive mode ---

	// Set up graceful shutdown signal listener
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-signalChan
		fmt.Println("\nReceived shutdown signal. Please wait...")
		cancel() // Cancel the context

		// timeout
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

	h.HandleHelp()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down gracefully...")
			return
		default:
			fmt.Print(prompt())
			scanner.Scan()

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Parse command and arguments
			parts := strings.Fields(input)
			cmd := parts[0]
			args := parts[1:]

			// Execute command
			if quit := commandRegistry.Execute(scanner, cmd, args); quit {
				return
			}
		}
	}
}

func prompt() string {
	return "\nleetsolv ❯ "
}

// hasJSONFlag quickly checks if --json is present in command-line arguments.
// Used in Phase 1 before DI initialization to determine error output format.
func hasJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
	}
	return false
}

// parseGlobalFlags extracts global CLI flags and command name from args.
// Accepts os.Args[1:]. The first non-"--" prefixed arg is the command name.
// Unknown --flags before the command name are rejected as errors.
func parseGlobalFlags(args []string) (command string, filtered []string, opt handler.CLIOptions, err error) {
	commandFound := false
	for _, arg := range args {
		switch {
		case arg == "--json":
			opt.JSON = true
			opt.NoColor = true
			opt.NoPager = true
		case arg == "--no-color":
			opt.NoColor = true
		case arg == "--no-pager":
			opt.NoPager = true
		default:
			if !commandFound {
				if strings.HasPrefix(arg, "--") {
					err = fmt.Errorf("unknown global flag: %s. Place command-specific flags after the command name", arg)
					return
				}
				command = arg
				commandFound = true
			} else {
				filtered = append(filtered, arg)
			}
		}
	}
	return
}

// mapExitCode maps an error to the appropriate exit code.
// nil → 0, ValidationError/BusinessError → 1, SystemError → 2
func mapExitCode(err error) int {
	if err == nil {
		return 0
	}
	var codedErr *errs.CodedError
	if errors.As(err, &codedErr) {
		switch codedErr.Kind {
		case errs.SystemErrorKind:
			return 2
		default:
			return 1
		}
	}
	return 1
}

// exitWithError outputs an error message and exits with the given code.
// If wantJSON is true, outputs a JSON envelope; otherwise plain text.
func exitWithError(wantJSON bool, err error, code int) {
	if wantJSON {
		resp := struct {
			Success bool   `json:"success"`
			Data    any    `json:"data"`
			Error   string `json:"error"`
		}{
			Success: false,
			Data:    nil,
			Error:   err.Error(),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.Encode(resp)
	} else {
		fmt.Println(err.Error())
	}
	os.Exit(code)
}
