package screen

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hculpan/go6502/keyboard"
	"github.com/hculpan/go6502/resources"
	"github.com/hculpan/go6502/utils"
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
	computerStatus     *utils.ComputerStatus
	prevComputerStatus *utils.ComputerStatus

	cursor *CursorPos

	textCols   int
	textRows   int
	background sdl.Color
	foreground sdl.Color

	window       *sdl.Window
	renderer     *sdl.Renderer
	screenWidth  int32
	screenHeight int32

	font        *ttf.Font
	fontmetrics *ttf.GlyphMetrics
	charHeight  int32
	charWidth   int32

	Busy bool

	escapeMode     bool
	escapeSequence string
	escapeCodes    []escapeCode

	debugScreen *DebugScreen

	//	done                chan bool
	cursorNextSequence  bool
	cursorCurrentSymbol rune

	symbols [128]*sdl.Texture

	capsLock bool
	shiftOn  bool

	escapeTexture      *sdl.Texture
	romFileTexture     *sdl.Texture
	keyOptionsTexture  *sdl.Texture
	emulatorOnTexture  *sdl.Texture
	emulatorOffTexture *sdl.Texture

	emulatorOnOffRect *sdl.Rect

	videoRAM []rune

	screenDirty bool
}

// NewScreen creates a new screen object
func NewScreen(cols int, rows int, status *utils.ComputerStatus) *Screen {
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
	s.Busy = false
	s.shiftOn = false
	s.computerStatus = status
	s.prevComputerStatus = nil
	s.screenDirty = true
	return s
}

// GetPosition returns the position of the screen
// returns -1, -1 if window not created
func (s Screen) GetPosition() (x, y int32) {
	if s.window != nil {
		return s.window.GetPosition()
	}

	return -1, -1
}

// GetSize returns the size of the screen
// returns -1,-1 if window not create
func (s Screen) GetSize() (w, h int32) {
	if s.window != nil {
		return s.window.GetSize()
	}

	return -1, -1
}

// IsBusy returns if the screen is busy
func (s Screen) IsBusy() bool {
	return s.Busy
}

// UpdateScreen allows an external process
// to force an update of the screen
func (s *Screen) UpdateScreen() {
	s.screenDirty = true
}

// ProcessRune processes an ASCII character
func (s *Screen) ProcessRune(r rune) {
	s.screenDirty = true
	s.Busy = true
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
	s.Busy = false
}

// EnableDebug turns on debugging/single step
func (s *Screen) EnableDebug(em EmulatorInterface) {
	s.computerStatus.SingleStep = true
	s.debugScreen.Show(em, s.computerStatus)
}

// DisableDebug turns off debugging/single step
func (s *Screen) DisableDebug() {
	s.computerStatus.SingleStep = false
	s.debugScreen.Hide()
}

func (s *Screen) computerStatusChanged() bool {
	if s.prevComputerStatus == nil {
		s.prevComputerStatus = s.computerStatus.Copy()
		return false
	}

	result := s.computerStatus.Running != s.prevComputerStatus.Running ||
		s.computerStatus.SingleStep != s.prevComputerStatus.SingleStep ||
		s.computerStatus.RomFilename != s.computerStatus.RomFilename

	if result {
		s.prevComputerStatus = s.computerStatus.Copy()
	}

	return result
}

// CleanUp cleans up all the resources creates by the screen object
func (s *Screen) CleanUp() {
	if s.debugScreen != nil {
		s.debugScreen.CleanUp()
	}

	for x := 0; x < 128; x++ {
		if s.symbols[x] != nil {
			s.symbols[x].Destroy()
		}
	}
	if s.escapeTexture != nil {
		s.escapeTexture.Destroy()
	}
	if s.romFileTexture != nil {
		s.romFileTexture.Destroy()
	}
	if s.keyOptionsTexture != nil {
		s.keyOptionsTexture.Destroy()
	}
	if s.emulatorOffTexture != nil {
		s.emulatorOffTexture.Destroy()
	}
	if s.emulatorOnTexture != nil {
		s.emulatorOnTexture.Destroy()
	}

	s.font.Close()
	s.renderer.Destroy()
	s.window.Destroy()
}

// GetWindowPosition returns the current window position
func (s *Screen) GetWindowPosition() (x, y int32) {
	return s.window.GetPosition()
}

func (s *Screen) initializeSymbols() error {
	for x := 32; x < 127; x++ {
		t, err := utils.CreateTexture(string(rune(x)), s.foreground, s.font, s.renderer)
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

// Reset resets the screen to startup state
func (s *Screen) Reset() {
	s.screenDirty = true
	s.cursor.X = 0
	s.cursor.Y = 0
	s.cursor.ClearScroll()
	for i := range s.videoRAM {
		s.videoRAM[i] = 0
	}
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
	font, fontmetrics, err := utils.LoadFont("OldComputerManualMonospaced-KmlZ.ttf")
	if err != nil {
		return err
	}
	s.font = font
	s.fontmetrics = fontmetrics

	s.charWidth, s.charHeight = utils.GetCharacterMetrics(s.fontmetrics)

	return nil
}

// Show initializes the main screen and shows it
func (s *Screen) Show() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	if err := ttf.Init(); err != nil {
		panic(err)
	}

	if err := s.initializeFontStuff(); err != nil {
		panic(err)
	}

	s.screenHeight = int32((s.fontmetrics.MaxY + s.fontmetrics.Advance) * s.textRows)
	s.screenWidth = int32((s.fontmetrics.MaxX) * s.textCols)

	window, err := sdl.CreateWindow(
		"Kabputer",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		s.screenWidth,
		s.screenHeight+90,
		sdl.WINDOW_ALLOW_HIGHDPI,
	)
	if err != nil {
		panic(err)
	}
	s.window = window

	renderer, err := sdl.CreateRenderer(s.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	s.renderer = renderer

	s.initializeSymbols()

	s.initEscapeCodes()

	s.debugScreen = NewDebugScreen(s)

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
	if s.screenDirty {
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

		if err := s.drawUI(); err != nil {
			panic(err)
		}

		s.renderer.Present()
	}

	if s.computerStatus.SingleStep {
		s.debugScreen.DrawScreen()
	}

	s.screenDirty = false
}

func (s *Screen) createBarTexture(msg string) (*sdl.Texture, error) {
	font, _, err := utils.LoadFont("Aileron-Bold.otf")
	if err != nil {
		return nil, err
	}

	texture, err := utils.CreateTexture(msg, s.foreground, font, s.renderer)
	if err != nil {
		return nil, err
	}

	return texture, nil
}

func (s *Screen) createPowerTextures() error {
	data, err := resources.Asset("resources/computer_off.bmp")
	if err != nil {
		return err
	}

	rwops, err := sdl.RWFromMem(data)
	if err != nil {
		return err
	}

	surface, err := sdl.LoadBMPRW(rwops, true)
	if err != nil {
		return err
	}

	texture, err := s.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return err
	}
	s.emulatorOffTexture = texture
	surface.Free()

	data, err = resources.Asset("resources/computer_on.bmp")
	if err != nil {
		return err
	}

	rwops, err = sdl.RWFromMem(data)
	if err != nil {
		return err
	}

	surface, err = sdl.LoadBMPRW(rwops, true)
	if err != nil {
		return err
	}

	texture, err = s.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return err
	}
	s.emulatorOnTexture = texture
	surface.Free()

	_, _, w, h, err := s.emulatorOffTexture.Query()
	if err != nil {
		return err
	}
	s.emulatorOnOffRect = &sdl.Rect{X: 10, Y: s.screenHeight + (45 - (h / 10)), W: w / 5, H: h / 5}

	return nil
}

func (s *Screen) generateStatusTextures() error {
	var msg string
	switch {
	case s.computerStatus.Running && !s.computerStatus.SingleStep:
		msg = "F2: Off       F5: Pause       F9: Reload/Reset"
	case s.computerStatus.Running && s.computerStatus.SingleStep:
		msg = "F2: Off        F6: Single step       F7: Resume       F9: Reload/Reset"
	default:
		msg = "F2: On        F3: On/Single step        F9: Reload/Reset"
	}

	texture, err := s.createBarTexture(msg)
	if err != nil {
		return err
	}
	s.keyOptionsTexture = texture

	texture, err = s.createBarTexture(fmt.Sprintf("ROM: %s", filepath.Base(s.computerStatus.RomFilename)))
	if err != nil {
		return err
	}
	s.romFileTexture = texture

	return nil
}

func (s *Screen) drawUI() error {
	changed := s.prevComputerStatus == nil || !s.computerStatus.Equals(*s.prevComputerStatus)
	r, g, b, a, _ := s.renderer.GetDrawColor()

	s.renderer.SetDrawColor(128, 128, 128, 128)
	s.renderer.DrawLine(0, s.screenHeight+1, s.screenWidth, s.screenHeight+1)

	if changed {
		s.generateStatusTextures()
	}

	if s.emulatorOffTexture == nil {
		if err := s.createPowerTextures(); err != nil {
			return err
		}
	}

	if s.escapeTexture == nil {
		texture, err := s.createBarTexture("ESC: Exit")
		if err != nil {
			return err
		}
		s.escapeTexture = texture
	}

	if s.computerStatus.Running {
		_, _, w, h, err := s.emulatorOffTexture.Query()
		if err != nil {
			return err
		}

		s.renderer.Copy(
			s.emulatorOnTexture,
			&sdl.Rect{X: 0, Y: 0, W: w, H: h},
			s.emulatorOnOffRect,
		)
	} else {
		_, _, w, h, err := s.emulatorOffTexture.Query()
		if err != nil {
			return err
		}

		s.renderer.Copy(
			s.emulatorOffTexture,
			&sdl.Rect{X: 0, Y: 0, W: w, H: h},
			s.emulatorOnOffRect,
		)
	}

	if s.romFileTexture != nil {
		_, _, w, h, err := s.romFileTexture.Query()
		if err != nil {
			return err
		}

		s.renderer.Copy(
			s.romFileTexture,
			&sdl.Rect{X: 0, Y: 0, W: w, H: h},
			&sdl.Rect{X: s.screenWidth - (w + 20), Y: s.screenHeight + (20 - (h / 2)), W: w, H: h},
		)
	}

	if s.escapeTexture != nil {
		_, _, w, h, err := s.escapeTexture.Query()
		if err != nil {
			return err
		}

		s.renderer.Copy(
			s.escapeTexture,
			&sdl.Rect{X: 0, Y: 0, W: w, H: h},
			&sdl.Rect{X: s.screenWidth - (w + 20), Y: s.screenHeight + (20 - (h / 2) + 40), W: w, H: h},
		)
	}

	_, _, w, h, err := s.keyOptionsTexture.Query()
	if err != nil {
		return err
	}

	s.renderer.Copy(
		s.keyOptionsTexture,
		&sdl.Rect{X: 0, Y: 0, W: w, H: h},
		&sdl.Rect{X: (s.screenWidth / 2) - (w / 2) - 100, Y: s.screenHeight + (20 - (h / 2)), W: w, H: h},
	)

	s.renderer.SetDrawColor(r, g, b, a)

	return nil
}

// GetFontWidth returns the width of an individual character
func (s *Screen) GetFontWidth() int32 {
	return s.charWidth
}

// GetFontHeight return the height of a row
func (s *Screen) GetFontHeight() int32 {
	return s.charHeight
}

func (s *Screen) scrollDisplayUp() {
	s.screenDirty = true
	for i := range s.videoRAM[:len(s.videoRAM)-s.textCols] {
		s.videoRAM[i] = s.videoRAM[i+s.textCols]
	}
	for i := 0; i < s.textCols; i++ {
		idx := i + ((s.textRows - 1) * s.textCols)
		s.videoRAM[idx] = 0
	}
}

func (s *Screen) scrollDisplayDown() {
	s.screenDirty = true
	for i := len(s.videoRAM) - s.textCols - 1; i >= 0; i-- {
		s.videoRAM[i+s.textCols] = s.videoRAM[i]
	}
	for i := 0; i < s.textCols; i++ {
		s.videoRAM[i] = 0
	}
	s.cursor.ClearScroll()
}

func (s *Screen) displayCursor() {
	if !s.computerStatus.Running {
		return
	}

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

// IsEmulatorOnOffClicked returns if x,y is within the rectangle
// of the graphic on/off switch
func (s *Screen) IsEmulatorOnOffClicked(x int32, y int32) bool {
	if s.emulatorOnOffRect == nil {
		return false
	}

	return x > s.emulatorOnOffRect.X &&
		x < s.emulatorOnOffRect.X+s.emulatorOnOffRect.W &&
		y > s.emulatorOnOffRect.Y &&
		y < s.emulatorOnOffRect.Y+s.emulatorOnOffRect.H
}
