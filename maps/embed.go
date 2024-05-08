package maps

import "embed"

// IntroFS is the embed.FS for the intro map.
//
//go:embed intro/intro.tmx intro/*.png
var IntroFS embed.FS
