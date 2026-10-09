package appenv

import (
	"reflect"
	"testing"
)

// TestEnvironmentIsDefinedType guards against Environment becoming a type
// alias of string, which lets any string pass where an Environment is expected.
func TestEnvironmentIsDefinedType(t *testing.T) {
	t.Parallel()

	environmentType := reflect.TypeOf(Development)
	if environmentType.Name() != "Environment" || environmentType.PkgPath() != "timeful/server/appenv" {
		t.Fatalf("Environment resolves to %s.%s, want the defined type timeful/server/appenv.Environment", environmentType.PkgPath(), environmentType.Name())
	}
}
