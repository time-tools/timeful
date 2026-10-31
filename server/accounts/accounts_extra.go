package accounts

// Profile carries the profile fields a sign-in provider or OTP flow supplies.
//
// The struct stays handwritten rather than moving into accounts.gala: a
// GALA-declared struct emits Copy, Equal, Unapply, NewProfile, and
// StructMeta_Profile, which widens this package's exported surface even though
// a Go caller still constructs it by field name.
type Profile struct {
	Email          string
	FirstName      string
	LastName       string
	Picture        string
	TimezoneOffset int
}
