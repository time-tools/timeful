package commands

import pgstore "timeful/server/postgres"

// Handwritten sibling beside the generated active_users.go.
//
// Slice literals, make, append, and map literals are not first-class GALA
// constructs and their GALA substitutes name a runtime package, so this
// sibling supplies the Go values; see docs/GO_INTEROP.MD Part 3 in the GALA
// repository. The count helpers are here because `.Size()` on a slice whose
// type is declared in a handwritten sibling in another package is emitted
// unchanged and does not build. The chart response is here because its
// `[]map[string]any` blocks field is the same map-literal case.

func activeUsersDayStrings() []string {
	return []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
}

func activeUsersEmptyLabels() []string {
	return make([]string, 0)
}

func activeUsersEmptyData() []int {
	return make([]int, 0)
}

func activeUsersAppendLabel(labels []string, label string) []string {
	return append(labels, label)
}

func activeUsersAppendCount(data []int, count int) []int {
	return append(data, count)
}

func activeUsersLogsCount(logs []pgstore.DailyUserLog) int {
	return len(logs)
}

func activeUsersMembersCount(log pgstore.DailyUserLog) int {
	return len(log.Members)
}

func activeUsersChart(labels []string, data []int) map[string]any {
	return map[string]any{
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
}

func activeUsersChartResponse(chartUrl string) *Response {
	return &Response{ResponseType: "in_channel", Blocks: []map[string]any{
		{
			"type":      "image",
			"image_url": chartUrl,
			"alt_text":  "Active users chart",
		},
	}}
}
