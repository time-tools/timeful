package discord_bot

import "timeful/server/discord_bot/commands"

// newCommandMap builds the package-level command map. GALA forbids a map
// literal or make, so the mixed-package sibling supplies the Go value; see
// docs/GO_INTEROP.MD Part 3 in the GALA repository.
func newCommandMap() map[string]commands.Command {
	return map[string]commands.Command{}
}

// argsFrom returns the arguments after the command name. GALA has no slice
// expressions, so the slicing stays in the sibling.
func argsFrom(args []string, index int) []string {
	return args[index:]
}
