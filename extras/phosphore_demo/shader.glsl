#define SCALE 2.0

vec2 resolution() {
    return iResolution.xy / SCALE;
}

vec4 texel(sampler2D tex, vec2 uv) {
    //return texture(tex, uv);
    vec2 xy = uv * iChannelResolution[0].xy;
    
    return texelFetch(tex, ivec2(xy), 0);
}

vec3 chromaticAberration(sampler2D tex, vec2 uv, vec2 offset) {
    float r = texel(tex, uv + offset).r;
    float g = texel(tex, uv).g;
    float b = texel(tex, uv - offset).b;
    return vec3(r, g, b);
}

float scanline(vec2 uv) {
    return 0.95 + 0.05 * sin(uv.y * resolution().y * 1.5);
}

float bayerDither4x4(vec2 pos) {
    int x = int(mod(pos.x, 4.0));
    int y = int(mod(pos.y, 4.0));
    int index = x + y * 4;
    float thresholdMatrix[16];
    thresholdMatrix[ 0] =  0.0 / 16.0;
    thresholdMatrix[ 1] =  8.0 / 16.0;
    thresholdMatrix[ 2] =  2.0 / 16.0;
    thresholdMatrix[ 3] = 10.0 / 16.0;
    thresholdMatrix[ 4] = 12.0 / 16.0;
    thresholdMatrix[ 5] =  4.0 / 16.0;
    thresholdMatrix[ 6] = 14.0 / 16.0;
    thresholdMatrix[ 7] =  6.0 / 16.0;
    thresholdMatrix[ 8] =  3.0 / 16.0;
    thresholdMatrix[ 9] = 11.0 / 16.0;
    thresholdMatrix[10] =  1.0 / 16.0;
    thresholdMatrix[11] =  9.0 / 16.0;
    thresholdMatrix[12] = 15.0 / 16.0;
    thresholdMatrix[13] =  7.0 / 16.0;
    thresholdMatrix[14] = 13.0 / 16.0;
    thresholdMatrix[15] =  5.0 / 16.0;
    return thresholdMatrix[index];
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    fragCoord /= SCALE;
    vec2 uv = fragCoord.xy / resolution().xy;

    // === Chromatic Aberration ===
    vec2 aberrationOffset = vec2(1.0 / resolution().x, 0.0) * 1.5;
    vec3 color = chromaticAberration(iChannel0, uv, aberrationOffset);
    
    // === Subtle Glow ===
    float glowStrength = 0.12;
    vec3 blur = vec3(0.0);
    for (float x = -1.0; x <= 1.0; x++) {
        for (float y = -1.0; y <= 1.0; y++) {
            vec2 offset = vec2(x, y) / resolution().xy;
            blur += texture(iChannel0, uv + offset).rgb;
        }
    }
    blur /= 9.0;
    color += blur * glowStrength;
    
    // === Bayer Dithering ===
    float var = 2.0;
    float threshold = bayerDither4x4(fragCoord.xy);
    color = floor(color * var + threshold) / var;
    
    // === Chromatic Aberration ===
    vec2 aberrationOffset2 = vec2(-1.0 / resolution().x, 0.0) * 1.5;
    vec3 color2 = chromaticAberration(iChannel0, uv, aberrationOffset);
    color = mix(color, color2, 0.5);

    // === Scanlines ===
    float scan = scanline(uv);
    color *= scan;
    
    // Output
    fragColor = vec4(color, 1.0);
}

