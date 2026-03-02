package handler

// CLIOptions holds global CLI flags parsed from command-line arguments.
type CLIOptions struct {
	JSON    bool // --json: output JSON envelope format
	NoColor bool // --no-color: disable ANSI color codes
	NoPager bool // --no-pager: disable interactive pagination
}
