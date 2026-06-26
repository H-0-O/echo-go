# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Phase 1: `ConnectionStatus`, `OnConnectionChange`, expanded `Config` (`Auth`, `UserAuthentication`, `BearerToken`, `CSRFToken`, `AutoConnect`, `Cluster`, `Namespace`)
- Phase 2: `Leave`, `LeaveChannel`, `LeaveAllChannels`, `Join`, `Listen` shorthand; channel registry prefixed keys
- Phase 3: `Subscribed`, `Error`, `Notification`, `ListenToAll`, optional-callback `StopListening` variants; presence payload normalization (`Here`, `Joining`, `Leaving`)
- Phase 4: `Signin()` and `UserAuthentication` wiring for Pusher user auth
- Phase 5: `EncryptedPrivate` for `private-encrypted-` channels
- Phase 6: Complete null broadcaster types; `Config.Connector` for custom backends; exported `Connector` interface
- Phase 7: Reverb integration test harness (`internal/integration/`), worker and HTTP examples, README parity table, CI workflow

### Changed

- `New` returns an error for unknown broadcasters (no silent null fallback)
- `AutoConnect` defaults to `true` to match Laravel Echo constructor behavior
- `ListenForWhisper` uses `.client-{event}` through the event formatter (leading `.` skips namespace)

### Fixed

- Channel registry cache key collision for private/presence/encrypted channels (prefixed keys)

### Breaking

- `New(Config) (*Echo, error)` — callers must handle the error return value
- Unknown `Broadcaster` values no longer fall back to the null connector
- `Channel` interface expanded with Phase 3 methods (`Subscribed`, `Error`, `Notification`, `ListenToAll`, optional-callback stop methods)
- Deprecated top-level `AuthEndpoint` / `AuthHeaders` in favor of `Auth.Endpoint` / `Auth.Headers` (still migrated in `New`)
