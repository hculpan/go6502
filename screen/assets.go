package screen

import (
	"fmt"

	"github.com/hculpan/go6502/resources"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var count int

func loadFont(filename string) (*ttf.Font, *ttf.GlyphMetrics, error) {
	// Load data from bindata in resources/resources.go
	data, err := resources.Asset("resources/" + filename)
	if err != nil {
		return nil, nil, err
	}

	rwops, err := sdl.RWFromMem(data)
	if err != nil {
		return nil, nil, err
	}

	font, err := ttf.OpenFontRW(rwops, 1, 18)
	if err != nil {
		return nil, nil, err
	}

	fontmetrics, _ := font.GlyphMetrics('A')

	return font, fontmetrics, nil
}

func getCharacterMetrics(g *ttf.GlyphMetrics) (x, y int32) {
	return int32(g.MaxX), int32(g.MaxY + g.Advance)
}

func createTexture(msg string, c sdl.Color, font *ttf.Font, renderer *sdl.Renderer) (*sdl.Texture, error) {
	surface, err := font.RenderUTF8Solid(msg, c)
	if err != nil {
		if err.Error() == "Text has zero width" {
			return nil, nil
		}

		return nil, fmt.Errorf("Unable to create texture surface: %v", err)
	}
	defer surface.Free()

	msgtext, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, fmt.Errorf("Unable to create texture from surface: %v", err)
	}

	return msgtext, nil
}
