package commands

import (
	"github.com/bwmarrin/discordgo"
)

// Command and the Commands registry stay handwritten: Execute is a
// Go-facing function field and Commands is read by the bot's init loop, so a
// GALA-declared struct or slice would widen their exported surface with
// synthesized members.
type Command struct {
	Name        string
	Description string
	Usage       string
	Execute     func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)
}

var Commands []Command = make([]Command, 0)

func Init() {
	Commands = append(Commands, activeUsers)
	Commands = append(Commands, numUsers)

	Commands = append(Commands, help)
}
