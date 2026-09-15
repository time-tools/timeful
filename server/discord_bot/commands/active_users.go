package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"timeful/server/logger"
	pgstore "timeful/server/postgres"
	"timeful/server/utils"
)

var activeUsers Command = Command{
	Name: "!active_users",
	Description: `Gets the number of active users in the database, based on last sign in date. 
  - if LIST is true, it will list the name/email of all users, otherwise, it will show a bar graph
  - DAYS is the amount of days since last sign in
  `,
	Usage: "!active_users [LIST=false] [DAYS=7]",
	Execute: func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		var err error

		// Parse args
		list := false
		days := 7
		if len(args) >= 1 {
			if args[0] == "true" {
				list = true
			} else if args[0] == "false" {
				list = false
			} else {
				sendMessage(s, m, fmt.Sprintf("LIST=%s is not a valid boolean!", args[0]))
				return
			}
		}
		if len(args) >= 2 {
			days, err = strconv.Atoi(args[1])
			if err != nil {
				sendMessage(s, m, fmt.Sprintf("DAYS=%s is not a valid number!", args[1]))
				return
			}
		}

		// Read daily user logs starting from `days` days before the current
		// date from the authoritative PostgreSQL store. Empty days are padded by
		// the repository so the list and chart output are unchanged.
		startDate := time.Now().AddDate(0, 0, -days)
		startDate = utils.GetDateAtTime(startDate, "00:00:00")

		repository, err := pgstore.DefaultRepository()
		if err != nil {
			logger.StdErr.Panicln(err)
		}
		logs, err := repository.ListActiveUserDays(context.Background(), startDate, time.Now())
		if err != nil {
			logger.StdErr.Panicln(err)
		}

		// Define constants
		dayStrings := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

		if list {
			// Display a list of all active users
			sendMessage(s, m, "Active Users:\n")
			message := ""
			for _, log := range logs {
				date := log.LogDate
				message += dayStrings[date.Weekday()] + " "
				message += utils.GetDateString(date) + " | "
				message += fmt.Sprintf("Count: %d\n", len(log.Members))

				for _, user := range log.Members {
					message += fmt.Sprintf("\t- %s %s (%s)\n", user.FirstName, user.LastName, user.Email)
				}
			}

			for _, msg := range splitLongMessage(message, "```") {
				sendMessage(s, m, msg)
			}
		} else {
			// Display a bar graph of active users over time

			// Generate labels and data based on logs
			labels := make([]string, 0)
			data := make([]int, 0)
			for i := len(logs) - 1; i >= 0; i-- {
				labels = append(labels, utils.GetDateString(logs[i].LogDate))
				data = append(data, len(logs[i].Members))
			}

			// Generate chart using QuickChart API
			chart := map[string]any{
				"type": "bar",
				"data": map[string]any{
					"labels": labels,
					"datasets": []any{map[string]any{
						"label": "Active Users",
						"data":  data,
					}},
				},
				"options": map[string]any{
					"scales": map[string]any{
						"yAxes": []any{map[string]any{
							"ticks": map[string]any{
								"stepSize": 1,
							},
						}},
					},
				},
			}
			jsonStr, _ := json.Marshal(chart)

			encodedChart := url.PathEscape(string(jsonStr))
			chartUrl := fmt.Sprintf(`https://quickchart.io/chart?c=%s&backgroundColor=white`, encodedChart)
			chartEmbed := &discordgo.MessageEmbed{
				Title: "Active Users",
				Image: &discordgo.MessageEmbedImage{
					URL: chartUrl,
				},
			}

			sendEmbed(s, m, chartEmbed)
		}
	},
}
