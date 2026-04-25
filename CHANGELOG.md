# Changelog

## [0.3.1] - 2026-04-25

### Fixed
- Check all multipart write errors during upload instead of silently discarding them.
- Move `bestScoreRe` and `rankRe` to package-level vars (compiled once, not on every call).
- Decode HTML entities in challenge titles and player names (e.g., `&#39;` now renders as `'`).
- Exit with a clear error if the home directory cannot be determined.
- Trim whitespace from `GOLFVIM`/`GOLFSHOWDIFF` env vars to prevent panic on whitespace-only values.
- Clean up orphaned files on partial challenge download failure.

### Added
- `--reset` flag for `next` command to restart from the oldest challenge.
- Tests for the `upload` function in the cmd package (4 new tests, 39 total).

### Changed
- Removed unnecessary `fmt.Sprintf` wrappers on string literals.

## [0.3.0] - 2026-04-25

### Added
- `vimgolf next` command: play challenges sequentially from oldest to newest, repeating each until you match the #1 score.
- Persistent progress tracking for `next` via `~/.vimgolf/next.yaml`.
- `NextState` load/save helpers in the config package.

### Changed
- `vimgolf setup` now preserves existing key/username when re-running (press Enter to keep current values).
- Setup username prompt clarifies it should be the @handle shown on vimgolf.com after sign-in.
- Updated README with `next` command docs, upstream Ruby repo link, and AI assistant attribution.

## [0.2.0] - 2026-04-25

### Added
- Random challenge selection: `vimgolf put` with no arguments picks a random unplayed challenge.
- `--min-entries` flag for `put` command to filter challenges by minimum entry count (default: 100).
- New `listing` package to scrape challenge listings from vimgolf.com with pagination support.
- Tests for the listing package (8 new tests, 34 total).

### Changed
- `put` command now accepts 0 or 1 arguments (was exactly 1).
- Exported `FetchUserEntry` from the leaderboard package for reuse.

## [0.1.3] - 2026-04-24

### Added
- `--version` flag on the root command (`vimgolf --version`).
- Tests for `fetchUserEntry` fallback path and user-not-found case (26 tests total).

### Fixed
- Upload-and-quit (`[x]`) now loops back to the prompt if upload fails, matching Ruby behavior.
- Show diff (`[d]`) reports errors to the user instead of silently ignoring them.

## [0.1.2] - 2026-04-24

### Added
- Test suites for config, challenge, leaderboard, and UI packages (24 tests total).
- Challenge ID validation in `put` and `leaderboard` commands.

### Fixed
- Leaderboard parser uses a single unified regex per entry instead of two separate lists that could desync.
- `AskChoice` extracts valid letters dynamically from the `[x]` prefix in choices instead of a hardcoded set.
- Download errors now include the server response body (truncated to 200 chars) for better diagnostics.
- Temp file cleanup logs a debug warning on failure instead of silently ignoring errors.

### Changed
- Set `go.mod` minimum Go version to 1.21 (was incorrectly set to a future version).

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
