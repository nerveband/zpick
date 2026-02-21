package picker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/zmosh-picker/internal/zmosh"
	"golang.org/x/term"
)

// ANSI color codes matching the zsh script
const (
	reset    = "\033[0m"
	dim      = "\033[2m"
	red      = "\033[31m"
	cyan     = "\033[36m"
	green    = "\033[32m"
	yellow   = "\033[33m"
	magenta  = "\033[35m"
	boldRed  = "\033[1;31m"
	boldCyan = "\033[1;36m"
	boldGrn  = "\033[1;32m"
	boldYel  = "\033[1;33m"
	boldWht  = "\033[1;97m"
)

// ActionType represents the user's chosen action.
type ActionType int

const (
	ActionAttach ActionType = iota
	ActionNew
	ActionNewDate
	ActionCustom
	ActionZoxide
	ActionKill
	ActionEscape
)

// Action represents a user action from the picker.
type Action struct {
	Type ActionType
	Name string
}

// Run is the main interactive picker loop.
func Run() error {
	// Guard: skip if already in a zmosh session
	if os.Getenv("ZMX_SESSION") != "" && os.Getenv("ZPICK") == "" {
		return nil
	}
	// Guard: skip if not interactive
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	// Guard: if zmosh not found, show install guidance
	if _, err := exec.LookPath("zmosh"); err != nil {
		fmt.Fprintf(os.Stderr, "\n  %szmosh-picker:%s zmosh not found\n", boldCyan, reset)
		fmt.Fprintf(os.Stderr, "  %sInstall:%s brew install mmonad/tap/zmosh\n", dim, reset)
		fmt.Fprintf(os.Stderr, "  %sMore info:%s https://github.com/mmonad/zmosh\n", dim, reset)
		fmt.Fprintf(os.Stderr, "  %sRun 'zmosh-picker check' for full dependency status%s\n\n", dim, reset)
		return nil
	}

	for {
		sessions, err := zmosh.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		action, err := showPicker(sessions)
		if err != nil {
			return err
		}

		switch action.Type {
		case ActionAttach:
			return zmosh.Attach(action.Name)
		case ActionNew:
			cwd, _ := os.Getwd()
			name := CounterName(cwd, sessions)
			fmt.Printf("\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, name, reset)
			return zmosh.Attach(name)
		case ActionNewDate:
			cwd, _ := os.Getwd()
			name := DateName(cwd)
			fmt.Printf("\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, name, reset)
			return zmosh.Attach(name)
		case ActionCustom:
			return handleCustom(sessions)
		case ActionZoxide:
			dir, err := runZoxide()
			if err != nil || dir == "" {
				continue
			}
			name := CounterName(dir, sessions)
			fmt.Printf("\n  %s>%s %s%s%s %s%s%s\n\n", boldGrn, reset, boldWht, name, reset, dim, dir, reset)
			return zmosh.AttachInDir(name, dir)
		case ActionKill:
			if err := confirmAndKill(action.Name); err != nil {
				fmt.Fprintf(os.Stderr, "  %sfailed: %v%s\n", dim, err, reset)
			} else {
				fmt.Printf("  %skilled%s %s%s%s\n", boldRed, reset, boldWht, action.Name, reset)
			}
			continue // loop back to show updated list
		case ActionEscape:
			return nil // drop to plain shell
		}
	}
}

func showPicker(sessions []zmosh.Session) (Action, error) {
	fmt.Println()

	if len(sessions) > 0 {
		plural := ""
		if len(sessions) > 1 {
			plural = "s"
		}
		fmt.Printf("  %szmosh%s %s%d session%s%s\n\n", boldCyan, reset, dim, len(sessions), plural, reset)

		for i, s := range sessions {
			if i >= MaxSessions {
				break
			}
			indicator := fmt.Sprintf("%s.%s", dim, reset)
			if s.Active {
				indicator = fmt.Sprintf("%s*%s", boldGrn, reset)
			}
			dir := truncatePath(s.StartedIn, 40)
			fmt.Printf("  %s%c%s  %s%s%s %s %s%s%s\n",
				boldYel, KeyForIndex(i), reset,
				boldWht, s.Name, reset,
				indicator,
				dim, dir, reset)
		}
		fmt.Println()
	} else {
		fmt.Printf("  %szmosh%s %sno sessions%s\n\n", boldCyan, reset, dim, reset)
	}

	// Show actions
	cwd, _ := os.Getwd()
	defaultName := CounterName(cwd, sessions)
	fmt.Printf("  %senter%s %snew%s %s%s%s\n", boldGrn, reset, dim, reset, boldWht, defaultName, reset)
	fmt.Printf("  %sc%s %scustom%s  %sz%s %spick dir%s  %sd%s %s+date%s  %sk%s %skill%s  %sesc%s %sskip%s\n",
		magenta, reset, dim, reset,
		magenta, reset, dim, reset,
		cyan, reset, dim, reset,
		red, reset, dim, reset,
		yellow, reset, dim, reset)
	fmt.Println()

	// Read single keypress in raw mode
	fmt.Printf("  %s>%s ", boldCyan, reset)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return Action{}, fmt.Errorf("failed to set raw mode: %w", err)
	}

	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Println()

	if err != nil {
		return Action{}, err
	}

	key := buf[0]

	// Handle escape
	if n == 1 && key == 27 {
		return Action{Type: ActionEscape}, nil
	}
	// Handle Enter
	if n == 1 && (key == 13 || key == 10) {
		return Action{Type: ActionNew}, nil
	}

	switch key {
	case 'z':
		return Action{Type: ActionZoxide}, nil
	case 'd':
		return Action{Type: ActionNewDate}, nil
	case 'c':
		return Action{Type: ActionCustom}, nil
	case 'k':
		return enterKillMode(sessions)
	default:
		if idx, ok := IndexForKey(key); ok && idx < len(sessions) {
			fmt.Printf("\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, sessions[idx].Name, reset)
			return Action{Type: ActionAttach, Name: sessions[idx].Name}, nil
		}
	}

	return Action{Type: ActionEscape}, nil
}

func enterKillMode(sessions []zmosh.Session) (Action, error) {
	if len(sessions) == 0 {
		fmt.Printf("  %sno sessions to kill%s\n", dim, reset)
		return Action{Type: ActionEscape}, nil
	}

	fmt.Printf("\n  %skill%s %swhich session?%s ", boldRed, reset, dim, reset)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return Action{}, err
	}

	buf := make([]byte, 3)
	n, _ := os.Stdin.Read(buf)
	term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Println()

	if n == 1 && buf[0] == 27 {
		fmt.Printf("  %scancelled%s\n", dim, reset)
		return Action{Type: ActionEscape}, nil
	}

	if idx, ok := IndexForKey(buf[0]); ok && idx < len(sessions) {
		return Action{Type: ActionKill, Name: sessions[idx].Name}, nil
	}

	fmt.Printf("  %scancelled%s\n", dim, reset)
	return Action{Type: ActionEscape}, nil
}

func confirmAndKill(name string) error {
	if os.Getenv("ZMOSH_PICKER_NO_CONFIRM") == "1" {
		return zmosh.Kill(name)
	}

	fmt.Printf("  %skill %s%s%s?%s %sy/n%s ", boldRed, boldWht, name, boldRed, reset, dim, reset)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Println()

	if buf[0] == 'y' || buf[0] == 'Y' {
		return zmosh.Kill(name)
	}
	fmt.Printf("  %scancelled%s\n", dim, reset)
	return nil
}

func handleCustom(sessions []zmosh.Session) error {
	fmt.Printf("\n  %sname:%s ", magenta, reset)

	// Restore cooked mode for line input
	var customName string
	fmt.Scanln(&customName)
	customName = strings.TrimSpace(customName)
	if customName == "" {
		return nil // will re-enter picker loop from caller... but we returned
	}

	// Sub-menu: create in ~, pick dir, or cancel
	fmt.Printf("\n  %senter%s %screate in ~%s  %sz%s %spick dir%s  %sesc%s %scancel%s\n\n",
		boldGrn, reset, dim, reset,
		magenta, reset, dim, reset,
		yellow, reset, dim, reset)
	fmt.Printf("  %s>%s ", boldCyan, reset)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	buf := make([]byte, 3)
	n, _ := os.Stdin.Read(buf)
	term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Println()

	key := buf[0]

	if n == 1 && (key == 13 || key == 10) {
		// Create in current dir
		fmt.Printf("\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, customName, reset)
		return zmosh.Attach(customName)
	}

	if key == 'z' {
		dir, err := runZoxide()
		if err != nil || dir == "" {
			return nil
		}
		fmt.Printf("\n  %s>%s %s%s%s %s%s%s\n\n", boldGrn, reset, boldWht, customName, reset, dim, dir, reset)
		return zmosh.AttachInDir(customName, dir)
	}

	return nil
}

func runZoxide() (string, error) {
	if _, err := exec.LookPath("zoxide"); err != nil {
		fmt.Printf("  %szoxide not installed%s\n", yellow, reset)
		return "", nil
	}
	fmt.Println()
	cmd := exec.Command("zoxide", "query", "-i")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", nil // user cancelled fzf
	}
	return strings.TrimSpace(string(out)), nil
}

func truncatePath(path string, maxLen int) string {
	home := os.Getenv("HOME")
	if home != "" {
		path = strings.Replace(path, home, "~", 1)
	}
	if len(path) <= maxLen {
		return path
	}
	// Show last 3 path components like the zsh script
	parts := strings.Split(path, "/")
	if len(parts) > 4 {
		if strings.HasPrefix(path, "~") {
			path = "~/" + strings.Join(parts[len(parts)-3:], "/")
		} else {
			path = ".../" + strings.Join(parts[len(parts)-3:], "/")
		}
	}
	return path
}
