package main

import (
	"fmt"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// SDL_VIDEODRIVER=rpi ./gosdl-linux-arm6

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_INFO)
	// sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)

	if err := ttf.Init(); err != nil {
		log.Panic(err)
	}
	defer ttf.Quit()

	// window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN|sdl.WINDOW_FULLSCREEN)
	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 640, 480, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL) //|sdl.WINDOW_OPENGL
	if err != nil {
		log.Panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Panic(err)
	}
	defer renderer.Destroy()

	// surface, err := window.GetSurface()
	// if err != nil {
	// 	log.Panic(err)
	// }

	font, err := ttf.OpenFont("marken.ttf", 8)
	if err != nil {
		log.Panic(err)
	}

	// surface.FillRect(nil, 0)
	// renderer.SetDrawColor(0, 0, 0, 255)
	// renderer.FillRect(nil)
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	rect := sdl.Rect{X: 16, Y: 16, W: 200, H: 200}
	// surface.FillRect(&rect, 0x424242)
	renderer.SetDrawColor(66, 66, 66, 255)
	renderer.FillRect(&rect)

	fontSurf, err := font.RenderUTF8Solid("Hola Werld", sdl.Color{R: 255, G: 255, B: 255})
	if err != nil {
		log.Panic(err)
	}

	fontTex, err := renderer.CreateTextureFromSurface(fontSurf)
	if err != nil {
		log.Panic(err)
	}

	// fontSurf.Blit(&sdl.Rect{X: 0, Y: 0, W: fontSurf.W, H: fontSurf.H}, surface, &sdl.Rect{X: 20, Y: 20})

	renderer.Copy(fontTex, &sdl.Rect{X: 0, Y: 0, W: fontSurf.W, H: fontSurf.H}, &sdl.Rect{X: 20, Y: 20, W: fontSurf.W, H: fontSurf.H})
	renderer.Present()

	fontSurf.Free()
	fontTex.Destroy()

	// window.UpdateSurface()

	lastTicks := sdl.GetTicks()
	var frameNum uint32
	var fps uint32
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {

			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break

			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE {
					println("Quit")
					running = false
				}
				break
			}
		}

		// time.Now().UnixNano()
		fontSurf, err := font.RenderUTF8Solid(fmt.Sprint("FPS: ", fps), sdl.Color{R: 255, G: 255, B: 255})
		if err != nil {
			log.Panic(err)
		}

		fontTex, err := renderer.CreateTextureFromSurface(fontSurf)
		if err != nil {
			log.Panic(err)
		}

		// Frame counter
		sampleDuration := sdl.GetTicks() - lastTicks

		frameNum++

		if sampleDuration > 100 {
			fps = frameNum * 1000 / sampleDuration
			// log.Printf("frameNum: %d\tsampleDuration: %d\tfps: %d", frameNum, sampleDuration, fps)

			lastTicks = sdl.GetTicks()
			frameNum = 0
		}

		// fontSurf.Blit(&sdl.Rect{X: 0, Y: 0, W: fontSurf.W, H: fontSurf.H}, surface, &sdl.Rect{X: 20, Y: 20})

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		rect := sdl.Rect{X: 16, Y: 16, W: 200, H: 200}
		// surface.FillRect(&rect, 0x424242)
		renderer.SetDrawColor(66, 66, 66, 255)
		renderer.FillRect(&rect)

		// renderer.SetDrawBlendMode(sdl.BLENDMODE_ADD)
		renderer.Copy(fontTex, nil, &sdl.Rect{X: 20, Y: 20, W: fontSurf.W, H: fontSurf.H})

		fontTex.Destroy()
		fontSurf.Free()

		renderer.Present()
		// window.UpdateSurface()
	}
}
