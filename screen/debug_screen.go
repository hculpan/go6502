package screen

import (
	"fmt"

	"github.com/ariejan/i6502"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	textCols = 80
	textRows = 26
)

// EmulatorInterface provides a de-coupled interface
// to the emulator to avoid circular package dependencies
type EmulatorInterface interface {
	GetCPU() *i6502.Cpu
	ReadMemory(address uint16) uint8
}

// DebugScreen is the output for the debug info
type DebugScreen struct {
	parent   *Screen
	window   *sdl.Window
	renderer *sdl.Renderer
	font     *ttf.Font

	fontmetrics *ttf.GlyphMetrics
	charWidth   int32
	charHeight  int32

	screenWidth  int32
	screenHeight int32

	em EmulatorInterface
}

// NewDebugScreen creates a new debug window
func NewDebugScreen(parent *Screen) *DebugScreen {
	return &DebugScreen{parent: parent}
}

func (s *DebugScreen) initializeFonts() error {
	font, fontmetrics, err := loadFont("UbuntuMono-B.ttf")
	if err != nil {
		return fmt.Errorf("Error loading font UbuntuMono-B: %v", err)
	}
	s.font = font
	s.fontmetrics = fontmetrics

	s.charWidth, s.charHeight = getCharacterMetrics(s.fontmetrics)

	return nil
}

func (s *DebugScreen) createDebugWindow(parent *Screen) error {
	if err := s.initializeFonts(); err != nil {
		panic(err)
	}

	s.screenHeight = int32((s.fontmetrics.MaxY + s.fontmetrics.Advance) * textRows)
	s.screenWidth = int32((s.fontmetrics.MaxX) * textCols)

	x, y := s.parent.GetPosition()

	window, err := sdl.CreateWindow(
		"Debug Monitor",
		x-s.screenWidth,
		y,
		s.screenWidth,
		s.screenHeight,
		sdl.WINDOW_ALLOW_HIGHDPI,
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

	return nil
}

func (s *DebugScreen) displayDebugInfo() error {
	if s.window == nil {
		return nil
	}

	c := s.em.GetCPU()

	renderer, err := s.window.GetRenderer()
	if err != nil {
		return nil
	}

	instr := s.em.ReadMemory(c.PC)
	msg := fmt.Sprintf("Addr:%04X   A:%02X    X:%02X    Y:%02X    SP:%02X    FLAGS:%02X    Instr:%02X", c.PC, c.A, c.X, c.Y, c.SP, c.P, instr)
	texture, err := createTexture(msg, s.parent.foreground, s.font, renderer)
	if err != nil {
		return err
	} else if texture == nil {
		return nil
	}

	_, _, w, h, err := texture.Query()
	if err != nil {
		return fmt.Errorf("Unable to query texture: %v", err)
	}

	renderer.Copy(
		texture,
		&sdl.Rect{X: 0, Y: 0, W: w, H: h},
		&sdl.Rect{X: 20, Y: 20 + (h / 2), W: w, H: h},
	)

	return nil
}

// DrawScreen draws the entire screen
func (s *DebugScreen) DrawScreen() {
	s.renderer.SetDrawColor(s.parent.background.R, s.parent.background.G, s.parent.background.B, s.parent.background.A)
	s.renderer.Clear()

	if err := s.displayDebugInfo(); err != nil {
		panic(err)
	}

	s.renderer.Present()
}

// CleanUp cleans all resources created by the debug window
func (s *DebugScreen) CleanUp() {
	if s.window == nil {
		return
	}

	s.font.Close()
	if err := s.renderer.Destroy(); err != nil {
		panic(err)
	}
	if err := s.window.Destroy(); err != nil {
		panic(err)
	}
	s.window = nil
}

// Show shows the debug window
func (s *DebugScreen) Show(em EmulatorInterface) {
	if s.window == nil {
		s.em = em
		if err := s.createDebugWindow(s.parent); err != nil {
			panic(err)
		}
	}
}

// Hide hides the debug window
func (s *DebugScreen) Hide() {
	s.CleanUp()
}
