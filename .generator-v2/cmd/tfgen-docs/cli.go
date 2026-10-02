package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// renderCLI walks the assembled tfgen command tree and renders one Markdown
// page describing every command and flag. It reads the live cobra tree rather
// than a hand-maintained list, which is the whole point: a flag that is added,
// renamed, or given a new default shows up here on the next `make tfgen-docs`,
// and the CI no-diff check fails until the page is regenerated.
//
// Only the flag *inventory* is derived. Behavioural prose — what --reconcile
// refuses to combine with, which exit code means what — is not inferable from
// the tree and lives in the hand-written preamble below, which is the one part
// of this page a human maintains.
// inertFlags names the flags that are registered — so a caller may pass them
// without error — but that the generator does not yet act on. This is the one
// hand-maintained fact on this page, because "accepted and ignored" is not
// something the cobra tree can express.
//
// Keep it honest: if you implement one of these, delete its entry. A stale
// entry here is the same class of bug this generated page exists to prevent,
// but it is a four-line surface rather than a prose page.
var inertFlags = map[string]string{
	"quiet":      "Accepted; informational logging is not yet suppressed.",
	"hooks-root": "Accepted; hook discovery is not yet implemented.",
	"strict":     "Accepted; `verify` performs no checks, so there is nothing to escalate.",
}

// inertCommands names commands that parse and exit zero without doing work.
var inertCommands = map[string]string{
	"verify": "Registered but not implemented: it performs no checks and returns success. " +
		"Do not treat a passing `tfgen verify` as evidence of anything.",
}

func renderCLI(root *cobra.Command) string {
	var b strings.Builder

	writeFrontMatter(&b, frontMatter{
		Title:     "CLI reference",
		Audience:  "anyone running `tfgen` directly, or reading a pipeline invocation.",
		Source:    "the cobra command tree in `internal/cli/`",
		Generator: "cmd/tfgen-docs",
	})

	b.WriteString(`tfgen is invoked through a single binary. Build it with ` + "`make tfgen-build`" + ` (which
writes ` + "`bin/tfgen`" + `) and run it from the provider checkout root.

Every flag below is read from the command tree at generation time, so this inventory
cannot drift from the flags actually registered. Flags marked **inert** are accepted
without error but not yet acted on — they are still part of the contract, because
passing one neither fails nor does anything.

`)

	b.WriteString("## Exit codes\n\n")
	b.WriteString("| Code | Meaning |\n|---|---|\n")
	b.WriteString("| `0` | The command completed and, in `--check` mode, found nothing to change. |\n")
	b.WriteString("| `1` | The command failed: parsing, generation, validation, I/O, or report writing. |\n")
	b.WriteString("| `3` | `generate --check` found generated files or wiring that would change. |\n\n")
	b.WriteString("Exit code `2` is not used. See [Interpreting a failed run](diagnostics.md).\n\n")

	cmds := []*cobra.Command{root}
	cmds = append(cmds, sortedSubcommands(root)...)

	for _, cmd := range cmds {
		renderCommand(&b, cmd, cmd == root)
	}

	return b.String()
}

func sortedSubcommands(root *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, c := range root.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func renderCommand(b *strings.Builder, cmd *cobra.Command, isRoot bool) {
	if isRoot {
		b.WriteString("## Global flags\n\n")
		b.WriteString("Accepted by every subcommand.\n\n")
		writeFlagTable(b, cmd.PersistentFlags())
		b.WriteString("`--version` prints the generator version; `--help` prints command help.\n\n")
		return
	}

	fmt.Fprintf(b, "## `tfgen %s`\n\n", cmd.Name())
	if cmd.Short != "" {
		fmt.Fprintf(b, "%s.\n\n", strings.TrimSuffix(cmd.Short, "."))
	}
	if note, ok := inertCommands[cmd.Name()]; ok {
		fmt.Fprintf(b, "> **Not implemented.** %s\n\n", note)
	}

	// Both persistent and local flags are part of the invocation surface; the
	// distinction only matters for subcommands of subcommands, which tfgen has
	// none of, so they are presented as one table.
	merged := pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
	merged.AddFlagSet(cmd.PersistentFlags())
	merged.AddFlagSet(cmd.Flags())

	if !merged.HasFlags() {
		b.WriteString("Takes no flags of its own.\n\n")
		return
	}
	writeFlagTable(b, merged)
}

func writeFlagTable(b *strings.Builder, fs *pflag.FlagSet) {
	b.WriteString("| Flag | Type | Default | Purpose |\n|---|---|---|---|\n")
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		name := "`--" + f.Name + "`"
		if f.Shorthand != "" {
			name = "`-" + f.Shorthand + "`, " + name
		}
		def := "`" + f.DefValue + "`"
		if f.DefValue == "" {
			def = "_empty_"
		}
		purpose := escapePipes(f.Usage)
		if note, ok := inertFlags[f.Name]; ok {
			purpose = "**Inert.** " + escapePipes(note)
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n",
			name, "`"+f.Value.Type()+"`", def, purpose)
	})
	b.WriteString("\n")
}

func escapePipes(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
