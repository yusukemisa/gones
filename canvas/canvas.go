package canvas

import (
	"fmt"
	"image/color"
	"os"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type SDL2Canvas struct {
	windowWidth  int
	windowHeight int
	window       *sdl.Window
	Renderer     *sdl.Renderer
	texture      *sdl.Texture
	pixels       []byte
	event        sdl.Event
	err          error
	Running      bool
	// Mouse Event Handling
	MouseClicked bool
	MouseX       int32
	MouseY       int32
}

// Setup Window / Renderer / texture
func (s *SDL2Canvas) Setup(title string, windowWidth int, windowHeight int) {
	sdl.Init(sdl.INIT_EVERYTHING)

	var flags uint32 = sdl.WINDOW_SHOWN

	s.windowWidth = windowWidth
	s.windowHeight = windowHeight

	s.window, s.err = sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_CENTERED, // 画面上のどこにウィンドウを表示するか。0,0の場合左上。基本centerでいいはず
		sdl.WINDOWPOS_CENTERED,
		int32(windowWidth),
		int32(windowHeight),
		flags,
	)
	if s.err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Window: %s\n", s.err)
		os.Exit(1)
	}

	s.Renderer, s.err = sdl.CreateRenderer(s.window, -1, sdl.RENDERER_ACCELERATED)
	if s.err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Renderer: %s\n", s.err)
		os.Exit(1)
	}

	s.texture, s.err = s.Renderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_STREAMING,
		int32(windowWidth), int32(windowHeight))
	if s.err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", s.texture)
		os.Exit(1)
	}

	//s.pixels = InitPixels()
	//p := &s.pixels
	//fmt.Println(p)
	//
	//s.Renderer.Clear()
	//
	//s.Renderer.SetDrawColor(0xFF, 0xFF, 0xFF, 0)
	//s.Renderer.DrawLine(0, 0, 256, 240)
	//s.Renderer.ReadPixels(nil, sdl.PIXELFORMAT_RGB24, unsafe.Pointer(p), 256*240*3)
	//fmt.Println(p)
	//s.Renderer.DrawLines(
	//	[]sdl.Point{
	//		{0, 0},
	//		{10, 10},
	//		{50, 10},
	//		{30, 40},
	//	})
	//s.Renderer.Present()
	//s.Update()
	//s.Render()
	s.Running = true
}

func (s *SDL2Canvas) SetPixel(x int, y int, c *color.RGBA) {
	s.Renderer.SetDrawColor(c.R, c.R, c.B, 0)
	s.Renderer.DrawPoint(int32(x), int32(y))
}

func (s *SDL2Canvas) Update() {
	s.texture.Update(nil, unsafe.Pointer(&s.pixels), s.windowWidth*3)
}

func (s *SDL2Canvas) Render() {
	s.Renderer.Clear()

	s.Renderer.Copy(s.texture, nil, nil)
	s.Renderer.Present()
}

func (s *SDL2Canvas) Shutdown() {
	s.texture.Destroy()
	s.Renderer.Destroy()
	s.window.Destroy()
	sdl.Quit()
}
