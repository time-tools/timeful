package observability

// Config is the environment-scoped OpenObserve ingest contract established by
// the deployment environment files.
type Config struct {
	Endpoint       string
	OrganizationID string
	Username       string
	Password       string
	Environment    string
}
