package listmonk

// The Listmonk subscriber-search response shape. It stays handwritten because
// its json tags define the wire shape that Decode depends on.
type listmonkSubscriberResult struct {
	Id int `json:"id"`
}

type listmonkResults struct {
	Results []listmonkSubscriberResult `json:"results"`
}

type listmonkResponse struct {
	Data listmonkResults `json:"data"`
}
