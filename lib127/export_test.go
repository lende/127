package lib127

// WithRandFunc configure a copy of Hosts with the given random function. Only
// exported for tests.
func (h Hosts) WithRandFunc(fn func(uint32) (uint32, error)) *Hosts {
	h.randFunc = fn
	return &h
}
