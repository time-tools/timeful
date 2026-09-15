package commands

import (
	"context"
	"fmt"

	"timeful/server/logger"
	pgstore "timeful/server/postgres"
)

var numUsers Command = Command{
	Name:        "/num_users",
	Description: "Returns the number of signed up users",
	Execute: func(args []string, webhookUrl string) {
		repository, err := pgstore.DefaultRepository()
		if err != nil {
			logger.StdErr.Panicln(err)
		}
		n, err := repository.CountAccounts(context.Background())
		if err != nil {
			logger.StdErr.Panicln(err)
		}

		response := Response{
			ResponseType: "in_channel",
			Text:         fmt.Sprintf("Number of currently signed up users: %v", n),
		}
		SendRawMessage(&response, webhookUrl)
	},
}
