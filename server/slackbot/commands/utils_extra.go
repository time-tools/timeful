package commands

// Blocks documentation: https://api.slack.com/reference/block-kit/blocks
// Block builder: https://app.slack.com/block-kit-builder/

// Response stays handwritten because its struct tags define the Slack wire
// shape.
type Response struct {
	ResponseType string           `json:"response_type,omitempty"` // ephemeral or in_channel
	Text         string           `json:"text,omitempty"`
	Blocks       []map[string]any `json:"blocks,omitempty"`
}

// Command stays handwritten because its Execute field makes it a Go-facing
// value with a function type and a GALA-declared struct would widen its
// exported surface with synthesized Copy, Equal, and Unapply members.
type Command struct {
	Name        string
	Description string
	Usage       string
	Execute     func(args []string, webhookUrl string)
}

var CommandMap = map[string]Command{
	activeUsers.Name: activeUsers,
	numUsers.Name:    numUsers,
}

// newResponse exists for the .gala sources: the transpiler resolves a bare
// Response literal to postgres.Response, so the GALA source calls this helper
// instead of naming the local type.
func newResponse(responseType string, text string) *Response {
	return &Response{ResponseType: responseType, Text: text}
}
