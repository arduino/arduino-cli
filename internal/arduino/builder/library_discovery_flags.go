// This file is part of arduino-cli.
//
// Copyright (C) Arduino s.r.l. and/or its affiliated companies
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html

package builder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	semver "go.bug.st/relaxed-semver"
)

var (
	invalidIdentifierChars = regexp.MustCompile(`[^A-Za-z0-9_]+`)
	leadingDigit           = regexp.MustCompile(`^[0-9]`)
)

// librarySlug converts a library name into a valid uppercase C identifier.
// Runs of characters outside [A-Za-z0-9_] collapse to a single "_"; the
// result is uppercased, and a leading "_" is added if it would otherwise
// start with a digit.
func librarySlug(name string) string {
	slug := invalidIdentifierChars.ReplaceAllString(name, "_")
	slug = strings.ToUpper(slug)
	if leadingDigit.MatchString(slug) {
		slug = "_" + slug
	}
	return slug
}

var versionComponents = regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// encodeLibraryVersion packs a library version into a 24-bit
// major/minor/patch value (0xMMmmpp). A nil version, or a version string
// with no leading numeric major component, encodes to 1 (the lowest
// non-zero encoded value, i.e. 0.0.1) per the "unknown version" convention.
// Each component is clamped to [0, 255] so it can't overflow into the next
// byte. Minor/patch default to 0 when absent (e.g. "1" or "1.2" are valid).
func encodeLibraryVersion(v *semver.Version) uint32 {
	if v == nil {
		return 1
	}
	m := versionComponents.FindStringSubmatch(v.String())
	if m == nil {
		return 1
	}
	major := clampVersionComponent(m[1])
	minor := clampVersionComponent(m[2])
	patch := clampVersionComponent(m[3])
	return uint32(major)<<16 | uint32(minor)<<8 | uint32(patch)
}

// clampVersionComponent parses a numeric version component (possibly
// empty, meaning "absent") and clamps it to [0, 255]. The uint8 return
// type makes that bound visible to the type system, so callers can widen
// it to a larger integer type without any risk of overflow.
func clampVersionComponent(s string) uint8 {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n > 255 {
		return 255
	}
	if n < 0 {
		return 0
	}
	return uint8(n)
}

// librarySourceSuffix maps a library's discovery location to the macro
// suffix identifying where it was found. Returns "" for any location with
// no defined suffix (currently none — every known LibraryLocation value is
// mapped — but this keeps the function total instead of panicking on a
// future enum addition).
func librarySourceSuffix(location libraries.LibraryLocation) string {
	switch location {
	case libraries.User:
		return "IN_SKETCHBOOK"
	case libraries.PlatformBuiltIn, libraries.ReferencedPlatformBuiltIn:
		return "IN_PLATFORM"
	case libraries.Profile:
		return "IN_PROFILE"
	case libraries.IDEBuiltIn:
		return "IN_IDE"
	case libraries.Unmanaged:
		return "IN_SPECIFIED_PATH"
	default:
		return ""
	}
}

// buildLibraryDiscoveryFlags renders the discovery macros for each library
// in libs, joined by spaces, in libs' existing order: a
// "-DFOUND_<SLUG>_LIB=0x<hex>" version macro, plus a
// "-DFOUND_<SLUG>_LIB_<SUFFIX>=1" macro identifying where the library was
// found (omitted if librarySourceSuffix has no mapping for its location).
func buildLibraryDiscoveryFlags(libs libraries.List) string {
	flags := make([]string, 0, len(libs))
	for _, lib := range libs {
		slug := librarySlug(lib.Name)
		flags = append(flags, fmt.Sprintf("-DFOUND_%s_LIB=0x%06X", slug, encodeLibraryVersion(lib.Version)))
		if suffix := librarySourceSuffix(lib.Location); suffix != "" {
			flags = append(flags, fmt.Sprintf("-DFOUND_%s_LIB_%s=1", slug, suffix))
		}
	}
	return strings.Join(flags, " ")
}
