// Package version holds the Centralizer release identity.
package version

// Version is the SemVer identifier for this build. Release automation
// may overwrite this via -ldflags.
const Version = "0.1.1"

// Protocol is the Centralizer Protocol major.minor spoken by this build.
const Protocol = "1.0"

// ProtocolMajor is used during handshake negotiation.
const ProtocolMajor = 1
