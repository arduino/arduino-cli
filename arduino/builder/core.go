package builder

import "github.com/arduino/go-paths-helper"

// CoreBuildCachePath fixdoc
func (b *Builder) CoreBuildCachePath() *paths.Path {
	return b.coreBuildCachePath
}
