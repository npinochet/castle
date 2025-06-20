// https://www.shadertoy.com/view/4ss3W7
// https://glsl.app#~c49awpep2dn33nx4s7m3fl33

#version 300 es

precision highp float;
precision highp sampler2D;

in vec2 uv;
out vec4 out_color;

uniform vec2 u_resolution;
uniform float u_time;
uniform vec4 u_mouse;
uniform sampler2D u_textures[16];

#define ENABLE_LIGHTING

#define OFFSET_X 1
#define OFFSET_Y 1
#define DEPTH	 5.5

vec3 texsample(const int x, const int y) {
    vec2 textureSize2d = vec2(textureSize(u_textures[0], 0));
	vec2 uve = uv + (vec2(x, y) / textureSize2d.xy);

	return texture(u_textures[0], uve).xyz;
}

float luminance(vec3 c) {
	return dot(c, vec3(.2126, .7152, .0722));
}

vec3 normal(in vec2 fragCoord) {
	float R = abs(luminance(texsample(OFFSET_X, 0)));
	float L = abs(luminance(texsample(-OFFSET_X, 0)));
	float D = abs(luminance(texsample(0, OFFSET_Y)));
	float U = abs(luminance(texsample(0, -OFFSET_Y)));
				 
	float X = (L-R) * .5;
	float Y = (U-D) * .5;

	return normalize(vec3(X, Y, 1. / DEPTH));
}

void main() {
    vec3 n = normal(uv);

#ifdef ENABLE_LIGHTING
    vec2 textureSize2d = vec2(textureSize(u_textures[0], 0));
    vec2 mouse = vec2(u_mouse.x, u_resolution.y-u_mouse.y);
	vec3 lp = vec3(mouse / u_resolution.xy * textureSize2d.xy, 200.);
	vec3 sp = vec3((u_resolution * uv).xy / u_resolution.xy * textureSize2d.xy, 0.);
	
	vec3 c = texsample(0, 0) * dot(n, normalize(lp - sp));
	
#else
	vec3 c = n;
#endif /* ENABLE_LIGHTING */
	out_color = vec4(c, 1);
}
