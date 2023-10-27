package lib127

// SetRandFunc sets the random number generator used by Hosts. Only exported for
// tests.
func (h *Hosts) SetRandFunc(fn func(uint32) (uint32, error)) {
	h.randFunc = fn
}
