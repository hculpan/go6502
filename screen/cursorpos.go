package screen

// CursorPos represents the position of the cursor on the screen
// It also manages the extents of the screen, so that the cursor
// Cannot go outside the screen.
//
// The Scroll flag is there to indicate to the calling object that
// it should scroll up.  This object will not handle the scrolling,
// and the flag can be cleared by either moving the cursor or by
// calling the ClearScroll method.
type CursorPos struct {
	X        int
	Y        int
	Display  bool
	Scroll   bool
	Sequence []rune

	currSequence int
	textCols     int
	textRows     int
}

// NewCursorPos creates a new cursor pos
func NewCursorPos(cols int, rows int, sequence []rune) *CursorPos {
	return &CursorPos{
		X:            0,
		Y:            0,
		Display:      true,
		Scroll:       false,
		textCols:     cols,
		textRows:     rows,
		Sequence:     sequence,
		currSequence: -1,
	}
}

// NextSequence returns the next rune in the cursor sequence
func (cpos *CursorPos) NextSequence() rune {
	cpos.currSequence++
	if cpos.currSequence == len(cpos.Sequence) {
		cpos.currSequence = 0
	}
	return cpos.Sequence[cpos.currSequence]
}

// ClearScroll clears the scroll flag
func (cpos *CursorPos) ClearScroll() {
	cpos.Scroll = false
}

// NextLocation moves the cursor right by one
// It will move to the next line if you're at the end
// of a line
func (cpos *CursorPos) NextLocation() {
	cpos.ClearScroll()
	cpos.X++
	if cpos.X >= cpos.textCols {
		cpos.NewLine()
	}
}

// NextLine moves the cursor to the next line, but at the same
// col.  If it exceeds the number of rows on the screen, will
// set the Scroll flag.
func (cpos *CursorPos) NextLine() {
	cpos.ClearScroll()
	cpos.Y++
	if cpos.Y >= cpos.textRows {
		cpos.Y = cpos.textRows - 1
		cpos.Scroll = true
	}
}

// NewLine moves the cursor to the beginning of the next line.
// If it exceeds the number of rows on the screen, it will
// set the Scroll flag.
func (cpos *CursorPos) NewLine() {
	cpos.ClearScroll()
	cpos.X = 0
	cpos.Y++
	if cpos.Y >= cpos.textRows {
		cpos.Y = cpos.textRows - 1
		cpos.Scroll = true
	}
}

// Backspace moves the cursor back one space.  If it
// reaches the beginning of the line, it will do nothing.
func (cpos *CursorPos) Backspace() {
	cpos.ClearScroll()
	if cpos.X > 0 {
		cpos.X--
	}
}
