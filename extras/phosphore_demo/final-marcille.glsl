// glsl.app version
#version 300 es

precision highp float;
precision highp sampler2D;

in vec2 uv;
out vec4 out_color;

uniform vec2 u_resolution;
uniform float u_time;
uniform vec4 u_mouse;
uniform sampler2D u_textures[16];

#pragma region START

#define SCALE 4.0

uniform vec2 uDisplaySize;
uniform float uAspect;
uniform sampler2D uMaskSampler;
uniform float uMaskBrightness;
uniform float uDoBGR;
uniform float uVerticalScan;
uniform float uScanIntensity;
uniform float uWarp;
uniform float uGlowAmount;
uniform float uDoIntScale;
uniform float uInterlaceTick;
uniform float uDoInterlace;
uniform float uDeconverge;
uniform float uHalation;
uniform sampler2D uNoiseSampler;
uniform sampler2D uBackimg;
uniform float uDoBackimg;
uniform float uZoom;

vec2 resolution() {
    return u_resolution / SCALE;
}

vec4 quinticTexture( sampler2D sampler, vec2 p ) {
  ivec2 size = textureSize(sampler, 0);

  vec2 res = resolution();
  p.x -= mod(p.x, 1.0 / res.x);
  p.y -= mod(p.y, 1.0 / res.y);
  p.y = 1. - p.y;
  vec2 coords = p * res * SCALE;
  coords.y = float(size.y) - coords.y;

  return texelFetch(sampler, ivec2(coords), 0);
}

void applyMask( inout vec4 color) {
	/// Gets a color from the phosphor mask and brightens the shadowed pragma regions if desired
	/// Multiplies the resulting color into the output
  //return;
	
	// Repeat for as many textures fit in the display window
  vec2 uDisplaySize = resolution();
  float uMaskScale = 0.5;
	vec2 maskUV = uv * ( uDisplaySize / 256.0) * ( 1.0 / uMaskScale); //* vec2(1.0 , 5.0 / (uDisplaySize.y / resolution().y));
	
	// Repeat on non-power-of-two intervals to match the mask texture's odd interval
	//maskUV.x = mod(maskUV.x, 1.0 - ((uDoBGR / 256.0) * 196.0));
	//maskUV.y = mod(maskUV.y, 1.0 - ((1.0 / 256.0) * 226.0));
	
	// Get color from the mask texture
	vec4 maskColor = texture( u_textures[2], maskUV);
	
	// Brighten the black parts of the mask to improve overall brightness
  float uMaskBrightness = 0.8;
	maskColor.rgb += color.rgb * uMaskBrightness;
	
	// Apply the result
	color *= maskColor;
}

void applyScanlines( inout vec4 color, in vec2 uv) {
	/// Gets the brightness of the current fragment with respect to scanlines
	/// Multiplies the resulting brightness into the output
	
	// Linear sawtooth wave applied either horizontally or vertically
  float uVerticalScan = 0.0;
  float uScanIntensity = 0.4;
  float uZoom = 1.0;
	float scanLum = ( mod( uv.x, 1.0 / resolution().x) * resolution().x * uVerticalScan) + ( mod( uv.y * uZoom, 1.0 / resolution().y) * resolution().y * ( 1.0 - uVerticalScan));
	
	// Turn the sawtooth into a triangle wave, smoothen it, and scale by the intensity uniform
	scanLum = ( smoothstep( 0.0, 0.5, abs( scanLum - 0.5)) * ( -2.0 * uScanIntensity)) + 1.0;
	
	// Apply the result
	color.rgb *= scanLum;
}

void applyGlow( inout vec4 color, in vec2 uv) {
	/// Applies a short-range bloom effect to bleed bright pixels into the surrounding area
	/// Should typically be applied after scanlines and the phosphor mask
	
	// See if the fragment is on the border
	bool is_border = ( (uv.x > 1.0) || (uv.x < 0.0) || (uv.y > 1.0) || (uv.y < 0.0) );
	
  float uGlowAmount = 0.3;
	// Only do the expensive blur effect if it will be seen
	if (true) {
			// Save the size of a texel into a vec2
			vec2 texelSize = 1.0 / resolution();
			
			// Accumulate samples from the surrounding texels
			vec4 glowColor = quinticTexture( u_textures[0], uv);
			// Diagonals
			glowColor += quinticTexture( u_textures[0], uv + vec2( -texelSize.x, -texelSize.y));
			glowColor += quinticTexture( u_textures[0], uv + vec2( -texelSize.x,  texelSize.y));
			glowColor += quinticTexture( u_textures[0], uv + vec2(  texelSize.x, -texelSize.y));
			glowColor += quinticTexture( u_textures[0], uv + vec2(  texelSize.x,  texelSize.y));
			// Cardinals
			glowColor += quinticTexture( u_textures[0], uv + vec2( 0.0, -texelSize.y));
			glowColor += quinticTexture( u_textures[0], uv + vec2( 0.0,  texelSize.y));
			glowColor += quinticTexture( u_textures[0], uv + vec2(  -texelSize.x, 0.0));
			glowColor += quinticTexture( u_textures[0], uv + vec2(  texelSize.x,  0.0));
			// Reduce the values back down to 0-1
			glowColor *= 0.11111111;
	
			// Add the result to the output
			color.rgb += glowColor.rgb * uGlowAmount;
		}
}

vec4 getColor( in vec2 uv) {
	/// Gets a base color from the game and applies color corrections
	
	// Get a starting color using better-than-bilinear filtering
	vec4 col = quinticTexture( u_textures[0], uv);
	
	// Offset the UVs for the red and blue channels to simulate deconvergence
	col.r = quinticTexture( u_textures[0], uv - ( 0.5 / resolution())).r;
	col.b = quinticTexture( u_textures[0], uv + ( 0.5 / resolution())).b;
	
	// Calculate a grayscale color, keeping luminosity relatively constant
	float grayf = (0.2989 * col.r) + (0.5870 * col.g) + (0.1140 * col.b);
	vec4 gray = vec4( grayf, grayf, grayf, 1.0);
	
	// Return the final color
	return mix(col, gray, 0.4);
}

void main() {
	vec4 color = getColor(uv);
	
	// Apply the phosphor mask
	applyMask( color);
	
	// Apply scanlines
	applyScanlines( color, uv);
	
	// Apply glow
	applyGlow( color, uv);
	
  out_color = color;
  return;
	
  out_color = color;
}

