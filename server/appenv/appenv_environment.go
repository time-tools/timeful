package appenv

// Handwritten sibling beside the generated appenv.go.
//
// Environment is transpiled from appenv.gala as `opaque type Environment
// string`, because GALA's `type X Y` is a type alias and would let any string
// pass as an Environment. Hash and Compare suppress the synthesized std-backed
// methods so the generated Go stays runtime-free.

// Hash hashes the environment name with the FNV-1a byte mixing the synthesized
// method used.
func (env Environment) Hash() uint32 {
	value := string(env)
	var h uint32 = 2166136261
	for index := 0; index < len(value); index++ {
		h = h ^ uint32(value[index])
		h = h * 16777619
	}
	return h
}

// Compare orders two environment names lexicographically, matching the
// synthesized method.
func (env Environment) Compare(other Environment) int {
	value := string(env)
	compare := string(other)
	if value < compare {
		return -1
	}
	if value > compare {
		return 1
	}
	return 0
}
