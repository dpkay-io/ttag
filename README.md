[![Platform](https://img.shields.io/badge/platform-windows%20%7C%20macos%20%7C%20linux-blue)](https://github.com/dpkay-io/ttag) [![License](https://img.shields.io/github/license/dpkay-io/ttag)](https://github.com/dpkay-io/ttag) [![Productivity](https://img.shields.io/badge/productivity-blue)](https://github.com/dpkay-io/ttag) [![Developer Tools](https://img.shields.io/badge/developer--tools-2ea44f)](https://github.com/dpkay-io/ttag) [![Downloads](https://img.shields.io/github/downloads/dpkay-io/ttag/total)](https://github.com/dpkay-io/ttag) [![Release](https://img.shields.io/github/v/release/dpkay-io/ttag)](https://github.com/dpkay-io/ttag)

# ttag — Terminal Tab Tagger

![Demo](cli.png)

Dynamically color your terminal tabs and set titles based on your current directory. Zero-latency, zero-dependency, cross-platform.

`ttag` injects a lightweight hook into your shell prompt. When you `cd` into a tagged directory, it instantly recolors the tab and updates the title — so you always know where you are at a glance.

## Supported Terminals

| Terminal | Tab Color | Text Contrast | Title |
|---|---|---|---|
| Windows Terminal | ✅ native | ✅ | ✅ |
| iTerm2 | ✅ native | auto | ✅ |
| tmux | ✅ native (status-bar window tab) | ✅ | ✅ |
| Apple Terminal, GNOME Terminal, VS Code, Ghostty, kitty, WezTerm, others | 🟥 emoji in title | ✅ (background mode) | ✅ |

### Color mode

Set `TTAG_COLOR_MODE` (e.g. in your shell profile) to choose where the color is shown:

| Value | Behavior |
|---|---|
| `tab` (default) | Use the terminal's native tab color where one exists (iTerm2, Windows Terminal, tmux). Other terminals have no tab color a shell can set, so the closest colored emoji (🟥🟧🟨🟩🟦🟪🟫⬛⬜) is added to the title instead. |
| `background` | Color the whole terminal background (and set a contrasting text color). |
| `both` | Tab color (or emoji) and background. |

A new mode is applied on the next directory change or `ttag set`. On iTerm2, Windows Terminal and tmux, a background left over from `background`/`both` mode stays until you open a new tab/pane after switching back to `tab`.

VS Code only shows the emoji if its tab title uses the shell's title: set `"terminal.integrated.tabs.title": "${sequence}"`.

## Install

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/dpkay-io/ttag/main/install.ps1 | iex
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/dpkay-io/ttag/main/install.sh | sh
```

## Usage

### Tag a directory

Navigate to any directory and tag it with a color:

```bash
ttag set --color "#ff6b6b" --title "API"
```

The `--title` flag is optional. If omitted, the folder name is used automatically.
If the directory is already tagged or if it inherits a color from a parent directory, the `--color` flag is also optional. You can just set a new title:

```bash
ttag set --title "Just Title"
```

### Clear a tag

Remove the color tag from the current directory:

```bash
ttag clear
```

You can also clear a specific directory without navigating to it:

```bash
ttag clear ~/projects/api
```

### List tags

View all currently tagged directories:

```bash
ttag list
```

### How it works

Once installed, `ttag` adds a tiny hook to your shell prompt. Every time you change directories, it:

1. Checks your [config file](#configuration) for a matching path
2. Walks parent directories for inherited rules (tag `~/projects` and every subdirectory inherits it)
3. Calculates optimal text contrast (white or black) using W3C luminance
4. Emits escape sequences to color the tab and set the title

### Examples

```bash
# Tag your API project red
cd ~/projects/api
ttag set --color "#ff6b6b" --title "API"

# Tag all projects with a base color (child dirs inherit)
cd ~/projects
ttag set --color "#4ecdc4" --title "Projects"

# Clear tagging
ttag clear

# Manually trigger the hook (usually automatic)
ttag hook /path/to/directory
```

## Configuration

Tags are stored in:

- macOS: `~/Library/Application Support/ttag/config.json`
- Linux: `~/.config/ttag/config.json` (or `$XDG_CONFIG_HOME/ttag/config.json`)
- Windows: `%APPDATA%\ttag\config.json`

Example:

```json
{
  "/home/user/projects": {
    "color": "#4ecdc4",
    "title": "Projects"
  },
  "/home/user/projects/api": {
    "color": "#ff6b6b",
    "title": "API"
  }
}
```

You can edit this file directly if you prefer, or use the `ttag list` command to view it.

## Shell Support

| Shell | Profile | Hook Method |
|---|---|---|
| PowerShell | `$PROFILE` | `prompt` function wrapper |
| Zsh | `~/.zshrc` | `precmd` via `add-zsh-hook` |
| Bash | `~/.bashrc` | `PROMPT_COMMAND` |

## Building from Source

```bash
go build -ldflags="-s -w" -o ttag .
```

## License

MIT
