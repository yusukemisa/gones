package canvas

import (
	"fmt"
	"image/color"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// Canvas defines the interface for rendering operations
type Canvas interface {
	Setup(title string, width, height int) error
	SetPixel(x, y int, c *color.RGBA)
	Update()
	Render()
	Shutdown()
	Clear()
	Present()
}

type SDL2Canvas struct {
	windowWidth  int
	windowHeight int
	window       *sdl.Window
	renderer     *sdl.Renderer
	texture      *sdl.Texture
	pixels       []byte
	event        sdl.Event
	err          error
	Running      bool
	debug        bool
	// Mouse Event Handling
	MouseClicked bool
	MouseX       int32
	MouseY       int32
}

// Setup Window / Renderer / texture
func (s *SDL2Canvas) Setup(title string, windowWidth int, windowHeight int) error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return fmt.Errorf("failed to initialize SDL: %v", err)
	}

	var flags uint32 = sdl.WINDOW_SHOWN

	s.windowWidth = windowWidth
	s.windowHeight = windowHeight
	s.debug = true

	window, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(windowWidth),
		int32(windowHeight),
		flags,
	)
	if err != nil {
		return fmt.Errorf("failed to create window: %v", err)
	}
	s.window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %v", err)
	}
	s.renderer = renderer

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		int32(windowWidth),
		int32(windowHeight),
	)
	if err != nil {
		return fmt.Errorf("failed to create texture: %v", err)
	}
	s.texture = texture

	// Initialize pixel buffer
	s.pixels = make([]byte, windowWidth*windowHeight*4)

	s.Running = true
	return nil
}

func (s *SDL2Canvas) SetPixel(x int, y int, c *color.RGBA) {
	if x < 0 || x >= s.windowWidth || y < 0 || y >= s.windowHeight {
		return
	}
	offset := (y*s.windowWidth + x) * 4
	s.pixels[offset] = c.R
	s.pixels[offset+1] = c.G
	s.pixels[offset+2] = c.B
	s.pixels[offset+3] = c.A
}

func (s *SDL2Canvas) Update() {
	s.texture.Update(nil, unsafe.Pointer(&s.pixels[0]), s.windowWidth*4)
}

func (s *SDL2Canvas) Render() {
	s.Update()
	s.renderer.Copy(s.texture, nil, nil)
}

func (s *SDL2Canvas) Clear() {
	for i := range s.pixels {
		s.pixels[i] = 0
	}
}

func (s *SDL2Canvas) Present() {
	s.Render()
	s.renderer.Present()
}

func (s *SDL2Canvas) Shutdown() {
	if s.texture != nil {
		s.texture.Destroy()
	}
	if s.renderer != nil {
		s.renderer.Destroy()
	}
	if s.window != nil {
		s.window.Destroy()
	}
	sdl.Quit()
}
