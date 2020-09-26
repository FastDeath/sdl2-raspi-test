package main

import (
	"fmt"
	"log"
	"os"
	"unsafe"

	// "github.com/go-gl/gl/v2.1/gl"
	// gl "github.com/go-gl/gl/v3.1/gles2"
	gl "github.com/remogatto/opengles2"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"github.com/inkyblackness/imgui-go"
)

// SDL_VIDEODRIVER=rpi ./gosdl-linux-arm6

var SomeValue float32 = 13

func main() {
	os.Chdir("/root")

	// Initialize ImGui
	context := imgui.CreateContext(nil)
	defer context.Destroy()

	io := imgui.CurrentIO()
	io.SetDisplaySize(imgui.Vec2{X: 256, Y: 64})
	io.Fonts().AddFontFromFileTTF("ProggyTiny.ttf", 10)
	io.Fonts().BuildWithFreeTypeV(imgui.FreeTypeRasterizerFlagsMonochrome | imgui.FreeTypeRasterizerFlagsMonoHinting)
	setStyle()
	// io.Fonts().TextureDataRGBA32()

	// Initialize SDL2
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

	_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 2)
	_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)

	_ = sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	_ = sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)
	_ = sdl.GLSetAttribute(sdl.GL_STENCIL_SIZE, 8)

	// window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN|sdl.WINDOW_FULLSCREEN)
	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 256, 64, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL) //|sdl.WINDOW_OPENGL
	if err != nil {
		log.Panic(err)
	}
	defer window.Destroy()

	glContext, err := window.GLCreateContext()
	if err != nil {
		log.Panic("failed to create OpenGL context: ", err)
	}
	err = window.GLMakeCurrent(glContext)
	if err != nil {
		log.Panic("failed to set current OpenGL context: ", err)
	}

	_ = sdl.GLSetSwapInterval(1)

	// Important! Call gl.Init only under the presence of an active OpenGL context,
	// i.e., after MakeContextCurrent.
	// if err := gl.Init(); err != nil {
	// 	log.Fatalln("gl.Init() failed: ", err)
	// }

	log.Println("VENDOR = ", gl.GetString(gl.VENDOR))
	log.Println("RENDERER = ", gl.GetString(gl.RENDERER))
	log.Println("VERSION = ", gl.GetString(gl.VERSION))

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	//*******************************************************************
	// Define the viewport dimensions
	//*******************************************************************
	gl.Viewport(0, 0, 256, 64)

	var shader = initShader()
	var verticesVBO, elementsHandle = initPaint(io, &shader)

	log.Println("painting...")

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
		imgui.NewFrame()

		imgui.SetNextWindowPos(imgui.Vec2{0, 0})
		imgui.SetNextWindowSize(imgui.Vec2{256, 64})

		title := fmt.Sprintf("%v %v %v", gl.GetString(gl.VENDOR), gl.GetString(gl.RENDERER), gl.GetString(gl.VERSION))
		imgui.Begin(title)
		// imgui.LabelText("Hello", "World")
		SomeValue = float32(sdl.GetTicks()) / 1000

		imgui.ColumnsV(4, "Label", true)

		imgui.DragFloatV("1.", &SomeValue, 10, 0, 100, "%.2fV", 1)
		imgui.DragFloatV("2.", &SomeValue, 10, 0, 100, "%.2fV", 1)
		imgui.DragFloatV("3.", &SomeValue, 10, 0, 100, "%.2fV", 1)

		imgui.NextColumn()

		imgui.DragFloatV("Floaty", &SomeValue, 10, 0, 100, "%.3f", 1)
		imgui.DragFloatV("Some value", &SomeValue, 10, 0, 100, "%.3f", 1)

		imgui.End()
		imgui.Render()

		// gl.GenBuffers()

		paint(verticesVBO, elementsHandle, &shader)

		window.GLSwap()

		// RenderGL([2]float32{256, 64}, [2]float32{256, 64}, drawData)
	}
}

// ShaderProps contains the location to shader uniforms/attributes
type ShaderProps struct {
	Handle   uint32
	Uniforms struct {
		Texture, ProjMtx int32
	}
	Attributes struct {
		Position, UV, Color uint32
	}
}

func initShader() (props ShaderProps) {

	vertexSource := `
uniform mat4 ProjMtx;
attribute vec2 Position;
attribute vec2 UV;
attribute vec4 Color;
varying vec2 Frag_UV;
varying vec4 Frag_Color;
void main()
{
	Frag_UV = UV;
	Frag_Color = Color;
	gl_Position = ProjMtx * vec4(Position.xy,0,1);
}
`
	fragmentSource := `
uniform sampler2D Texture;
varying vec2 Frag_UV;
varying vec4 Frag_Color;
void main()
{
	// gl_FragColor = Frag_Color;
	// gl_FragColor = vec4(Frag_Color.rgb, Frag_Color.a * texture2D(Texture, Frag_UV.st).r);
	gl_FragColor = vec4(Frag_Color.rgb, Frag_Color.a * texture2D(Texture, Frag_UV.st).a);
}
`

	log.Println("creating shader program...")
	props.Handle = gl.CreateProgram()
	var vertHandle = gl.CreateShader(gl.VERTEX_SHADER)
	var fragHandle = gl.CreateShader(gl.FRAGMENT_SHADER)
	log.Println("shaderHandle =", props.Handle)

	gl.ShaderSource(vertHandle, 1, &vertexSource, nil)
	gl.ShaderSource(fragHandle, 1, &fragmentSource, nil)

	var vstatus int32

	gl.CompileShader(vertHandle)
	gl.GetShaderiv(vertHandle, gl.COMPILE_STATUS, &vstatus)
	fmt.Printf("Compiled Vertex Shader: %v\n", vstatus)

	var nBytes gl.Sizei
	var strLog = gl.GetShaderInfoLog(vertHandle, 256, &nBytes)
	log.Println("Vertex log:", strLog)

	gl.CompileShader(fragHandle)
	gl.GetShaderiv(fragHandle, gl.COMPILE_STATUS, &vstatus)
	fmt.Printf("Compiled Fragment Shader: %v\n", vstatus)

	strLog = gl.GetShaderInfoLog(vertHandle, 256, &nBytes)
	log.Println("Fragment log:", strLog)

	gl.AttachShader(props.Handle, vertHandle)
	gl.AttachShader(props.Handle, fragHandle)

	// gl.BindAttribLocation(props.Handle, 0, "a_vertex")
	gl.LinkProgram(props.Handle)
	log.Println("is program:", gl.IsProgram(props.Handle))

	strLog = gl.GetProgramInfoLog(props.Handle, 256, &nBytes)
	log.Println("program log:", strLog)

	props.Uniforms.Texture = int32(gl.GetUniformLocation(props.Handle, "Texture"))
	props.Uniforms.ProjMtx = int32(gl.GetUniformLocation(props.Handle, "ProjMtx"))
	props.Attributes.Position = gl.GetAttribLocation(props.Handle, "Position")
	props.Attributes.UV = gl.GetAttribLocation(props.Handle, "UV")
	props.Attributes.Color = gl.GetAttribLocation(props.Handle, "Color")

	log.Println("props.Uniforms.Texture:", props.Uniforms.Texture)
	log.Println("props.Uniforms.ProjMtx:", props.Uniforms.ProjMtx)
	log.Println("props.Attributes.Position:", props.Attributes.Position)
	log.Println("props.Attributes.UV:", props.Attributes.UV)
	log.Println("props.Attributes.Color:", props.Attributes.Color)
	return
}

func initPaint(io imgui.IO, shader *ShaderProps) (verticesVBO, elementsHandle uint32) {
	gl.GenBuffers(1, &verticesVBO)
	log.Println("verticesVBO =", verticesVBO)
	gl.GenBuffers(1, &elementsHandle)
	log.Println("elementsHandle =", elementsHandle)

	// gl.BindBuffer(gl.ARRAY_BUFFER, verticesVBO)

	// var points = []float32{
	// 	-0.5, 0.0, 0.0, 1.0,
	// 	0.5, 0.0, 0.0, 1.0,
	// 	0.0, 0.5, 0.0, 1.0,
	// }
	// // gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(points)), unsafe.Pointer(&points), gl.STATIC_DRAW)

	// log.Println("sizeof(points) =", unsafe.Sizeof(points))
	// gl.BufferData(gl.ARRAY_BUFFER, 4*3*4, gl.Void(&points[0]), gl.STATIC_DRAW)
	// gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.UseProgram(shader.Handle)
	// var ptr interface{} = gl.Void(uintptr(0))
	// log.Println("ptr", reflect.ValueOf(ptr).Pointer(), unsafe.Pointer(reflect.ValueOf(ptr).Pointer()))
	// gl.VertexAttribPointer(vertexLocation, 4, gl.FLOAT, false, 0, ptr)
	// gl.EnableVertexAttribArray(shader.Attributes.Position)
	// gl.VertexAttribPointer(shader.Attributes.Position, 4, gl.FLOAT, false, 0, unsafe.Pointer(uintptr(0)))

	image := io.Fonts().TextureDataAlpha8()

	var fontTexture uint32
	// Upload texture to graphics system
	gl.GenTextures(1, &fontTexture)
	gl.BindTexture(gl.TEXTURE_2D, fontTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	// gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, gl.Sizei(image.Width), gl.Sizei(image.Height),
		0, gl.ALPHA, gl.UNSIGNED_BYTE, gl.Void(image.Pixels))

	log.Println("texture id =", fontTexture)
	// Store our identifier
	io.Fonts().SetTextureID(imgui.TextureID(fontTexture))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.ClearColor(0, 0, 0, 0)
	return
}

func paint(verticesVBO, elementsHandle uint32, shader *ShaderProps) {
	// gl.Viewport(0, 0, 256, 64)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	// gl.Enable(gl.SCISSOR_TEST)
	// gl.Scissor(120, 0, gl.Sizei(256-120), gl.Sizei(64))

	// gl.BindBuffer(gl.ARRAY_BUFFER, verticesVBO)
	// gl.DrawArrays(gl.TRIANGLES, 0, 3)
	drawData := imgui.RenderedDrawData()

	// Setup render state: alpha-blending enabled, no face culling, no depth testing, scissor enabled, polygon fill
	gl.Enable(gl.BLEND)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.SCISSOR_TEST)
	// gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	// Setup viewport, orthographic projection matrix
	// Our visible imgui space lies from draw_data->DisplayPos (top left) to draw_data->DisplayPos+data_data->DisplaySize (bottom right).
	// DisplayMin is typically (0,0) for single viewport apps.
	gl.Viewport(0, 0, gl.Sizei(256), gl.Sizei(64))
	orthoProjection := [4][4]float32{
		{2.0 / 256, 0.0, 0.0, 0.0},
		{0.0, 2.0 / -64, 0.0, 0.0},
		{0.0, 0.0, -1.0, 0.0},
		{-1.0, 1.0, 0.0, 1.0},
	}
	gl.UseProgram(shader.Handle)
	gl.Uniform1i(shader.Uniforms.Texture, 0)
	gl.UniformMatrix4fv(shader.Uniforms.ProjMtx, gl.Sizei(1), false, &orthoProjection[0][0])
	// gl.BindSampler(0, 0) // Rely on combined texture/sampler state.

	gl.EnableVertexAttribArray(uint32(shader.Attributes.Position))
	gl.EnableVertexAttribArray(uint32(shader.Attributes.UV))
	gl.EnableVertexAttribArray(uint32(shader.Attributes.Color))

	vertexSize, vertexOffsetPos, vertexOffsetUv, vertexOffsetCol := imgui.VertexBufferLayout()
	gl.VertexAttribPointer(uint32(shader.Attributes.Position), 2, gl.FLOAT, false, gl.Sizei(vertexSize), unsafe.Pointer(uintptr(vertexOffsetPos)))
	gl.VertexAttribPointer(uint32(shader.Attributes.UV), 2, gl.FLOAT, false, gl.Sizei(vertexSize), unsafe.Pointer(uintptr(vertexOffsetUv)))
	gl.VertexAttribPointer(uint32(shader.Attributes.Color), 4, gl.UNSIGNED_BYTE, true, gl.Sizei(vertexSize), unsafe.Pointer(uintptr(vertexOffsetCol)))

	indexSize := imgui.IndexBufferLayout()
	drawType := gl.UNSIGNED_SHORT
	const bytesPerUint32 = 4
	if indexSize == bytesPerUint32 {
		drawType = gl.UNSIGNED_INT
	}

	// Draw
	for _, list := range drawData.CommandLists() {
		var indexBufferOffset uintptr

		vertexBuffer, vertexBufferSize := list.VertexBuffer()
		gl.BindBuffer(gl.ARRAY_BUFFER, verticesVBO)
		gl.BufferData(gl.ARRAY_BUFFER, gl.SizeiPtr(vertexBufferSize), gl.Void(vertexBuffer), gl.STREAM_DRAW)

		indexBuffer, indexBufferSize := list.IndexBuffer()
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, elementsHandle)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.SizeiPtr(indexBufferSize), gl.Void(indexBuffer), gl.STREAM_DRAW)

		for _, cmd := range list.Commands() {
			if cmd.HasUserCallback() {
				cmd.CallUserCallback(list)
			} else {
				gl.BindTexture(gl.TEXTURE_2D, uint32(cmd.TextureID()))
				clipRect := cmd.ClipRect()
				gl.Scissor(int32(clipRect.X), int32(64)-int32(clipRect.W), gl.Sizei(clipRect.Z-clipRect.X), gl.Sizei(clipRect.W-clipRect.Y))
				gl.DrawElements(gl.TRIANGLES, gl.Sizei(cmd.ElementCount()), gl.Enum(drawType), gl.Void(indexBufferOffset))
			}
			indexBufferOffset += uintptr(cmd.ElementCount() * indexSize)
		}
	}

	gl.Flush()
	gl.Finish()
}

// // PreRender clears the framebuffer.
// func PreRender(clearColor [3]float32) {
// 	gl.ClearColor(clearColor[0], clearColor[1], clearColor[2], 1.0)
// 	gl.Clear(gl.COLOR_BUFFER_BIT)
// }

// // Render translates the ImGui draw data to OpenGL3 commands.
// func Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData) {
// 	// Avoid rendering when minimized, scale coordinates for retina displays (screen coordinates != framebuffer coordinates)
// 	displayWidth, displayHeight := displaySize[0], displaySize[1]
// 	fbWidth, fbHeight := framebufferSize[0], framebufferSize[1]
// 	if (fbWidth <= 0) || (fbHeight <= 0) {
// 		return
// 	}
// 	drawData.ScaleClipRects(imgui.Vec2{
// 		X: fbWidth / displayWidth,
// 		Y: fbHeight / displayHeight,
// 	})

// 	// Backup GL state
// 	var lastActiveTexture = gl.GetInteger(gl.ACTIVE_TEXTURE)
// 	gl.ActiveTexture(gl.TEXTURE0)
// 	var lastProgram = gl.GetInteger(gl.CURRENT_PROGRAM)
// 	var lastTexture = gl.GetInteger(gl.TEXTURE_BINDING_2D)
// 	// var lastSampler = gl.GetInteger(gl.SAMPLER_BINDING)
// 	var lastArrayBuffer = gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
// 	var lastElementArrayBuffer = gl.GetInteger(gl.ELEMENT_ARRAY_BUFFER_BINDING)
// 	// var lastVertexArray = gl.GetInteger(gl.VERTEX_ARRAY_BINDING)
// 	// var lastPolygonMode = make([]int32, 2)
// 	// gl.GetIntegerv(gl.POLYGON_MODE, lastPolygonMode)
// 	var lastViewport = make([]int32, 4)
// 	gl.GetIntegerv(gl.VIEWPORT, lastViewport)
// 	var lastScissorBox = make([]int32, 4)
// 	gl.GetIntegerv(gl.SCISSOR_BOX, lastScissorBox)
// 	var lastBlendSrcRgb = gl.GetInteger(gl.BLEND_SRC_RGB)
// 	var lastBlendDstRgb = gl.GetInteger(gl.BLEND_DST_RGB)
// 	var lastBlendSrcAlpha = gl.GetInteger(gl.BLEND_SRC_ALPHA)
// 	var lastBlendDstAlpha = gl.GetInteger(gl.BLEND_DST_ALPHA)
// 	var lastBlendEquationRgb = gl.GetInteger(gl.BLEND_EQUATION_RGB)
// 	var lastBlendEquationAlpha = gl.GetInteger(gl.BLEND_EQUATION_ALPHA)
// 	lastEnableBlend := gl.IsEnabled(gl.BLEND)
// 	lastEnableCullFace := gl.IsEnabled(gl.CULL_FACE)
// 	lastEnableDepthTest := gl.IsEnabled(gl.DEPTH_TEST)
// 	lastEnableScissorTest := gl.IsEnabled(gl.SCISSOR_TEST)

// 	// Setup render state: alpha-blending enabled, no face culling, no depth testing, scissor enabled, polygon fill
// 	gl.Enable(gl.BLEND)
// 	gl.BlendEquation(gl.FUNC_ADD)
// 	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
// 	gl.Disable(gl.CULL_FACE)
// 	gl.Disable(gl.DEPTH_TEST)
// 	gl.Enable(gl.SCISSOR_TEST)
// 	// gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

// 	// Setup viewport, orthographic projection matrix
// 	// Our visible imgui space lies from draw_data->DisplayPos (top left) to draw_data->DisplayPos+data_data->DisplaySize (bottom right).
// 	// DisplayMin is typically (0,0) for single viewport apps.
// 	gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))
// 	orthoProjection := [4][4]float32{
// 		{2.0 / displayWidth, 0.0, 0.0, 0.0},
// 		{0.0, 2.0 / -displayHeight, 0.0, 0.0},
// 		{0.0, 0.0, -1.0, 0.0},
// 		{-1.0, 1.0, 0.0, 1.0},
// 	}
// 	gl.UseProgram(renderer.shaderHandle)
// 	gl.Uniform1i(renderer.attribLocationTex, 0)
// 	gl.UniformMatrix4fv(renderer.attribLocationProjMtx, 1, false, &orthoProjection[0][0])
// 	// gl.BindSampler(0, 0) // Rely on combined texture/sampler state.

// 	// Recreate the VAO every time
// 	// (This is to easily allow multiple GL contexts. VAO are not shared among GL contexts, and
// 	// we don't track creation/deletion of windows so we don't have an obvious key to use to cache them.)
// 	var vaoHandle uint32
// 	gl.GenVertexArrays(1, &vaoHandle)
// 	gl.BindVertexArray(vaoHandle)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, renderer.vboHandle)
// 	gl.EnableVertexAttribArray(uint32(renderer.attribLocationPosition))
// 	gl.EnableVertexAttribArray(uint32(renderer.attribLocationUV))
// 	gl.EnableVertexAttribArray(uint32(renderer.attribLocationColor))
// 	vertexSize, vertexOffsetPos, vertexOffsetUv, vertexOffsetCol := imgui.VertexBufferLayout()
// 	gl.VertexAttribPointer(uint32(renderer.attribLocationPosition), 2, gl.FLOAT, false, int32(vertexSize), unsafe.Pointer(uintptr(vertexOffsetPos)))
// 	gl.VertexAttribPointer(uint32(renderer.attribLocationUV), 2, gl.FLOAT, false, int32(vertexSize), unsafe.Pointer(uintptr(vertexOffsetUv)))
// 	gl.VertexAttribPointer(uint32(renderer.attribLocationColor), 4, gl.UNSIGNED_BYTE, true, int32(vertexSize), unsafe.Pointer(uintptr(vertexOffsetCol)))
// 	indexSize := imgui.IndexBufferLayout()
// 	drawType := gl.UNSIGNED_SHORT
// 	const bytesPerUint32 = 4
// 	if indexSize == bytesPerUint32 {
// 		drawType = gl.UNSIGNED_INT
// 	}

// 	// Draw
// 	for _, list := range drawData.CommandLists() {
// 		var indexBufferOffset uintptr

// 		vertexBuffer, vertexBufferSize := list.VertexBuffer()
// 		gl.BindBuffer(gl.ARRAY_BUFFER, renderer.vboHandle)
// 		gl.BufferData(gl.ARRAY_BUFFER, vertexBufferSize, vertexBuffer, gl.STREAM_DRAW)

// 		indexBuffer, indexBufferSize := list.IndexBuffer()
// 		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, renderer.elementsHandle)
// 		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexBufferSize, indexBuffer, gl.STREAM_DRAW)

// 		for _, cmd := range list.Commands() {
// 			if cmd.HasUserCallback() {
// 				cmd.CallUserCallback(list)
// 			} else {
// 				gl.BindTexture(gl.TEXTURE_2D, uint32(cmd.TextureID()))
// 				clipRect := cmd.ClipRect()
// 				gl.Scissor(int32(clipRect.X), int32(fbHeight)-int32(clipRect.W), int32(clipRect.Z-clipRect.X), int32(clipRect.W-clipRect.Y))
// 				gl.DrawElements(gl.TRIANGLES, int32(cmd.ElementCount()), uint32(drawType), unsafe.Pointer(indexBufferOffset))
// 			}
// 			indexBufferOffset += uintptr(cmd.ElementCount() * indexSize)
// 		}
// 	}
// 	gl.DeleteVertexArrays(1, &vaoHandle)

// 	// Restore modified GL state
// 	gl.UseProgram(uint32(lastProgram))
// 	gl.BindTexture(gl.TEXTURE_2D, uint32(lastTexture))
// 	gl.BindSampler(0, uint32(lastSampler))
// 	gl.ActiveTexture(uint32(lastActiveTexture))
// 	gl.BindVertexArray(uint32(lastVertexArray))
// 	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(lastArrayBuffer))
// 	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, uint32(lastElementArrayBuffer))
// 	gl.BlendEquationSeparate(uint32(lastBlendEquationRgb), uint32(lastBlendEquationAlpha))
// 	gl.BlendFuncSeparate(uint32(lastBlendSrcRgb), uint32(lastBlendDstRgb), uint32(lastBlendSrcAlpha), uint32(lastBlendDstAlpha))
// 	if lastEnableBlend {
// 		gl.Enable(gl.BLEND)
// 	} else {
// 		gl.Disable(gl.BLEND)
// 	}
// 	if lastEnableCullFace {
// 		gl.Enable(gl.CULL_FACE)
// 	} else {
// 		gl.Disable(gl.CULL_FACE)
// 	}
// 	if lastEnableDepthTest {
// 		gl.Enable(gl.DEPTH_TEST)
// 	} else {
// 		gl.Disable(gl.DEPTH_TEST)
// 	}
// 	if lastEnableScissorTest {
// 		gl.Enable(gl.SCISSOR_TEST)
// 	} else {
// 		gl.Disable(gl.SCISSOR_TEST)
// 	}
// 	gl.PolygonMode(gl.FRONT_AND_BACK, uint32(lastPolygonMode[0]))
// 	gl.Viewport(lastViewport[0], lastViewport[1], lastViewport[2], lastViewport[3])
// 	gl.Scissor(lastScissorBox[0], lastScissorBox[1], lastScissorBox[2], lastScissorBox[3])
// }

// func RenderGL2(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData) {
// 	// Avoid rendering when minimized, scale coordinates for retina displays (screen coordinates != framebuffer coordinates)
// 	displayWidth, displayHeight := displaySize[0], displaySize[1]
// 	fbWidth, fbHeight := framebufferSize[0], framebufferSize[1]
// 	if (fbWidth <= 0) || (fbHeight <= 0) {
// 		return
// 	}
// 	drawData.ScaleClipRects(imgui.Vec2{
// 		X: fbWidth / displayWidth,
// 		Y: fbHeight / displayHeight,
// 	})

// 	// Setup render state: alpha-blending enabled, no face culling, no depth testing, scissor enabled, vertex/texcoord/color pointers, polygon fill.
// 	var lastTexture = gl.GetInteger(gl.TEXTURE_BINDING_2D)
// 	// var lastPolygonMode = make([]int32, 2)
// 	// gl.GetIntegerv(gl.POLYGON_MODE, lastPolygonMode)
// 	var lastViewport = make([]int32, 4)
// 	gl.GetIntegerv(gl.VIEWPORT, lastViewport)
// 	var lastScissorBox = make([]int32, 4)
// 	gl.GetIntegerv(gl.SCISSOR_BOX, lastScissorBox)
// 	// gl.PushAttrib(gl.ENABLE_BIT | gl.COLOR_BUFFER_BIT | gl.TRANSFORM_BIT)
// 	gl.Enable(gl.BLEND)
// 	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
// 	gl.Disable(gl.CULL_FACE)
// 	gl.Disable(gl.DEPTH_TEST)
// 	// gl.Disable(gl.LIGHTING)
// 	// gl.Disable(gl.COLOR_MATERIAL)
// 	gl.Enable(gl.SCISSOR_TEST)
// 	// gl.EnableClientState(gl.VERTEX_ARRAY)
// 	// gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
// 	// gl.EnableClientState(gl.COLOR_ARRAY)
// 	gl.Enable(gl.TEXTURE_2D)
// 	// gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

// 	// You may want this if using this code in an OpenGL 3+ context where shaders may be bound
// 	// gl.UseProgram(0)

// 	// Setup viewport, orthographic projection matrix
// 	// Our visible imgui space lies from draw_data->DisplayPos (top left) to draw_data->DisplayPos+data_data->DisplaySize (bottom right).
// 	// DisplayMin is typically (0,0) for single viewport apps.
// 	gl.Viewport(0, 0, int(fbWidth), int(fbHeight))
// 	// gl.MatrixMode(gl.PROJECTION)
// 	// gl.PushMatrix()
// 	// gl.LoadIdentity()
// 	// gl.Ortho(0, float64(displayWidth), float64(displayHeight), 0, -1, 1)
// 	// gl.MatrixMode(gl.MODELVIEW)
// 	// gl.PushMatrix()
// 	// gl.LoadIdentity()

// 	vertexSize, vertexOffsetPos, vertexOffsetUv, vertexOffsetCol := imgui.VertexBufferLayout()
// 	indexSize := imgui.IndexBufferLayout()

// 	drawType := gl.UNSIGNED_SHORT
// 	const bytesPerUint32 = 4
// 	if indexSize == bytesPerUint32 {
// 		drawType = gl.UNSIGNED_INT
// 	}

// 	// Render command lists
// 	for _, commandList := range drawData.CommandLists() {
// 		vertexBuffer, _ := commandList.VertexBuffer()
// 		indexBuffer, _ := commandList.IndexBuffer()
// 		indexBufferOffset := uintptr(indexBuffer)

// 		gl.VertexPointer(2, gl.FLOAT, int32(vertexSize), unsafe.Pointer(uintptr(vertexBuffer)+uintptr(vertexOffsetPos)))
// 		gl.TexCoordPointer(2, gl.FLOAT, int32(vertexSize), unsafe.Pointer(uintptr(vertexBuffer)+uintptr(vertexOffsetUv)))
// 		gl.ColorPointer(4, gl.UNSIGNED_BYTE, int32(vertexSize), unsafe.Pointer(uintptr(vertexBuffer)+uintptr(vertexOffsetCol)))

// 		for _, command := range commandList.Commands() {
// 			if command.HasUserCallback() {
// 				command.CallUserCallback(commandList)
// 			} else {
// 				clipRect := command.ClipRect()
// 				gl.Scissor(int32(clipRect.X), int32(fbHeight)-int32(clipRect.W), int32(clipRect.Z-clipRect.X), int32(clipRect.W-clipRect.Y))
// 				gl.BindTexture(gl.TEXTURE_2D, uint32(command.TextureID()))
// 				gl.DrawElements(gl.TRIANGLES, int32(command.ElementCount()), uint32(drawType), unsafe.Pointer(indexBufferOffset))
// 			}

// 			indexBufferOffset += uintptr(command.ElementCount() * indexSize)
// 		}
// 	}

// 	// Restore modified state
// 	gl.DisableClientState(gl.COLOR_ARRAY)
// 	gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
// 	gl.DisableClientState(gl.VERTEX_ARRAY)
// 	gl.BindTexture(gl.TEXTURE_2D, uint32(lastTexture))
// 	gl.MatrixMode(gl.MODELVIEW)
// 	gl.PopMatrix()
// 	gl.MatrixMode(gl.PROJECTION)
// 	gl.PopMatrix()
// 	gl.PopAttrib()
// 	gl.PolygonMode(gl.FRONT, uint32(lastPolygonMode[0]))
// 	gl.PolygonMode(gl.BACK, uint32(lastPolygonMode[1]))
// 	gl.Viewport(lastViewport[0], lastViewport[1], lastViewport[2], lastViewport[3])
// 	gl.Scissor(lastScissorBox[0], lastScissorBox[1], lastScissorBox[2], lastScissorBox[3])
// }

// func createDeviceObjects() {
// 	// // Backup GL state
// 	// var lastTexture = gl.GetInteger(gl.TEXTURE_BINDING_2D)
// 	// var lastArrayBuffer = gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
// 	// var lastVertexArray = gl.GetInteger(gl.VERTEX_ARRAY_BINDING)

// 	vertexShader := `#version 150
// uniform mat4 ProjMtx;
// in vec2 Position;
// in vec2 UV;
// in vec4 Color;
// out vec2 Frag_UV;
// out vec4 Frag_Color;
// void main()
// {
// 	Frag_UV = UV;
// 	Frag_Color = Color;
// 	gl_Position = ProjMtx * vec4(Position.xy,0,1);
// }
// `
// 	fragmentShader := `#version 150
// uniform sampler2D Texture;
// in vec2 Frag_UV;
// in vec4 Frag_Color;
// out vec4 Out_Color;
// void main()
// {
// 	Out_Color = vec4(Frag_Color.rgb, Frag_Color.a * texture( Texture, Frag_UV.st).r);
// }
// `
// 	var shaderHandle = gl.CreateProgram()
// 	var vertHandle = gl.CreateShader(gl.VERTEX_SHADER)
// 	var fragHandle = gl.CreateShader(gl.FRAGMENT_SHADER)

// 	glShaderSource := func(handle uint32, source string) {
// 		csource, free := gl.Strs(source + "\x00")
// 		defer free()

// 		gl.ShaderSource(handle, 1, csource, nil)
// 	}

// 	glShaderSource(vertHandle, vertexShader)
// 	glShaderSource(fragHandle, fragmentShader)
// 	gl.CompileShader(vertHandle)
// 	gl.CompileShader(fragHandle)
// 	gl.AttachShader(shaderHandle, vertHandle)
// 	gl.AttachShader(shaderHandle, fragHandle)
// 	gl.LinkProgram(shaderHandle)

// 	var attribLocationTex = gl.GetUniformLocation(shaderHandle, gl.Str("Texture"+"\x00"))
// 	var attribLocationProjMtx = gl.GetUniformLocation(shaderHandle, gl.Str("ProjMtx"+"\x00"))
// 	var attribLocationPosition = gl.GetAttribLocation(shaderHandle, gl.Str("Position"+"\x00"))
// 	var attribLocationUV = gl.GetAttribLocation(shaderHandle, gl.Str("UV"+"\x00"))
// 	var attribLocationColor = gl.GetAttribLocation(shaderHandle, gl.Str("Color"+"\x00"))

// 	gl.GenBuffers(1, &renderer.vboHandle)
// 	gl.GenBuffers(1, &renderer.elementsHandle)
// }

func setStyle() {

	var style = imgui.CurrentStyle()
	// style.WindowRounding = 5.3
	// style.FrameRounding = 2.3
	// style.ScrollbarRounding = 0

	style.SetColor(imgui.StyleColorText, imgui.Vec4{X: 0.90, Y: 0.90, Z: 0.90, W: 0.90})
	style.SetColor(imgui.StyleColorTextDisabled, imgui.Vec4{X: 0.60, Y: 0.60, Z: 0.60, W: 1.00})

	style.SetColor(imgui.StyleColorWindowBg, imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 1.00})
	style.SetColor(imgui.StyleColorChildBg, imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 0.00})
	style.SetColor(imgui.StyleColorPopupBg, imgui.Vec4{X: 0.05, Y: 0.05, Z: 0.10, W: 0.85})

	style.SetColor(imgui.StyleColorBorder, imgui.Vec4{X: 0.70, Y: 0.70, Z: 0.70, W: 0.65})
	style.SetColor(imgui.StyleColorBorderShadow, imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 0.00})

	style.SetColor(imgui.StyleColorFrameBg, imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 0.00})
	style.SetColor(imgui.StyleColorFrameBgHovered, imgui.Vec4{X: 0.90, Y: 0.80, Z: 0.80, W: 0.40})
	style.SetColor(imgui.StyleColorFrameBgActive, imgui.Vec4{X: 0.90, Y: 0.65, Z: 0.65, W: 0.45})

	style.SetColor(imgui.StyleColorTitleBg, imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 1.00})
	style.SetColor(imgui.StyleColorTitleBgCollapsed, imgui.Vec4{X: 0.40, Y: 0.40, Z: 0.80, W: 0.20})
	style.SetColor(imgui.StyleColorTitleBgActive, imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 0.87})

	style.SetColor(imgui.StyleColorMenuBarBg, imgui.Vec4{X: 0.01, Y: 0.01, Z: 0.02, W: 0.80})
	style.SetColor(imgui.StyleColorScrollbarBg, imgui.Vec4{X: 0.20, Y: 0.25, Z: 0.30, W: 0.60})
	style.SetColor(imgui.StyleColorScrollbarGrab, imgui.Vec4{X: 0.55, Y: 0.53, Z: 0.55, W: 0.51})
	style.SetColor(imgui.StyleColorScrollbarGrabHovered, imgui.Vec4{X: 0.56, Y: 0.56, Z: 0.56, W: 1.00})
	style.SetColor(imgui.StyleColorScrollbarGrabActive, imgui.Vec4{X: 0.56, Y: 0.56, Z: 0.56, W: 0.91})
	// style.SetColor(imgui.StyleColorComboBg, imgui.Vec4{X:0.1,Y: 0.1,Z: 0.1,W: 0.99})
	style.SetColor(imgui.StyleColorCheckMark, imgui.Vec4{X: 0.90, Y: 0.90, Z: 0.90, W: 0.83})
	style.SetColor(imgui.StyleColorSliderGrab, imgui.Vec4{X: 0.70, Y: 0.70, Z: 0.70, W: 0.62})
	style.SetColor(imgui.StyleColorSliderGrabActive, imgui.Vec4{X: 1.00, Y: 1.00, Z: 1.00, W: 0.84})
	style.SetColor(imgui.StyleColorButton, imgui.Vec4{X: 0.48, Y: 0.72, Z: 0.89, W: 0.49})
	style.SetColor(imgui.StyleColorButtonHovered, imgui.Vec4{X: 0.50, Y: 0.69, Z: 0.99, W: 0.68})
	style.SetColor(imgui.StyleColorButtonActive, imgui.Vec4{X: 0.80, Y: 0.50, Z: 0.50, W: 1.00})
	style.SetColor(imgui.StyleColorHeader, imgui.Vec4{X: 0.30, Y: 0.69, Z: 1.00, W: 0.53})
	style.SetColor(imgui.StyleColorHeaderHovered, imgui.Vec4{X: 0.44, Y: 0.61, Z: 0.86, W: 1.00})
	style.SetColor(imgui.StyleColorHeaderActive, imgui.Vec4{X: 0.38, Y: 0.62, Z: 0.83, W: 1.00})
	// style.SetColor(imgui.StyleColorColumn, imgui.Vec4{X:0.50,Y: 0.50,Z: 0.50,W: 1.00})
	// style.SetColor(imgui.StyleColorColumnHovered, imgui.Vec4{X:0.70,Y: 0.60,Z: 0.60,W: 1.00})
	// style.SetColor(imgui.StyleColorColumnActive, imgui.Vec4{X:0.90,Y: 0.70,Z: 0.70,W: 1.00})
	style.SetColor(imgui.StyleColorResizeGrip, imgui.Vec4{X: 1.00, Y: 1.00, Z: 1.00, W: 0.00})
	style.SetColor(imgui.StyleColorResizeGripHovered, imgui.Vec4{X: 1.00, Y: 1.00, Z: 1.00, W: 0.00})
	style.SetColor(imgui.StyleColorResizeGripActive, imgui.Vec4{X: 1.00, Y: 1.00, Z: 1.00, W: 0.00})
	// style.SetColor(imgui.StyleColorCloseButton, imgui.Vec4{X:0.50,Y: 0.50,Z: 0.90,W: 0.50})
	// style.SetColor(imgui.StyleColorCloseButtonHovered, imgui.Vec4{X:0.70,Y: 0.70,Z: 0.90,W: 0.60})
	// style.SetColor(imgui.StyleColorCloseButtonActive, imgui.Vec4{X:0.70,Y: 0.70,Z: 0.70,W: 1.00})
	style.SetColor(imgui.StyleColorPlotLines, imgui.Vec4{X: 1.00, Y: 1.00, Z: 1.00, W: 1.00})
	style.SetColor(imgui.StyleColorPlotLinesHovered, imgui.Vec4{X: 0.90, Y: 0.70, Z: 0.00, W: 1.00})
	style.SetColor(imgui.StyleColorPlotHistogram, imgui.Vec4{X: 0.90, Y: 0.70, Z: 0.00, W: 1.00})
	style.SetColor(imgui.StyleColorPlotHistogramHovered, imgui.Vec4{X: 1.00, Y: 0.60, Z: 0.00, W: 1.00})
	style.SetColor(imgui.StyleColorTextSelectedBg, imgui.Vec4{X: 0.00, Y: 0.00, Z: 1.00, W: 0.35})
	style.SetColor(imgui.StyleColorModalWindowDarkening, imgui.Vec4{X: 0.20, Y: 0.20, Z: 0.20, W: 0.35})

	style.SetColor(imgui.StyleColorSliderGrab, imgui.Vec4{X: 0.70, Y: 0.70, Z: 0.70, W: 0.65})

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0.0)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0.0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
	// imgui.PushStyleVarVec2(imgui.StyleVarItemInnerSpacing, imgui.Vec2{X: 0, Y: 0})
	imgui.PushStyleVarFloat(imgui.StyleVarScrollbarRounding, 0.0)
	imgui.PushStyleVarFloat(imgui.StyleVarScrollbarSize, 2.0)
	// imgui.PushStyleVarFloat(imgui.StyleVarGrabMinSize, 0.0)
}
