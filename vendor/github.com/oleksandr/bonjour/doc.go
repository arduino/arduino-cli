// bonjour
//
// This is a simple Multicast DNS-SD (Apple Bonjour) library written in Golang.
// You can use it to discover services in the LAN. Pay attention to the
// infrastructure you are planning to use it (clouds or shared infrastructures
// usually prevent mDNS from functioning). But it should work in the most
// office, home and private environments.
//
// **IMPORTANT**: It does NOT pretend to be a full & valid implementation of
// the RFC 6762 & RFC 6763, but it fulfils the requirements of its authors
// (we just needed service discovery in the LAN environment for our IoT
// products). The registration code needs a lot of improvements. This code was
// not tested for Bonjour conformance but have been manually verified to be
// working using built-in OSX utility `/usr/bin/dns-sd`.
//
package bonjour
