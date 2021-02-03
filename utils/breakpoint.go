package utils

import "fmt"

// Breakpoint is an address that we want to
// drop into single-step when the emulator gets
// to this point
type Breakpoint struct {
	Address uint16
	Number  int

	count int
}

// Breakpoints is a global list of breakpoints
var Breakpoints []Breakpoint

// NewBreakpoint creates a new breakpoint object
// The NumberTimes parameters specifies how many times
// we should pass by this address before break; 0 indicates
// that it will stop everytime
func NewBreakpoint(address uint16, NumberTimes int) *Breakpoint {
	return &Breakpoint{
		Address: address,
		Number:  NumberTimes,
		count:   NumberTimes,
	}
}

// FindBreakpoint returns a breakpoint for the specified
// address, if one exists
func FindBreakpoint(addr uint16) (*Breakpoint, bool) {
	for _, v := range Breakpoints {
		if v.Address == addr {
			return &v, true
		}
	}

	return nil, false
}

func findBreakpointIndex(addr uint16) int {
	for i, v := range Breakpoints {
		if v.Address == addr {
			return i
		}
	}

	return -1
}

// AddBreakpoint adds a breakpoint to the list
func AddBreakpoint(b Breakpoint) {
	Breakpoints = append(Breakpoints, b)
}

// RemoveBreakpoint removes a breakpoint from the list
func RemoveBreakpoint(b Breakpoint) {
	idx := findBreakpointIndex(b.Address)
	if idx >= 0 {
		tmp := Breakpoints[:idx]
		Breakpoints = append(tmp, Breakpoints[idx+1:]...)
	}
}

// BreakpointReady checks if the breakpoint meets
// the filter criteria to triggle
// Everytime this is called, it counts as another
// iteration of this breakpoint being reached, so
// the internal counter will be increased by one
// when the result is false
// When the result is true, the internal counter
// will be reset
func (b *Breakpoint) BreakpointReady() bool {
	b.count--

	fmt.Printf("BreakpointReady called: %d\n", b.count)

	if b.count <= 0 {
		b.count = b.Number
		return true
	}

	return false
}
