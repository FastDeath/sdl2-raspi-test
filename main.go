package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var test int
var stat string

func testGo() {
	for {
		test++
		t := time.Now()
		stat = t.Format(time.Kitchen)
		time.Sleep(1 * time.Millisecond)
	}
}

func drawText(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color, rect *sdl.Rect) {
	fontSurf, err := font.RenderUTF8Solid(text, color)
	if err != nil {
		log.Panic(err)
	}

	if rect.W == 0 && rect.H == 0 {
		rect.W = fontSurf.W
		rect.H = fontSurf.H
	}

	if rect.X < 0 {
		rect.X += 256 - rect.W
	}
	if rect.Y < 0 {
		rect.Y += 64 - rect.H
	}

	fontTex, err := renderer.CreateTextureFromSurface(fontSurf)
	if err != nil {
		log.Panic(err)
	}

	renderer.Copy(fontTex, nil, rect)

	fontSurf.Free()
	fontTex.Destroy()
}

// SDL_VIDEODRIVER=rpi ./gosdl-linux-arm6

func main() {
	os.Chdir("/root")

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_INFO)
	// sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	sdl.ShowCursor(0)

	if err := ttf.Init(); err != nil {
		log.Panic(err)
	}
	defer ttf.Quit()

	// window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN|sdl.WINDOW_FULLSCREEN)
	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 256, 64, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL) //|sdl.WINDOW_OPENGL
	if err != nil {
		log.Panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		log.Panic(err)
	}
	defer renderer.Destroy()

	// flags := img.INIT_TIF // |img.INIT_PNG
	// if flagsInit := img.Init(flags); flagsInit&flags != flags {
	// 	log.Panic("img.Init() returned ", flagsInit, ": ", img.GetError())
	// }
	// defer img.Quit()

	surfDB, err := img.Load("db.png")
	if err != nil {
		log.Panic("img.Load(): ", err)
	}

	texDB, err := renderer.CreateTextureFromSurface(surfDB)
	if err != nil {
		log.Panic("renderer.CreateTextureFromSurface(surfDB): ", err)
	}

	// texDB, err := img.LoadTexture(renderer, "db.png")
	// if err != nil {
	// 	log.Panic("img.LoadTexture(): ", err)
	// }

	// surface, err := window.GetSurface()
	// if err != nil {
	// 	log.Panic(err)
	// }

	font, err := ttf.OpenFont("marken.ttf", 8)
	if err != nil {
		log.Panic(err)
	}

	/*
		// surface.FillRect(nil, 0)
		// renderer.SetDrawColor(0, 0, 0, 255)
		// renderer.FillRect(nil)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		// rect := sdl.Rect{X: 16, Y: 16, W: 200, H: 200}
		// // surface.FillRect(&rect, 0x424242)
		// renderer.SetDrawColor(66, 66, 66, 255)
		// renderer.FillRect(&rect)

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
	*/

	// window.UpdateSurface()

	go testGo()

	var x, xo int32
	var xs int32 = 1

	lastTicks := sdl.GetTicks()
	var frameNum uint32
	var fps uint32
	var textColor uint8
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

		textColor = 127 + uint8(math.Sin(float64(sdl.GetTicks())/100)*128)

		drawText(renderer, font,
			fmt.Sprint("Döner, groß  6,50    FPS: ", fps, " test=", test),
			sdl.Color{R: textColor, G: textColor, B: textColor},
			&sdl.Rect{X: 0, Y: 0})
		// drawText(renderer, font, stat, sdl.Color{R: 255, G: 255, B: 255}, &sdl.Rect{X: -1, Y: 0})

		h := 26 + int32(math.Sin(float64(x)/4)*4)

		// rect := sdl.Rect{X: 16, Y: 16, W: xo + xs*x, H: 10}
		rect := sdl.Rect{X: 16 + xo + xs*x, Y: 46 - h, W: 21, H: h}
		// renderer.SetDrawColor(66, 66, 66, 255)
		// renderer.SetDrawColor(255, 255, 255, 255)
		// renderer.FillRect(&rect)
		// renderer.Copy(texDB, nil, &rect)

		// sdl.RenderCopyEx(renderer, texture, &srcrect, &dstrect, angle, &center, sdl.FLIP_VERTICAL)

		var flip sdl.RendererFlip
		if xs == 1 {
			flip = sdl.FLIP_HORIZONTAL
		} else {
			flip = sdl.FLIP_NONE
		}

		renderer.CopyEx(texDB, nil, &rect, 0, nil, flip)

		// renderer.Copy(texDB, nil, &sdl.Rect{X: 118, Y: 64 - h, W: 21, H: h})

		x += 1
		if x >= 200 {
			x %= 200
			xs *= -1
			xo = 200 - xo
			fps = fps
		}

		// renderer.SetDrawBlendMode(sdl.BLENDMODE_ADD)

		renderer.Present()

		// window.UpdateSurface()
	}
}
