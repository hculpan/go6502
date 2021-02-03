package utils

// ComputerStatus represents the status of
// the emulator.  This matches the screen.ComputerStatus
// interface so that the main method can communicate
// the status of the emulator to be displayed by
// the screen's UI
type ComputerStatus struct {
	Running     bool
	RomFilename string
	SingleStep  bool
}

// NewComputerStatus returns a new emulator status
func NewComputerStatus() *ComputerStatus {
	return &ComputerStatus{Running: false, RomFilename: "", SingleStep: false}
}

// IsRunning returns whether or not the emulator is executing code
func (e ComputerStatus) IsRunning() bool {
	return e.Running
}

// GetRomFileName gets the rom file name
func (e ComputerStatus) GetRomFileName() string {
	return e.RomFilename
}

// IsSingleStep returns whether or not the emulator is
// in single-step mode
func (e ComputerStatus) IsSingleStep() bool {
	return e.SingleStep
}

// Copy creates a new copy of the computer status
func (e ComputerStatus) Copy() *ComputerStatus {
	return &ComputerStatus{Running: e.Running, SingleStep: e.SingleStep, RomFilename: e.RomFilename}
}

// Equals returns true if and only if all the members of the provided
// ComputerStatus all match this object
func (e ComputerStatus) Equals(e2 ComputerStatus) bool {
	return e.Running == e2.Running && e.SingleStep == e2.SingleStep && e.RomFilename == e2.RomFilename
}
