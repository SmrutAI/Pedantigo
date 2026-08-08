# Go Version Behavior Nuances

Cases where Pedantigo's behavior depends on the Go toolchain version used to *build*
your application, not on anything declared in Pedantigo's own `go.mod`.

## `mac` constraint and `net.ParseMAC` (Go 1.26+)

**What changed:** Go's standard library `net.ParseMAC` gained support for parsing
MAC addresses with no separators (e.g. `001A2B3C4D5E`, IEEE EUI-48 base-16 form)
in Go 1.26. See the upstream tracking issue:
[golang/go#66682](https://github.com/golang/go/issues/66682).

**Why this matters for Pedantigo:** the `mac` constraint (used via `validate:"mac"`)
calls `net.ParseMAC` directly. Its behavior is therefore determined by the actual Go
toolchain used to compile *your* application — not by Pedantigo's `go.mod`, and not
by which Pedantigo version you're on.

**This does not affect whether Pedantigo builds or runs.** `go.mod`'s `go 1.21` line
is a minimum toolchain floor, not a behavior guarantee — any toolchain at or above
that floor (1.21, 1.23, 1.25, 1.26, ...) builds and runs Pedantigo identically in
every other respect.

A project on Go 1.21 can depend on, build, and run Pedantigo
in production with no issues. The *only* thing that differs is the validation
*result* of one specific input value on one specific constraint, at request-handling
time — same category of outcome as any other input a user submits that fails
validation, not a build or dependency problem:

| Toolchain used to build your app | `validate:"mac"` on the value `"001A2B3C4D5E"` (no separators) |
|---|---|
| Go 1.21 – 1.25 | Validation fails (treated as an invalid MAC address) |
| Go 1.26+ | Validation passes |

Colon-, hyphen-, and dot-separated MAC formats (`00:1A:2B:3C:4D:5E`,
`00-1A-2B-3C-4D-5E`, `001A.2B3C.4D5E`) are unaffected — they've always been
supported and behave identically across all Go versions Pedantigo supports. If your
application never validates MAC addresses, or never needs to accept the
no-separator form, this nuance has no effect on you at all.

**If you do rely on the `mac` constraint accepting no-separator input**, build your
application with Go 1.26 or later. If you're on an older toolchain, either add
separators to input before validation, or avoid the no-separator MAC form in data
you expect the `mac` constraint to accept.
