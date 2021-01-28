package screen

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hculpan/go6502/keyboard"
	"github.com/hculpan/go6502/resources"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const escapeMaxLength = 10

var numMatcher *regexp.Regexp = regexp.MustCompile("[0-9]+")

type escapeCode struct {
	Code    string
	Matcher *regexp.Regexp
	Action  func(s *Screen, m *regexp.Regexp, escapeSequence string)
}

// Screen represents the main object to display text output
type Screen struct {
	cursor *CursorPos

	textCols    int
	textRows    int
	fontmetrics *ttf.GlyphMetrics
	background  sdl.Color
	foreground  sdl.Color

	window   *sdl.Window
	renderer *sdl.Renderer
	font     *ttf.Font

	charHeight int32
	charWidth  int32

	escapeMode     bool
	escapeSequence string
	escapeCodes    []escapeCode

	done                chan bool
	cursorNextSequence  bool
	cursorCurrentSymbol rune

	symbols [128]*sdl.Texture

	capsLock bool
	shiftOn  bool

	videoRAM []rune
}

// NewScreen creates a new screen object
func NewScreen(cols int, rows int) *Screen {
	s := &Screen{textCols: cols, textRows: rows}
	s.cursor = NewCursorPos(cols, rows, []rune{95, 0})
	s.cursorNextSequence = true
	s.background = sdl.Color{R: 0, G: 0, B: 0, A: 0}
	s.foreground = sdl.Color{R: 255, G: 255, B: 255, A: 255}
	for x := 0; x < 128; x++ {
		s.symbols[x] = nil
	}
	s.videoRAM = make([]rune, cols*rows)
	for i := range s.videoRAM {
		s.videoRAM[i] = 0
	}
	s.capsLock = false
	s.shiftOn = false
	return s
}

// CleanUp cleans up all the resources creates by the screen object
func (s *Screen) CleanUp() {
	s.done <- true
	for x := 0; x < 128; x++ {
		if s.symbols[x] != nil {
			s.symbols[x].Destroy()
		}
	}
	s.font.Close()
	s.renderer.Destroy()
	s.window.Destroy()
}

func (s *Screen) initSymbolTexture(r rune) (*sdl.Texture, error) {
	var msg string = string(r)
	surface, err := s.font.RenderUTF8Solid(msg, s.foreground)
	if err != nil {
		return nil, err
	}
	defer surface.Free()

	msgtext, err := s.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, err
	}
	return msgtext, nil
}

func (s *Screen) initializeSymbols() error {
	for x := 32; x < 127; x++ {
		t, err := s.initSymbolTexture(rune(x))
		if err != nil {
			return err
		}
		s.symbols[x] = t
	}

	return nil
}

func findFirstNumber(s string, def int) int {
	numStr := numMatcher.FindString(s)
	result, err := strconv.Atoi(numStr)
	if err != nil {
		result = def
	}
	return result
}

func findTwoNumbers(s string, def int) (int, int) {
	strs := strings.Split(s, ";")
	if len(strs) != 2 {
		return def, def
	}
	n1 := findFirstNumber(strs[0], def)
	n2 := findFirstNumber(strs[1], def)
	return n1, n2
}

func (s *Screen) initEscapeCodes() {
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[D",
			Matcher: regexp.MustCompile(`\[D`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				s.scrollDisplayUp()
				s.cursor.ClearScroll()
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[M",
			Matcher: regexp.MustCompile(`\[M`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				s.scrollDisplayDown()
				s.cursor.ClearScroll()
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[H",
			Matcher: regexp.MustCompile(`\[H`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				s.cursor.X = 0
				s.cursor.Y = 0
				s.cursor.ClearScroll()
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[H",
			Matcher: regexp.MustCompile(`\[2J`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				for i := range s.videoRAM {
					s.videoRAM[i] = 0
				}
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[<n>A",
			Matcher: regexp.MustCompile(`\[[0-9]+A`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				n := findFirstNumber(s.escapeSequence, 0)
				for i := 0; i < n; i++ {
					s.cursor.Y--
				}
				if s.cursor.Y < 0 {
					s.cursor.Y = 0
				}
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[<n>B",
			Matcher: regexp.MustCompile(`\[[0-9]+B`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				n := findFirstNumber(s.escapeSequence, 0)
				for i := 0; i < n; i++ {
					s.cursor.Y++
				}
				if s.cursor.Y >= s.textRows {
					s.cursor.Y = s.textRows - 1
				}
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[<n>C",
			Matcher: regexp.MustCompile(`\[[0-9]+C`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				n := findFirstNumber(s.escapeSequence, 0)
				for i := 0; i < n; i++ {
					s.cursor.X++
				}
				if s.cursor.X >= s.textCols {
					s.cursor.X = s.textCols - 1
				}
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[<n>D",
			Matcher: regexp.MustCompile(`\[[0-9]+D`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				n := findFirstNumber(s.escapeSequence, 0)
				for i := 0; i < n; i++ {
					s.cursor.X--
				}
				if s.cursor.X < 0 {
					s.cursor.X = 0
				}
			},
		},
	)
	s.escapeCodes = append(s.escapeCodes,
		escapeCode{
			Code:    "[<n>;<n>H",
			Matcher: regexp.MustCompile(`\[[0-9]+;[0-9]+H`),
			Action: func(s *Screen, m *regexp.Regexp, escapeSequence string) {
				n1, n2 := findTwoNumbers(escapeSequence, -1)
				if n1 >= 0 && n1 < s.textCols && n2 >= 0 && n2 < s.textRows {
					s.cursor.X = n1
					s.cursor.Y = n2
				}
			},
		},
	)
}

func (s *Screen) initializeFontStuff() error {
	// Load data from bindata in resources/resources.go
	data, err := resources.Asset("resources/OldComputerManualMonospaced-KmlZ.ttf")
	if err != nil {
		return err
	}

	rwops, err := sdl.RWFromMem(data)
	if err != nil {
		return err
	}

	font, err := ttf.OpenFontRW(rwops, 1, 24)
	if err != nil {
		return err
	}
	s.font = font

	s.fontmetrics, _ = font.GlyphMetrics('A')
	s.charWidth = int32(s.fontmetrics.MaxX)
	s.charHeight = int32(s.fontmetrics.MaxY + s.fontmetrics.Advance)

	return nil
}

// Show initializes the main screen and shows it
func (s *Screen) Show() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}

	if err := ttf.Init(); err != nil {
		return err
	}

	if err := s.initializeFontStuff(); err != nil {
		return err
	}

	window, err := sdl.CreateWindow(
		"Kabputer",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32((s.fontmetrics.MaxX)*s.textCols),
		int32((s.fontmetrics.MaxY+s.fontmetrics.Advance)*s.textRows),
		sdl.WINDOW_OPENGL,
	)
	if err != nil {
		return err
	}
	s.window = window

	renderer, err := sdl.CreateRenderer(s.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}
	s.renderer = renderer

	s.initializeSymbols()

	s.initEscapeCodes()

	ticker := time.NewTicker(50 * time.Millisecond)
	s.done = make(chan bool)

	go func() {
		cnt := 0
		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				if cnt%6 == 0 {
					s.cursorNextSequence = true
				}
				s.DrawScreen()
				cnt++
			}
		}
	}()

	return nil
}

func (s *Screen) matchesEscapeCode(r rune) {
	s.escapeSequence += string(r)
	for _, v := range s.escapeCodes {
		if v.Matcher.MatchString(s.escapeSequence) {
			v.Action(s, v.Matcher, s.escapeSequence)
			s.escapeMode = false
		}
	}

	if s.escapeMode && len(s.escapeSequence) > escapeMaxLength {
		s.escapeMode = false
	}

}

func (s *Screen) calculateScreenLocationFromIndex(i int) (int, int) {
	y := i / s.textCols
	x := i - y*s.textCols

	return x, y
}

func (s *Screen) calculateIndexFromScreenLocation(x, y int) int {
	return (y * s.textCols) + x
}

// DrawScreen draws the entire screen
func (s *Screen) DrawScreen() {
	s.renderer.SetDrawColor(s.background.R, s.background.G, s.background.B, s.background.A)
	s.renderer.Clear()

	if s.cursor.Scroll {
		s.scrollDisplayUp()
		s.cursor.ClearScroll()
	}

	for i, v := range s.videoRAM {
		if v != 0 {
			x, y := s.calculateScreenLocationFromIndex(i)
			s.displayRuneAt(v, x, y)
		}
	}

	s.displayCursor()

	s.renderer.Present()
}

// GetFontWidth returns the width of an individual character
func (s *Screen) GetFontWidth() int32 {
	return s.charWidth
}

// GetFontHeight return the height of a row
func (s *Screen) GetFontHeight() int32 {
	return s.charHeight
}

// ProcessRune processes an ASCII character
func (s *Screen) ProcessRune(r rune) {
	if s.escapeMode {
		s.matchesEscapeCode(r)
	} else {
		switch {
		case r == keyboard.Backspace: // backspace
			s.cursor.Backspace()
			s.videoRAM[s.calculateIndexFromScreenLocation(s.cursor.X, s.cursor.Y)] = 0
		case r == keyboard.Escape:
			s.escapeMode = true
			s.escapeSequence = ""
		case r == keyboard.Enter: // enter
			s.cursor.NewLine()
		case r < 32 || r > 126:
			// Nothing
		default:
			s.displayRune(r)
		}
	}
}

func (s *Screen) scrollDisplayUp() {
	for i := range s.videoRAM[:len(s.videoRAM)-s.textCols] {
		s.videoRAM[i] = s.videoRAM[i+s.textCols]
	}
	for i := 0; i < s.textCols; i++ {
		idx := i + ((s.textRows - 1) * s.textCols)
		s.videoRAM[idx] = 0
	}
}

func (s *Screen) scrollDisplayDown() {
	for i := len(s.videoRAM) - s.textCols - 1; i >= 0; i-- {
		s.videoRAM[i+s.textCols] = s.videoRAM[i]
	}
	for i := 0; i < s.textCols; i++ {
		s.videoRAM[i] = 0
	}
	s.cursor.ClearScroll()
}

func (s *Screen) displayCursor() {
	// Now render new text
	if s.cursorNextSequence {
		s.cursorCurrentSymbol = s.cursor.NextSequence()
		s.cursorNextSequence = false
	}
	s.displayRuneAt(s.cursorCurrentSymbol, s.cursor.X, s.cursor.Y)
}

func (s *Screen) displayRuneAt(r rune, x int, y int) {
	rect := &sdl.Rect{
		X: int32(x) * s.GetFontWidth(),
		Y: int32(y) * s.GetFontHeight(),
		W: s.GetFontWidth(),
		H: s.GetFontHeight(),
	}

	// Now render new text
	s.renderer.Copy(s.symbols[r], &sdl.Rect{X: 0, Y: 0, W: s.GetFontWidth(), H: s.GetFontHeight()}, rect)
}

func (s *Screen) displayRune(r rune) error {
	i := s.calculateIndexFromScreenLocation(s.cursor.X, s.cursor.Y)
	s.videoRAM[i] = r
	s.cursor.NextLocation()
	return nil
}
