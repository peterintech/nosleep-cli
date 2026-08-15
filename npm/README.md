# nosleepp

Keep your computer awake while AI agents are actively working.

**[View on GitHub](https://github.com/peterintech/nosleepp)**

`nosleepp` is built for agent tools such as Codex, Claude Code, OpenCode, Antigravity, and Cursor. It does not treat an app as working just because it is open. It samples process activity and only reports agents that show measurable CPU or child-process activity.

## Install Globally

Install it globally so the `nosleepp` command is available from any terminal:

```bash
npm install -g nosleepp
```

Then run:

```bash
nosleepp watch
```

## Quick Start

Check whether any agent is actively working:

```bash
nosleepp list
```

Keep your computer awake while agents are working:

```bash
nosleepp watch
```

Show matching agents even when they are open but idle:

```bash
nosleepp list --all
```

Check once and exit with a script-friendly status code:

```bash
nosleepp watch --once
```

## Supported Platforms

| Platform            | Status    |
| ------------------- | --------- |
| Windows x64         | Supported |
| macOS Apple Silicon | Supported(testing) |
| macOS Intel         | Supported(testing) |
| Linux x64           | Supported |
| Linux arm64         | Supported |

`nosleepp` prevents system idle sleep. It does not force your display to stay awake.

## How It Detects Work

`nosleepp` samples running processes for 5 seconds by default.

An agent is marked `working` when either:

- the agent process or one of its descendant processes gains enough CPU time, or
- a descendant process appears or disappears during the sample window.

If an agent app is open but idle, `nosleepp list` hides it by default. Use `nosleepp list --all` to show idle matches too.

On Linux, sleep prevention uses `systemd-inhibit --what=idle:sleep`, which is available on many systemd-based desktop distributions. Inside WSL, this does not reliably block the Windows host from sleeping; run the Windows binary when you need to keep Windows awake.

## Useful Commands

```bash
nosleepp list
nosleepp list --json
nosleepp list --all
nosleepp watch
nosleepp watch --interval 5s
nosleepp watch --quiet 3m
nosleepp power-test --duration 2m   // to test if it keeps your pc awake (without agents running)
nosleepp version
```

## Add Custom Agents

Use `--agent` to add or override process names:

```bash
nosleepp watch --agent "MyAgent=myagent,myagent-helper"
```

You can pass `--agent` more than once.

## Exit Codes

| Code | Meaning                                          |
| ---- | ------------------------------------------------ |
| `0`  | Success, or `watch --once` found a working agent |
| `1`  | `watch --once` found no working agents           |
| `2`  | Invalid CLI usage                                |
| `3`  | Platform power-state or watcher runtime error    |

---

**Its nosleepp not nosleep - the extra 'p' stands for persistence till the job is done :)**

**Like nosleepp?** [Star it on GitHub](https://github.com/peterintech/nosleepp) — it helps others find it. Contributions, issues, and ideas are welcome!
