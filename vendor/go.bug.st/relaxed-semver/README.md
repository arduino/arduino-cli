

# go.bug.st/relaxed-semver [![build status](https://api.travis-ci.org/bugst/relaxed-semver.svg?branch=master)](https://travis-ci.org/bugst/relaxed-semver) [![codecov](https://codecov.io/gh/bugst/relaxed-semver/branch/master/graph/badge.svg)](https://codecov.io/gh/bugst/relaxed-semver)



A library for handling a superset of semantic versioning in golang.

## Documentation and examples

See the godoc here: https://godoc.org/go.bug.st/relaxed-semver

## Semantic versioning specification followed in this library

This library tries to implement the semantic versioning specification [2.0.0](https://semver.org/spec/v2.0.0.html) with an exception: the numeric  format `major.minor.patch` like `1.3.2` may be truncated if a number is zero, so:

   - `1.2.0`  or `1.2.0-beta` may be written as `1.2`  or `1.2-beta` respectively
   - `1.0.0` or `1.0.0-beta` may be written `1` or `1-beta` respectively
   - `0.0.0` may be written as the **empty string**, but `0.0.0-beta` may **not** be written as `-beta`
## Usage

You can parse a semver version string with the `Parse` function that returns a `Version` object that can be used to be compared with other `Version` objects using the `CompareTo`, `LessThan` , `LessThanOrEqual`, `Equal`, `GreaterThan` and `GreaterThanOrEqual` methods.

The `Parse` function returns an `error` if the string does not comply to the above specification. Alternatively the `MustParse` function can be used, it returns only the `Version` object or panics if a parsing error occurs.

## Why Relaxed?

This library allows the use of an even more relaxed semver specification using the `RelaxedVersion` object. It works with the following rules:

- If the parsed string is a valid semver (following the rules above), then the `RelaxedVersion` will behave exactly as a normal `Version` object
- if the parsed string is **not** a valid semver, then the string is kept as-is inside the `RelaxedVersion` object as a custom version string
- when comparing two `RelaxedVersion` the rule is simple: if both are valid semver, the semver rules applies; if both are custom version string they are compared as alphanumeric strings; if one is valid semver and the other is a custom version string the valid semver is always greater

The `RelaxedVersion` object is basically made to allow systems that do not use semver to soft transition to semantic versioning, because it allows an intermediate period where the invalid version is still tolerated.

To parse a `RelaxedVersion` you can use the `ParseRelaxed` function.

## Json parsable

The `Version` and`RelaxedVersion` have the JSON un/marshaler implemented so they can be JSON decoded/encoded.