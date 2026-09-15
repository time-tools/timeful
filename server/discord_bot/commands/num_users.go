package commands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"timeful/server/logger"
	pgstore "timeful/server/postgres"
)

var numUsers Command = Command{
	Name:        "!num_users",
	Description: "Returns the number of signed up users",
	Execute: func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		repository, err := pgstore.DefaultRepository()
		if err != nil {
			logger.StdErr.Panicln(err)
		}
		n, err := repository.CountAccounts(context.Background())
		if err != nil {
			logger.StdErr.Panicln(err)
		}

		sendMessage(s, m, fmt.Sprintf("Number of currently signed up users: %v", n))
	},
}
