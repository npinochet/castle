package maps

import "embed"

// IntroFS is the embed.FS for the intro map.
//
//go:embed intro/*.png intro/intro.tmx intro/tiles.tsx
var IntroFS embed.FS
