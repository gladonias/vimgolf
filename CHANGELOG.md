# Changelog

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
