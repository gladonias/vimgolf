# Changelog

## [0.1.1] - 2026-04-24

### Fixed
- Validate server's `client` version field on challenge download to prevent playing incompatible challenges.
- Accept underscores in API keys to match Ruby client's validation (`\w` includes `_`).
- Parse JSON response body on upload and check `status` field instead of relying on HTTP status code alone.
- Use multipart form encoding for upload to preserve binary keylog data (bytes >0x80 were corrupted by URL encoding).
- Return an error on invalid key during `vimgolf setup` instead of silently succeeding.
- Replace `os.Exit(1)` in `vimgolf local` with proper error return so deferred cleanup runs.
- Add 30-second HTTP timeout to all network requests (challenge download, upload, leaderboard).

## [0.1.0] - 2026-04-24

### Added
- Go CLI client under `cli/` — a single-binary alternative to the Ruby gem.
- `vimgolf setup` — configure API key and username.
- `vimgolf put <id>` — download and play a challenge from vimgolf.com.
- `vimgolf local <infile> <outfile>` — play a local challenge without uploading.
- `vimgolf leaderboard <id>` (alias `lb`) — display the top 10 leaderboard entries and your rank.
- `vimgolf version` — print client version.
- Full port of the Vim binary keylog parser (single-byte, multi-byte 0x80-prefixed keycodes, KE_SNIFF compatibility).
- Embedded `vimgolf.vimrc` via `go:embed` — no external files needed.
- Interactive play loop with retry, diff, and upload prompts.
- Support for `GOLFVIM`, `GOLFSHOWDIFF`, `GOLFHOST`, and `GOLFDEBUG` environment variables.
