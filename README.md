# nosleepp

`nosleepp` keeps your computer awake while AI agents are actively working.

It is built for agent tools such as Codex, Claude Code, OpenCode, Antigravity, and Cursor. It does not treat an app as working just because it is open. Instead, it samples process activity and reports agents that show measurable CPU or child-process activity.

## Supported Platforms

| Platform            | Status              | Sleep prevention                 |
| ------------------- | ------------------- | -------------------------------- |
| Windows x64         | Supported           | Native `SetThreadExecutionState` |
| macOS Apple Silicon | Supported           | Built-in `caffeinate -i`         |
| macOS Intel         | Supported           | Built-in `caffeinate -i`         |
| Linux               | Not supported in v2 | Planned for a later version      |

`nosleepp` prevents system idle sleep. It does not force the display to stay awake.

## Install

### Go Install

```bash
go install github.com/peterintech/nosleepp@latest
nosleepp watch
```

### npm

```bash
npm install -g nosleepp
nosleepp watch
```

### Local Build

Build and run directly from this repository:

```powershell
go build -o nosleepp.exe .
.\nosleepp.exe watch
```

On macOS:

```bash
go build -o nosleepp .
./nosleepp watch
```

## Quick Start

Check whether any agent is actively working:

```bash
nosleepp list
```

Keep the computer awake while agents are working:

```bash
nosleepp watch
```

Show open agents too, even if they are idle:

```bash
nosleepp list --all
```

Check once, then exit with a status code:

```bash
nosleepp watch --once
```

## How Working Detection Works

`nosleepp` takes two process snapshots separated by a sample window. The default sample window is `2s`.

An agent is marked `working` when either:

- the agent process or one of its descendant processes gains at least `250ms` of CPU time during the sample window, or
- a descendant process appears or disappears during the sample window.

An agent is marked `idle` when the process is open but no meaningful activity is detected.

By default, `list` only shows `working` agents. Use `--all` to include `idle` agents.

## Default Agents

`nosleepp` detects these agent process names by default:

| Agent       | Process names      |
| ----------- | ------------------ |
| Codex       | `codex`, `Codex`   |
| Claude Code | `claude`           |
| OpenCode    | `opencode`         |
| Antigravity | `antigravity`      |
| Cursor      | `Cursor`, `cursor` |

Matching is case-insensitive, and `.exe` is ignored.

## Commands & Flags

### `nosleepp list`

Lists agents that are actively working.

```bash
nosleepp list
nosleepp list --all
nosleepp list --json
nosleepp list --sample 5s --cpu-threshold 100ms
```

Default behavior:

- samples activity for `2s`
- prints only `working` agents
- prints `No working agents found.` if matching agent apps are open but idle
- does not change power state

Text output columns:

| Column      | Meaning                                                |
| ----------- | ------------------------------------------------------ |
| `AGENT`     | Human-readable agent name                              |
| `PID`       | Process ID                                             |
| `PROCESS`   | Executable name                                        |
| `STATUS`    | `working` or `idle`                                    |
| `CPU_DELTA` | CPU time gained during the sample                      |
| `CHILDREN`  | Number of descendant processes                         |
| `EVIDENCE`  | Why it was marked working, such as `cpu` or `children` |

### `nosleepp watch`

Watches for working agents and prevents system idle sleep while work is active.

```bash
nosleepp watch
nosleepp watch --interval 5s --quiet 1m
nosleepp watch --once
```

Default behavior:

- samples activity for `2s`
- polls every `10s`
- prevents system idle sleep when working agents are found
- keeps the computer awake for `30s` after activity stops
- releases the no-sleep lock and exits after the quiet window
- restores normal sleep behavior on `Ctrl+C`

`--once` checks once and exits. Useful for scripts. Exits `0` when at least one working agent is found, `1` when none are found.

### `nosleepp version`

Prints version/build information.

### Global Flags

| Flag       | Description                                                                                         |
| ---------- | --------------------------------------------------------------------------------------------------- |
| `--agent`  | Add or override an agent signature. Format: `name=process1,process2`. Can be passed more than once. |
| `--config` | Reserved for future config-file support. Accepted but not implemented yet.                          |

### `list` Flags

| Flag              | Default | Description                                                                 |
| ----------------- | ------- | --------------------------------------------------------------------------- |
| `--all`           | off     | Show idle/open matching agents as well as working agents                    |
| `--json`          | off     | Print machine-readable JSON                                                 |
| `--sample`        | `2s`    | How long to wait between snapshots. Longer reduces false negatives.         |
| `--cpu-threshold` | `250ms` | CPU activity required to mark an agent as working. Lower is more sensitive. |

### `watch` Flags

| Flag              | Default | Description                                                  |
| ----------------- | ------- | ------------------------------------------------------------ |
| `--interval`      | `10s`   | How often to check for activity                              |
| `--sample`        | `2s`    | Activity sample window used on each poll                     |
| `--cpu-threshold` | `250ms` | CPU activity required during the sample window               |
| `--quiet`         | `30s`   | How long to stay awake after the last detected activity      |
| `--once`          | off     | Check once and exit (`0` if working agent found, `1` if not) |
| `--json`          | off     | Print match details as JSON when state changes               |

## Exit Codes

| Code | Meaning                                                                 |
| ---- | ----------------------------------------------------------------------- |
| `0`  | Success, or `watch --once` found a working agent                        |
| `1`  | `watch --once` found no working agents                                  |
| `2`  | Invalid CLI usage or agent flag                                         |
| `3`  | Failed to acquire/release platform power state or watcher runtime error |

## Common Workflows

### Keep The Computer Awake While Codex Works

```bash
nosleepp watch
```

Leave this running in a terminal. When Codex is actively working, system idle sleep is blocked. When work stops and the quiet window expires, `nosleepp` releases the lock and exits.

### Check If Anything Is Working Right Now

```bash
nosleepp list
```

If it prints `No working agents found.`, your agent apps may still be open, but they are not currently doing observable work.

### Debug Agent Detection

```bash
nosleepp list --all
```

This shows matching agents even when idle. If your agent does not appear here, add a custom process signature with `--agent`.

### Make Detection More Sensitive

```bash
nosleepp watch --cpu-threshold 100ms --sample 3s
```

This catches lighter activity, at the cost of more possible false positives.

### Make Detection Stricter

```bash
nosleepp watch --cpu-threshold 1s
```

This reduces idle false positives, but may miss lightweight agent tasks.

## Limitations

- Linux support is out of scope for v2.
- Activity detection is based on CPU and descendant-process changes.
- If an agent is waiting on a network response without CPU or child-process activity, it may look idle after the quiet window.
- `--config` is reserved for future use and is not implemented yet.
- `nosleepp` prevents system idle sleep, not display sleep.

## Contributing

Contributions are welcome. Please open an issue first to discuss what you would like to change.

### Getting Started

```bash
git clone https://github.com/peterintech/nosleepp.git
cd nosleepp
go build .
```

### Running Tests

```bash
go test ./...
```

### Building

Build the local platform binary:

```bash
go build .
```

Cross-compile macOS from Windows:

```powershell
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build ./...
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build ./...
Remove-Item Env:\GOOS, Env:\GOARCH
```

### Project Structure

- `cmd/` — CLI command definitions
- `internal/` — core detection and platform logic
- `npm/` — npm package wrapper (launcher scripts and platform binaries)
