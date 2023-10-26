package lib127

// SetRandFunc sets the random number generator for the given Hosts. Only
// exported for tests.
func SetRandFunc(h *Hosts, fn func(uint32) (uint32, error)) {
	h.randFunc = fn
}
