// Package recty provides an OpenGL renderer of coloured squares. Support for
// textures is planned.
package recty

import (
    "errors"
    "github.com/go-gl/gl"
    "github.com/snorredc/gome"
)

// vertSource is the source code for the vertex shader.
const vertSource = `#version 150 core

in vec4 rect;
in vec4 color;
in vec2 texcoord;

out vec4 vColor;
out vec2 vTexcoord;

void main() {
    // tell OpenGL this is our vertex's location
    gl_Position = rect;
    vColor = color;
    vTexcoord = texcoord;
}
`

// geomSource is the source code for the geometry shader.
const geomSource = `#version 150 core

layout(points) in;
layout(triangle_strip, max_vertices = 4) out;

in vec4 vColor[];
in vec2 vTexcoord[];

uniform mat3 transform;

out vec4 fColor;
out vec2 fTexcoord;

void main() {
    fColor = vColor[0];
    fTexcoord = vTexcoord[0];
    vec4 rect = gl_in[0].gl_Position;
    
    gl_Position = vec4(transform * vec3(rect.xy, 1), 1.0);
    EmitVertex();
    gl_Position = vec4(transform * vec3(rect.xw, 1), 1.0);
    EmitVertex();
    gl_Position = vec4(transform * vec3(rect.zy, 1), 1.0);
    EmitVertex();
    gl_Position = vec4(transform * vec3(rect.zw, 1), 1.0);
    EmitVertex();
}
`

// fragSource is the source code for the fragment shader.
const fragSource = `#version 150 core

precision highp float;

in vec4 fColor;
in vec2 fTexcoord;

// fragColor is the colour of each point.
out vec4 outColor;

uniform sampler2D tex;

void main() {
    outColor = fColor + texture(tex, fTexcoord);
}
`

// Recty is a rendering object and context.
type Recty struct {
    Program gl.Program
    vao     gl.VertexArray
    vbo     gl.Buffer

    Transform gl.UniformLocation
}

// Init initialises the renderer. OpenGL should be initialised before calling
// Init.
func (recty *Recty) Init() error {
    recty.vao = gl.GenVertexArray()
    recty.vao.Bind()

    // set the shaders and program up
    vertShader := gl.CreateShader(gl.VERTEX_SHADER)
    geomShader := gl.CreateShader(gl.GEOMETRY_SHADER)
    fragShader := gl.CreateShader(gl.FRAGMENT_SHADER)
    defer vertShader.Delete()
    defer geomShader.Delete()
    defer fragShader.Delete()
    vertShader.Source(vertSource)
    geomShader.Source(geomSource)
    fragShader.Source(fragSource)
    vertShader.Compile()
    geomShader.Compile()
    fragShader.Compile()

    recty.Program = gl.CreateProgram()
    recty.Program.AttachShader(vertShader)
    recty.Program.AttachShader(geomShader)
    recty.Program.AttachShader(fragShader)
    // activate the program
    recty.Program.Link()
    recty.Program.Use()

    recty.vbo = gl.GenBuffer()
    recty.vbo.Bind(gl.ARRAY_BUFFER)

    attrRect := recty.Program.GetAttribLocation("rect")
    attrRect.AttribPointer(4, gl.FLOAT, false, 10*4, uintptr(0))
    attrRect.EnableArray()
    attrColor := recty.Program.GetAttribLocation("color")
    attrColor.AttribPointer(4, gl.FLOAT, false, 10*4, uintptr(4*4))
    attrColor.EnableArray()
    attrTexcoord := recty.Program.GetAttribLocation("texcoord")
    attrTexcoord.AttribPointer(2, gl.FLOAT, false, 10*4, uintptr(8*4))
    attrTexcoord.EnableArray()

    if err := gome.GetError(); err != nil {
        return errors.New(recty.Program.GetInfoLog())
    }
    // get handles for aMax and aMin.
    recty.Transform = recty.Program.GetUniformLocation("transform")
    recty.SetTransform(
        1, 0, 0,
        0, 1, 0,
    )
    return gome.GetError()
}

// SetTransform sets the transformation matrix. The arguments are the first to
// rows of the matrix, where the last is set to [0 0 1] to prevent 3D results.
func (recty *Recty) SetTransform(a, d, g, b, e, h float32) {
    recty.Transform.UniformMatrix3f(false, &[9]float32{a, b, 0, d, e, 0, g, h, 1})
}

// SetScale is a utility for setting the transformation matrix. It scales the
// matrix horizontally by w and vertically by h and offsets the result by
// (dx, dy). The offset is in OpenGL coordinates, that is in [-1, 1].
func (recty *Recty) SetScale(w, h, dx, dy float32) {
    recty.SetTransform(w, 0, dx, 0, h, dy)
}

// Draw draws rectangles directly to the screen. Each rectangle is represented
// as
//
//     []float32{x1, y1, x2, y2, r, g, b, a, tx, ty}
//
// where (x1, y1) is the lower left corner and (x2, y2) is the upper right one,
// and (r, g, b, a) is the RGBA colour. Since textures are not implemented yet
// tx and ty are unused.
func (recty *Recty) Draw(rects ...[10]float32) {
    recty.vbo.Bind(gl.ARRAY_BUFFER)
    gl.BufferData(gl.ARRAY_BUFFER, 10*4*len(rects), rects, gl.STATIC_DRAW)
    gl.DrawArrays(gl.POINTS, 0, len(rects))
}

// Delete deletes the Recty freeing any related resources.
func (recty *Recty) Delete() {
    recty.Program.Delete()
    recty.vao.Delete()
}
