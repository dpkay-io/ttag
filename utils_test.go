package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseHex(t *testing.T) {
	r, g, b := parseHex("#ff6b6b")
	if r != 255 || g != 107 || b != 107 {
		t.Errorf("Expected 255, 107, 107, got %d, %d, %d", r, g, b)
	}

	r, g, b = parseHex("000000")
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("Expected 0, 0, 0, got %d, %d, %d", r, g, b)
	}
}

func TestContrastColor(t *testing.T) {
	// Dark background -> white text
	r, g, b := contrastColor(0, 0, 0)
	if r != 255 || g != 255 || b != 255 {
		t.Errorf("Expected white text for dark bg, got %d, %d, %d", r, g, b)
	}

	// Light background -> black text
	r, g, b = contrastColor(255, 255, 255)
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("Expected black text for light bg, got %d, %d, %d", r, g, b)
	}
}

func TestDetectTerminal(t *testing.T) {
	t.Setenv("TMUX", "")
	os.Setenv("WT_SESSION", "1")
	if detectTerminal() != termWindowsTerminal {
		t.Errorf("Expected termWindowsTerminal")
	}
	os.Unsetenv("WT_SESSION")

	os.Setenv("TERM_PROGRAM", "iTerm.app")
	if detectTerminal() != termITerm2 {
		t.Errorf("Expected termITerm2")
	}
	os.Unsetenv("TERM_PROGRAM")

	if detectTerminal() != termStandard {
		t.Errorf("Expected termStandard")
	}
}

func TestColorSequences(t *testing.T) {
	t.Setenv("TTAG_COLOR_MODE", "")
	t.Setenv("TMUX", "")
	os.Setenv("TERM_PROGRAM", "iTerm.app")
	seq := colorSequences(255, 0, 0, 255, 255, 255, "Test")
	if !strings.Contains(seq, "bg;red;brightness;255") {
		t.Errorf("Expected iTerm2 sequence, got %q", seq)
	}
	os.Unsetenv("TERM_PROGRAM")

	os.Setenv("WT_SESSION", "1")
	seq = colorSequences(255, 0, 0, 255, 255, 255, "Test")
	if !strings.Contains(seq, "4;264;rgb:ff/00/00") {
		t.Errorf("Expected WT palette 264 for tab bg, got %q", seq)
	}
	if !strings.Contains(seq, "4;263;rgb:ff/ff/ff") {
		t.Errorf("Expected WT palette 263 for tab fg, got %q", seq)
	}
	os.Unsetenv("WT_SESSION")

	// Terminals without a native tab color leave the background alone by default
	seq = colorSequences(255, 0, 0, 255, 255, 255, "Test")
	if strings.Contains(seq, "11;rgb:") {
		t.Errorf("Expected no background color in tab mode, got %q", seq)
	}

	t.Setenv("TTAG_COLOR_MODE", "background")
	seq = colorSequences(255, 0, 0, 255, 255, 255, "Test")
	if !strings.Contains(seq, "11;rgb:ff/00/00") {
		t.Errorf("Expected OSC 11 background sequence, got %q", seq)
	}
	if !strings.Contains(seq, "10;rgb:ff/ff/ff") {
		t.Errorf("Expected OSC 10 foreground sequence, got %q", seq)
	}
}

func TestResetSequences(t *testing.T) {
	t.Setenv("TTAG_COLOR_MODE", "")
	t.Setenv("TMUX", "")
	os.Setenv("TERM_PROGRAM", "iTerm.app")
	seq := resetSequences()
	if !strings.Contains(seq, "bg;*;default") {
		t.Errorf("Expected iTerm2 reset, got %q", seq)
	}
	os.Unsetenv("TERM_PROGRAM")

	os.Setenv("WT_SESSION", "1")
	seq = resetSequences()
	if !strings.Contains(seq, "104;263;264") {
		t.Errorf("Expected WT palette reset, got %q", seq)
	}
	os.Unsetenv("WT_SESSION")

	seq = resetSequences()
	if !strings.Contains(seq, "110") {
		t.Errorf("Expected OSC 110 foreground reset, got %q", seq)
	}
	if !strings.Contains(seq, "111") {
		t.Errorf("Expected OSC 111 background reset, got %q", seq)
	}
}

func TestDetectShell(t *testing.T) {
	// We can't strictly test detectShell on Windows vs Linux completely because of runtime.GOOS,
	// but we can test environment variable parsing if not windows
	if runtimeOS := "linux"; runtimeOS == "linux" { // simulating non-windows behavior testing logic
		os.Setenv("SHELL", "/bin/zsh")
		if detectShell() != "zsh" && detectShell() != "powershell" { // powershell is forced on windows
			t.Errorf("Expected zsh")
		}
	}
}

func TestDefaultProfilePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	
	ps := defaultProfilePath("powershell")
	if ps != filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1") {
		t.Errorf("Unexpected powershell profile path: %s", ps)
	}

	zsh := defaultProfilePath("zsh")
	if zsh != filepath.Join(home, ".zshrc") {
		t.Errorf("Unexpected zsh profile path: %s", zsh)
	}

	bash := defaultProfilePath("bash")
	if bash != filepath.Join(home, ".bashrc") {
		t.Errorf("Unexpected bash profile path: %s", bash)
	}
}

func TestHookSnippetVersion(t *testing.T) {
	for _, shell := range []string{"powershell", "zsh", "bash"} {
		snippet := hookSnippet(shell)
		expected := "# ttag hook v" + version
		if !strings.Contains(snippet, expected) {
			t.Errorf("%s hook missing version line %q", shell, expected)
		}
	}
}

func TestInstalledHookVersion(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "profile")

	// No file → empty
	if v := installedHookVersion("bash", profile); v != "" {
		t.Errorf("Expected empty version for missing profile, got %q", v)
	}

	// Write hook with version
	snippet := hookSnippet("bash")
	os.WriteFile(profile, []byte("# stuff\n"+snippet+"\n"), 0644)
	if v := installedHookVersion("bash", profile); v != version {
		t.Errorf("Expected %q, got %q", version, v)
	}

	// Write hook without version line (simulating old install)
	old := hookMarkerStart + "\nsome old content\n" + hookMarkerEnd
	os.WriteFile(profile, []byte(old), 0644)
	if v := installedHookVersion("bash", profile); v != "" {
		t.Errorf("Expected empty version for old hook, got %q", v)
	}
}

func TestHookSnippet(t *testing.T) {
	ps := hookSnippet("powershell")
	if !strings.Contains(ps, "global:prompt") {
		t.Errorf("Expected PowerShell hook")
	}
	if strings.Contains(ps, "__ttag_focused") {
		t.Errorf("PowerShell hook should not contain focus tracking (PSReadLine cannot bind escape sequences)")
	}

	zsh := hookSnippet("zsh")
	if !strings.Contains(zsh, "add-zsh-hook") {
		t.Errorf("Expected Zsh hook")
	}
	if !strings.Contains(zsh, "?1004h") {
		t.Errorf("Expected focus reporting enable in Zsh hook")
	}
	if !strings.Contains(zsh, "__ttag_focus_in") {
		t.Errorf("Expected focus handlers in Zsh hook")
	}
	if !strings.Contains(zsh, "zshexit") {
		t.Errorf("Expected exit cleanup in Zsh hook")
	}

	bash := hookSnippet("bash")
	if !strings.Contains(bash, "PROMPT_COMMAND") {
		t.Errorf("Expected Bash hook")
	}
	if !strings.Contains(bash, "?1004h") {
		t.Errorf("Expected focus reporting enable in Bash hook")
	}
	if !strings.Contains(bash, "__ttag_focus_in") {
		t.Errorf("Expected focus handlers in Bash hook")
	}
	if !strings.Contains(bash, "__ttag_cleanup") {
		t.Errorf("Expected exit cleanup in Bash hook")
	}
	// bash < 4.3 cannot bind -x escape sequences; focus tracking must be gated
	if !strings.Contains(bash, `"${BASH_VERSINFO[1]}" -ge 3`) {
		t.Errorf("Expected bash 4.3+ gate for focus tracking")
	}
	gate := strings.Index(bash, `if [ "$__ttag_focus_ok" = "1" ]; then`)
	bindPos := strings.Index(bash, "bind -x")
	if gate < 0 || bindPos < gate {
		t.Errorf("bind -x must only run inside the bash version gate")
	}
}

func setTerm(t *testing.T, termProgram, wtSession, mode string) {
	t.Helper()
	t.Setenv("TERM_PROGRAM", termProgram)
	t.Setenv("WT_SESSION", wtSession)
	t.Setenv("TTAG_COLOR_MODE", mode)
	t.Setenv("TMUX", "")
	t.Setenv("TMUX_PANE", "")
}

func TestDetectAppleTerminal(t *testing.T) {
	setTerm(t, "Apple_Terminal", "", "")
	if detectTerminal() != termAppleTerminal {
		t.Errorf("Expected termAppleTerminal")
	}
}

func TestColorModeSequences(t *testing.T) {
	const bg = "11;rgb:ff/00/00"
	cases := []struct {
		name, term, wt, mode string
		wantTab, wantBg      bool
	}{
		{"apple default", "Apple_Terminal", "", "", false, false},
		{"apple background", "Apple_Terminal", "", "background", false, true},
		{"apple both", "Apple_Terminal", "", "both", false, true},
		{"iterm default", "iTerm.app", "", "", true, false},
		{"iterm background", "iTerm.app", "", "background", false, true},
		{"iterm both", "iTerm.app", "", "both", true, true},
		{"wt default", "", "1", "", true, false},
		{"wt background", "", "1", "bg", false, true},
		{"standard default", "", "", "", false, false},
		{"standard tab", "", "", "tab", false, false},
		{"standard background", "", "", "background", false, true},
		{"vscode default", "vscode", "", "", false, false},
		{"ghostty both", "ghostty", "", "both", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setTerm(t, c.term, c.wt, c.mode)
			seq := colorSequences(255, 0, 0, 255, 255, 255, "")
			gotTab := strings.Contains(seq, "bg;red;brightness;") || strings.Contains(seq, "]4;264;")
			if gotTab != c.wantTab {
				t.Errorf("tab color = %v, want %v (%q)", gotTab, c.wantTab, seq)
			}
			if gotBg := strings.Contains(seq, bg); gotBg != c.wantBg {
				t.Errorf("background = %v, want %v (%q)", gotBg, c.wantBg, seq)
			}
		})
	}
}

func TestResetSequencesModes(t *testing.T) {
	setTerm(t, "iTerm.app", "", "")
	if seq := resetSequences(); strings.Contains(seq, "111") || !strings.Contains(seq, "bg;*;default") {
		t.Errorf("iTerm2 tab mode should reset only the tab, got %q", seq)
	}
	setTerm(t, "iTerm.app", "", "both")
	if seq := resetSequences(); !strings.Contains(seq, "111") || !strings.Contains(seq, "bg;*;default") {
		t.Errorf("iTerm2 both mode should reset tab and background, got %q", seq)
	}
	setTerm(t, "Apple_Terminal", "", "")
	if seq := resetSequences(); !strings.Contains(seq, "111") {
		t.Errorf("Apple Terminal should always reset background, got %q", seq)
	}
	setTerm(t, "iTerm.app", "", "background")
	if seq := resetSequences(); !strings.Contains(seq, "111") || !strings.Contains(seq, "bg;*;default") {
		t.Errorf("iTerm2 background mode should reset tab and background, got %q", seq)
	}
	setTerm(t, "", "1", "background")
	if seq := resetSequences(); !strings.Contains(seq, "111") || !strings.Contains(seq, "104;263;264") {
		t.Errorf("WT background mode should reset tab and background, got %q", seq)
	}
	setTerm(t, "", "", "")
	if seq := resetSequences(); seq != "\033]110\a\033]111\a" {
		t.Errorf("standard reset should be unchanged, got %q", seq)
	}
	setTerm(t, "", "", "")
	t.Setenv("TMUX", "/tmp/tmux-1/default,1,0")
	if seq := resetSequences(); seq != "" {
		t.Errorf("tmux tab mode reset is a tmux command, not a sequence, got %q", seq)
	}
}

func TestStandardTabModeClearsBackground(t *testing.T) {
	setTerm(t, "", "", "")
	if seq := colorSequences(255, 0, 0, 255, 255, 255, ""); !strings.Contains(seq, "\033]111\a") {
		t.Errorf("tab mode without native tab color should clear a stale background, got %q", seq)
	}
}

func TestDetectTmux(t *testing.T) {
	// Panes inherit the outer terminal's variables; tmux must still win
	setTerm(t, "iTerm.app", "1", "")
	t.Setenv("TMUX", "/tmp/tmux-1/default,1,0")
	if detectTerminal() != termTmux {
		t.Errorf("Expected termTmux when TMUX is set")
	}
}

func TestTmuxTabArgs(t *testing.T) {
	t.Setenv("TMUX_PANE", "%3")
	got := strings.Join(tmuxTabArgs("#ff6b6b", "#000000"), " ")
	want := "set-option -w -t %3 window-status-style bg=#ff6b6b,fg=#000000 ; " +
		"set-option -w -t %3 window-status-current-style bg=#ff6b6b,fg=#000000"
	if got != want {
		t.Errorf("apply args:\n got %q\nwant %q", got, want)
	}
	got = strings.Join(tmuxTabArgs("", ""), " ")
	want = "set-option -w -u -t %3 window-status-style ; set-option -w -u -t %3 window-status-current-style"
	if got != want {
		t.Errorf("reset args:\n got %q\nwant %q", got, want)
	}
	t.Setenv("TMUX_PANE", "")
	if got := strings.Join(tmuxTabArgs("", ""), " "); strings.Contains(got, "-t") {
		t.Errorf("without TMUX_PANE no target should be passed, got %q", got)
	}
}

func TestBackgroundModeClearsTabColor(t *testing.T) {
	setTerm(t, "iTerm.app", "", "background")
	if seq := colorSequences(255, 0, 0, 255, 255, 255, ""); !strings.Contains(seq, "bg;*;default") {
		t.Errorf("iTerm2 background mode should clear a stale tab color, got %q", seq)
	}
	setTerm(t, "", "1", "background")
	if seq := colorSequences(255, 0, 0, 255, 255, 255, ""); !strings.Contains(seq, "104;263;264") {
		t.Errorf("WT background mode should clear a stale tab color, got %q", seq)
	}
}

func TestColorModeParsing(t *testing.T) {
	cases := map[string]colorMode{
		"": modeTab, "tab": modeTab, "garbage": modeTab,
		"background": modeBackground, "bg": modeBackground, " Background ": modeBackground,
		"both": modeBoth, "BOTH": modeBoth,
	}
	for v, want := range cases {
		t.Setenv("TTAG_COLOR_MODE", v)
		if got := currentColorMode(); got != want {
			t.Errorf("TTAG_COLOR_MODE=%q: got %v, want %v", v, got, want)
		}
	}
}

func TestUsesSwatch(t *testing.T) {
	for _, term := range []termType{termAppleTerminal, termStandard} {
		if !usesSwatch(term, modeTab) || !usesSwatch(term, modeBoth) {
			t.Errorf("terminal %d without native tab color should use swatch in tab and both modes", term)
		}
		if usesSwatch(term, modeBackground) {
			t.Errorf("terminal %d should not use swatch in background mode", term)
		}
	}
	for _, term := range []termType{termITerm2, termWindowsTerminal, termTmux} {
		if usesSwatch(term, modeTab) {
			t.Errorf("terminal %d has a native tab color and should not use swatch", term)
		}
	}
}

func TestSwatchFor(t *testing.T) {
	cases := map[string][3]uint8{
		"🟥": {0xff, 0x6b, 0x6b},
		"🟦": {0x00, 0x66, 0xff},
		"⬛": {0x11, 0x22, 0x33},
		"🟩": {0x2e, 0xcc, 0x40},
		"⬜": {0xff, 0xff, 0xff},
	}
	for want, c := range cases {
		if got := swatchFor(c[0], c[1], c[2]); got != want {
			t.Errorf("swatchFor(%v) = %s, want %s", c, got, want)
		}
	}
}

func TestSwatchForHues(t *testing.T) {
	cases := map[string]string{
		"#8b4513": "🟫", // saddle brown
		"#ffa500": "🟧",
		"#ffd700": "🟨",
		"#4ecdc4": "🟦", // teal falls into blue
		"#9b59b6": "🟪",
		"#ff6b6b": "🟥", // README example
		"#666666": "⬛",
		"#cccccc": "⬜",
	}
	for hex, want := range cases {
		r, g, b := parseHex(hex)
		if got := swatchFor(r, g, b); got != want {
			t.Errorf("swatchFor(%s) = %s, want %s", hex, got, want)
		}
	}
}
