// Package providerconfig resolves the outbound OAuth and calendar provider
// endpoints used by the server.
//
// The real provider endpoints are always the default. Each endpoint can be
// redirected through a TEST_-prefixed environment variable so the isolated test
// stack can reach a mock provider instead of the public internet. The overrides
// are intentionally named after the test stack: production and staging must
// never set them, and leaving them unset keeps the real provider URLs.
package providerconfig
