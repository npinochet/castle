package maps

import "embed"

// IntroFS is the embed.FS for the intro map.
//
//go:embed intro/*.png intro/*.tsx intro/playground_imp.tmx
var IntroFS embed.FS
