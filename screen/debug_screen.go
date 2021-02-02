package screen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

type codeLine struct {
	address uint16
	line    string
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

	lastPC               uint16
	lastDebugTexture     *sdl.Texture
	debugHeaderTexure    *sdl.Texture
	lastDebugCodeTexture *sdl.Texture
	lastStackTexture     *sdl.Texture

	status    *ComputerStatus
	debugCode []codeLine

	Active bool
}

// NewDebugScreen creates a new debug window
func NewDebugScreen(parent *Screen) *DebugScreen {
	result := &DebugScreen{parent: parent, Active: false}
	result.createDebugWindow()
	result.lastPC = 0
	result.lastDebugTexture = nil
	result.lastDebugCodeTexture = nil
	return result
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

func (s *DebugScreen) createDebugWindow() error {
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
		sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_HIDDEN,
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

	// This is a kludge.  I force it to load all the printable ASCII
	// characters in the font, and this seems to cause it to print
	// on the debug screen correctly.  Without this, I get all sorts
	// weird issues
	s.initializeSymbols()

	return nil
}

func (s *DebugScreen) createDebugHeaderTexture(renderer *sdl.Renderer, c *i6502.Cpu) (*sdl.Texture, error) {
	if s.debugHeaderTexure != nil {
		s.debugHeaderTexure.Destroy()
	}
	msg := fmt.Sprintf("           A       X       Y      FLAGS:NVxxDIZC      SP")
	texture, err := createTexture(msg, s.parent.foreground, s.font, renderer)
	if err != nil {
		return nil, fmt.Errorf("Error formatting header: %v", err)
	}
	return texture, nil
}

func (s *DebugScreen) createDebugTexture(renderer *sdl.Renderer, c *i6502.Cpu) (*sdl.Texture, error) {
	if s.lastDebugTexture != nil {
		s.lastDebugTexture.Destroy()
	}
	msg := fmt.Sprintf("          %02X      %02X      %02X            %08b      %02X", c.A, c.X, c.Y, c.P, c.SP)
	texture, err := createTexture(msg, s.parent.foreground, s.font, renderer)
	if err != nil {
		return nil, fmt.Errorf("Creating debug texture: %v", err)
	}

	return texture, nil
}

func (s *DebugScreen) createDebugCodeTexture(renderer *sdl.Renderer) (*sdl.Texture, error) {
	if s.lastDebugCodeTexture != nil {
		s.lastDebugCodeTexture.Destroy()
	}
	t, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, s.charWidth*42, s.charHeight*21)
	if err != nil {
		return nil, fmt.Errorf("Unable to create debug code texture: %v", err)
	}

	return t, nil
}

func (s *DebugScreen) createStackTexture(renderer *sdl.Renderer) (*sdl.Texture, error) {
	if s.lastStackTexture != nil {
		s.lastStackTexture.Destroy()
	}
	t, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, s.charWidth*11, s.charHeight*21)
	if err != nil {
		return nil, fmt.Errorf("Unable to create stack texture: %v", err)
	}

	return t, nil
}

func (s *DebugScreen) creatureAllTextures(renderer *sdl.Renderer) error {
	texture, err := s.createDebugTexture(renderer, s.em.GetCPU())
	if err != nil {
		return err
	}
	s.lastDebugTexture = texture

	texture, err = s.createDebugHeaderTexture(renderer, s.em.GetCPU())
	if err != nil {
		return err
	}
	s.debugHeaderTexure = texture

	texture, err = s.createDebugCodeTexture(renderer)
	if err != nil {
		return err
	}
	s.lastDebugCodeTexture = texture

	lastTarget := renderer.GetRenderTarget()
	renderer.SetRenderTarget(s.lastDebugCodeTexture)
	renderer.SetDrawColor(32, 32, 32, 255)
	renderer.Clear()
	renderer.SetDrawColor(s.parent.foreground.R, s.parent.foreground.B, s.parent.foreground.G, s.parent.foreground.A)
	s.DrawCodeLines(renderer, s.em.GetCPU().PC)
	renderer.SetRenderTarget(lastTarget)

	texture, err = s.createStackTexture(renderer)
	if err != nil {
		return err
	}
	s.lastStackTexture = texture

	lastTarget = renderer.GetRenderTarget()
	renderer.SetRenderTarget(s.lastStackTexture)
	renderer.SetDrawColor(32, 32, 32, 255)
	renderer.Clear()
	renderer.SetDrawColor(s.parent.foreground.R, s.parent.foreground.B, s.parent.foreground.G, s.parent.foreground.A)
	s.DrawStack(renderer, s.em)
	renderer.SetRenderTarget(lastTarget)

	s.lastPC = s.em.GetCPU().PC
	return nil
}

// DrawStack draws the stack info to the specified renderer
func (s *DebugScreen) DrawStack(renderer *sdl.Renderer, em EmulatorInterface) error {
	segmentAddress := em.GetCPU().SP
	for i := 0; i < 21; i++ {
		addr := uint16(segmentAddress) + 0x0100
		msg := fmt.Sprintf("%04X:%02X", addr, em.ReadMemory(addr))
		t, err := createTexture(msg, s.parent.foreground, s.font, renderer)
		if err != nil {
			return fmt.Errorf("Error drawing stack: %v", err)
		}
		_, _, w, h, err := t.Query()
		if err != nil {
			return fmt.Errorf("Unable to query stack texture: %v", err)
		}

		renderer.Copy(
			t,
			&sdl.Rect{X: 0, Y: 0, W: w, H: h},
			&sdl.Rect{X: s.charWidth * 2, Y: s.charHeight * int32(i), W: w, H: h},
		)

		segmentAddress++
	}

	return nil
}

func (s *DebugScreen) displayAllTextures(renderer *sdl.Renderer) error {
	_, _, w, h, err := s.debugHeaderTexure.Query()
	if err != nil {
		return fmt.Errorf("Unable to query header texture: %v", err)
	}

	renderer.Copy(
		s.debugHeaderTexure,
		&sdl.Rect{X: 0, Y: 0, W: w, H: h},
		&sdl.Rect{X: 20, Y: 2 + (h / 2), W: w, H: h},
	)

	_, _, w, h, err = s.lastDebugTexture.Query()
	if err != nil {
		return fmt.Errorf("Unable to query debug texture: %v", err)
	}

	renderer.Copy(
		s.lastDebugTexture,
		&sdl.Rect{X: 0, Y: 0, W: w, H: h},
		&sdl.Rect{X: 20, Y: 20 + (h / 2), W: w, H: h},
	)

	_, _, w, h, err = s.lastDebugCodeTexture.Query()
	if err != nil {
		return fmt.Errorf("Unable to query debug code texture: %v", err)
	}

	renderer.Copy(
		s.lastDebugCodeTexture,
		&sdl.Rect{X: 0, Y: 0, W: w, H: h},
		&sdl.Rect{X: 20, Y: 75, W: w, H: h},
	)

	_, _, w, h, err = s.lastStackTexture.Query()
	if err != nil {
		return fmt.Errorf("Unable to query stack texture: %v", err)
	}

	renderer.Copy(
		s.lastStackTexture,
		&sdl.Rect{X: 0, Y: 0, W: w, H: h},
		&sdl.Rect{X: s.charWidth * 52, Y: 75, W: w, H: h},
	)

	return nil
}

func (s *DebugScreen) displayDebugInfo(renderer *sdl.Renderer) error {
	if !s.Active {
		return nil
	}

	if s.lastDebugTexture == nil || s.lastDebugCodeTexture == nil || s.lastPC != s.em.GetCPU().PC {
		if err := s.creatureAllTextures(renderer); err != nil {
			return err
		}
	}

	return s.displayAllTextures(renderer)
}

// DrawCodeLines draws the debug code lines to the screen
func (s *DebugScreen) DrawCodeLines(renderer *sdl.Renderer, PC uint16) error {
	if len(s.debugCode) == 0 {
		return nil
	}

	index := -1
	for i, v := range s.debugCode {
		if v.address == PC {
			index = i
		}
	}

	if index > -1 {
		startIndex := index - 10
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex := index + 11
		if endIndex >= len(s.debugCode) {
			endIndex = len(s.debugCode) - 1
		}

		for i := 0; i < 21; i++ {
			idx := i + index - 10
			if idx < startIndex || idx > endIndex || len(s.debugCode[idx].line) == 0 {
				continue
			}
			t, err := createTexture(s.debugCode[idx].line, s.parent.foreground, s.font, renderer)
			if err != nil {
				return fmt.Errorf("Error rendering debug code lines: %v", err)
			}

			_, _, w, h, err := t.Query()
			if err != nil {
				return fmt.Errorf("Unable to query texture: %v", err)
			}

			if i == 10 {
				r, g, b, a, _ := renderer.GetDrawColor()
				renderer.SetDrawColor(128, 128, 128, 255)
				renderer.FillRect(&sdl.Rect{X: 0, Y: (int32(i) * h), W: s.charWidth * 42, H: h})
				renderer.SetDrawColor(r, g, b, a)
			}

			renderer.Copy(
				t,
				&sdl.Rect{X: 0, Y: 0, W: w, H: h},
				&sdl.Rect{X: s.charWidth, Y: (int32(i) * h), W: w, H: h},
			)

			t.Destroy()
		}
	}

	return nil
}

// DrawScreen draws the entire screen
func (s *DebugScreen) DrawScreen() {
	renderer, err := s.window.GetRenderer()
	if err != nil {
		panic(err)
	}

	renderer.SetDrawColor(s.parent.background.R, s.parent.background.G, s.parent.background.B, s.parent.background.A)
	renderer.Clear()

	if err := s.displayDebugInfo(renderer); err != nil {
		panic(err)
	}
	//	s.DrawCodeLines(renderer, s.em.GetCPU().PC)

	renderer.Present()
}

// CleanUp cleans all resources created by the debug window
func (s *DebugScreen) CleanUp() {
	if s.window == nil {
		return
	}

	if s.lastDebugTexture != nil {
		s.lastDebugTexture.Destroy()
	}
	if s.lastDebugCodeTexture != nil {
		s.lastDebugCodeTexture.Destroy()
	}
	if s.debugHeaderTexure != nil {
		s.debugHeaderTexure.Destroy()
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

func (s *DebugScreen) loadDebugCode() {
	s.debugCode = []codeLine{}
	if len(s.status.RomFilename) > 0 {
		f := filepath.Base(s.status.RomFilename)
		f = strings.Split(f, ".")[0]
		f = filepath.Dir(s.status.RomFilename) + "/" + f + ".debug_code"
		readFile, err := os.Open(f)

		if err != nil {
			return
		}

		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			line := fileScanner.Text()
			if len(strings.Trim(line, " \r\n\t")) > 0 {
				lines := strings.Split(line, " ")
				addrline := strings.Trim(lines[0], "$")
				addr, err := strconv.ParseInt(addrline, 16, 64)
				if err != nil {
					continue
				}
				s.debugCode = append(s.debugCode, codeLine{address: uint16(addr), line: line})
			}
		}

		readFile.Close()
	}
}

// Show shows the debug window
func (s *DebugScreen) Show(em EmulatorInterface, status *ComputerStatus) {
	if !s.Active {
		s.Active = true
		s.em = em
		s.window.Show()
		s.status = status
		s.loadDebugCode()
	}
}

// Hide hides the debug window
func (s *DebugScreen) Hide() {
	if s.Active {
		s.Active = false
		s.window.Hide()
	}
}

func (s *DebugScreen) initializeSymbols() error {
	for x := 32; x < 127; x++ {
		t, err := createTexture(string(rune(x)), s.parent.foreground, s.font, s.renderer)
		if err != nil {
			return err
		}
		t.Destroy()
	}

	return nil
}
