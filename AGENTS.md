# AGENTS.md

This file gives coding agents enough context to work safely on `nosleepp`.

## Project Summary

`nosleepp` is a Go CLI that keeps a computer awake while AI coding agents are actively working. It is built for tools such as Codex, Claude Code, OpenCode, Antigravity, and Cursor.

The key behavior is activity-aware: an agent app being open is not enough. `nosleepp` samples process activity and only treats an agent as working when there is measurable CPU activity or descendant-process activity.

Primary commands:

```bash
nosleepp list
nosleepp list --all
nosleepp watch
nosleepp watch --once
nosleepp power-test --duration 2m
nosleepp version
```

The npm package is also named `nosleepp` and installs a global `nosleepp` command:

```bash
npm install -g nosleepp
nosleepp watch
```

## Repository Layout

```txt
cmd/                    Cobra command definitions and CLI tests
internal/agent/         Agent profiles, process matching, activity detection
internal/defaults/      Shared defaults for interval/sample/quiet/thresholds
internal/process/       Platform process scanners
internal/power/         Platform sleep-prevention managers
internal/watch/         Watcher lifecycle and lock/release behavior
npm/                    Node launcher and npm package metadata
scripts/                Cross-platform npm binary build scripts
```

Prefer the existing package boundaries. CLI code should parse options and print output. Watcher code should coordinate scanning and power locking. Platform details belong in `internal/process` or `internal/power`.

## Platform Behavior

Windows:

- Process scanning uses Windows APIs.
- Sleep prevention uses `SetThreadExecutionState` with system-required semantics.
- Default CPU threshold is `250ms`.

macOS:

- Process scanning uses `ps`.
- Sleep prevention starts `caffeinate -i` and kills/waits it on release.
- Default CPU threshold is `250ms`.

Linux:

- Process scanning reads `/proc/[pid]/stat` and `/proc/[pid]/comm`.
- Sleep prevention starts `systemd-inhibit --what=idle:sleep --mode=block sleep infinity`.
- Release kills the process group so the child `sleep infinity` does not survive.
- Default CPU threshold is `10ms`, because Linux/OpenCode activity can be lighter than the Windows/macOS default.

Unsupported platforms should fail clearly rather than silently pretending to hold a sleep lock.

## Defaults

Source of truth is `internal/defaults/defaults.go`. Current defaults:

```txt
interval: 10s
sample: 5s
quiet: 3m
power-test duration: 2m
cpu threshold: 250ms on Windows/macOS, 10ms on Linux
```

Keep README, npm README, help text, and tests aligned with this package.

## Activity Detection

`nosleepp` takes two process snapshots separated by the sample window. An agent is marked `working` when either:

- the agent process or a descendant gains at least the CPU threshold during the sample window, or
- descendant processes appear or disappear during the sample window.

`nosleepp list` hides idle matches by default. `nosleepp list --all` includes idle/open matching agents.

The default agent profiles live in `internal/agent/agent.go`. Matching is case-insensitive and strips `.exe`.

## Important WSL Note

WSL is useful for testing the Linux binary, process scanning, and basic command behavior. It is not a reliable proof that Linux sleep prevention can block the Windows host from sleeping. If the physical machine is Windows, the Windows binary must hold the Windows sleep lock.

A future hybrid mode may detect Linux processes inside WSL while holding the Windows host lock, but that is not currently implemented.

## npm Package

The Node launcher in `npm/index.js` chooses a bundled binary using `process.platform` and `process.arch`. Supported npm targets include:

```txt
win32-x64 -> nosleepp-win32-x64.exe
darwin-arm64 -> nosleepp-darwin-arm64
darwin-x64 -> nosleepp-darwin-x64
linux-x64 -> nosleepp-linux-x64
linux-arm64 -> nosleepp-linux-arm64
```

The executable-permission helper applies to macOS and Linux binaries. Keep `npm/package.json` `files` in sync with any bundled binary changes.

## Test And Build Commands

Use these from the repo root:

```bash
go test ./...
npm test --prefix ./npm
./scripts/build-npm.sh 0.3.0
npm --prefix ./npm pack --dry-run
```

On Windows/PowerShell, the equivalent release build is:

```powershell
.\scripts\build-npm.ps1 -Version 0.3.0
```

Generated binaries in `npm/bin/` are ignored except `.gitkeep`. Do not commit generated binaries unless the release process intentionally changes.

## Manual Smoke Tests

Detection:

```bash
nosleepp list
nosleepp list --all
nosleepp watch --once
```

Power lock:

```bash
nosleepp power-test --duration 2m
```

Linux OpenCode sensitivity check:

```bash
nosleepp list --agent "OpenCode=opencode" --sample 5s
nosleepp list --agent "OpenCode=opencode" --sample 5s --cpu-threshold 10ms
```

## Development Notes

- Use `gofmt` on Go changes.
- Keep platform-specific code behind Go build tags.
- Prefer small interfaces and fake scanners/power managers in tests.
- Do not make Windows/macOS default threshold `10ms`; that value is Linux-specific.
- `--config` is accepted but not implemented yet.
- `power-test` is for proving the OS sleep lock path independently from agent detection.
