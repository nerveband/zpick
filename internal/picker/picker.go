package picker

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	"golang.org/x/term"
)

// ANSI color codes
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

type Action struct {
	Type ActionType
	Name string
}

// Run is the main interactive picker loop.
// Returns a shell command string to be eval'd by the caller, or empty string.
func Run(b backend.Backend) (string, error) {
	// Guard: skip if already in a session
	if b.InSession() && os.Getenv("ZPICK") == "" {
		return "", nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", nil
	}
	defer tty.Close()

	// Guard: if backend binary not found, show install guidance
	if ok, _ := b.Available(); !ok {
		fmt.Fprintf(tty, "\n  %szp:%s %s not found\n", boldCyan, reset, b.BinaryName())
		fmt.Fprintf(tty, "  %sRun 'zp check' for full dependency status%s\n\n", dim, reset)
		return "", nil
	}

	for {
		sessions, err := b.FastList()
		if err != nil {
			return "", fmt.Errorf("failed to list sessions: %w", err)
		}

		action, err := showPicker(tty, b, sessions)
		if err != nil {
			return "", err
		}

		switch action.Type {
		case ActionAttach:
			return "exec " + b.AttachCommand(action.Name, ""), nil
		case ActionNew:
			cwd, _ := os.Getwd()
			name := CounterName(cwd, sessions)
			fmt.Fprintf(tty, "\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, name, reset)
			return "exec " + b.AttachCommand(name, ""), nil
		case ActionNewDate:
			cwd, _ := os.Getwd()
			name := DateName(cwd)
			fmt.Fprintf(tty, "\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, name, reset)
			return "exec " + b.AttachCommand(name, ""), nil
		case ActionCustom:
			return handleCustom(tty, b, sessions)
		case ActionZoxide:
			dir, err := runZoxide(tty)
			if err != nil || dir == "" {
				continue
			}
			name := CounterName(dir, sessions)
			fmt.Fprintf(tty, "\n  %s>%s %s%s%s %s%s%s\n\n", boldGrn, reset, boldWht, name, reset, dim, dir, reset)
			return fmt.Sprintf("cd %q && exec %s", dir, b.AttachCommand(name, "")), nil
		case ActionKill:
			if err := confirmAndKill(tty, b, action.Name); err != nil {
				fmt.Fprintf(tty, "  %sfailed: %v%s\n", dim, err, reset)
			} else {
				fmt.Fprintf(tty, "  %skilled%s %s%s%s\n", boldRed, reset, boldWht, action.Name, reset)
			}
			continue
		case ActionEscape:
			return "", nil
		}
	}
}

func showPicker(tty *os.File, b backend.Backend, sessions []backend.Session) (Action, error) {
	fmt.Fprintln(tty)

	if len(sessions) > 0 {
		plural := ""
		if len(sessions) > 1 {
			plural = "s"
		}
		fmt.Fprintf(tty, "  %s%s%s %s%d session%s%s\n\n", boldCyan, b.Name(), reset, dim, len(sessions), plural, reset)

		for i, s := range sessions {
			if i >= MaxSessions {
				break
			}
			indicator := fmt.Sprintf("%s.%s", dim, reset)
			if s.Active {
				indicator = fmt.Sprintf("%s*%s", boldGrn, reset)
			}
			dir := truncatePath(s.StartedIn, 40)
			fmt.Fprintf(tty, "  %s%c%s  %s%s%s %s %s%s%s\n",
				boldYel, KeyForIndex(i), reset,
				boldWht, s.Name, reset,
				indicator,
				dim, dir, reset)
		}
		fmt.Fprintln(tty)
	} else {
		fmt.Fprintf(tty, "  %s%s%s %sno sessions%s\n\n", boldCyan, b.Name(), reset, dim, reset)
	}

	cwd, _ := os.Getwd()
	defaultName := CounterName(cwd, sessions)
	fmt.Fprintf(tty, "  %senter%s %snew%s %s%s%s\n", boldGrn, reset, dim, reset, boldWht, defaultName, reset)
	fmt.Fprintf(tty, "  %sc%s %scustom%s  %sz%s %spick dir%s  %sd%s %s+date%s  %sk%s %skill%s  %sesc%s %sskip%s\n",
		magenta, reset, dim, reset,
		magenta, reset, dim, reset,
		cyan, reset, dim, reset,
		red, reset, dim, reset,
		yellow, reset, dim, reset)
	fmt.Fprintln(tty)

	fmt.Fprintf(tty, "  %s>%s ", boldCyan, reset)

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return Action{}, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(tty.Fd()), oldState)

	buf := make([]byte, 3)
	n, err := tty.Read(buf)
	term.Restore(int(tty.Fd()), oldState)
	fmt.Fprintln(tty)

	if err != nil {
		return Action{}, err
	}

	key := buf[0]

	if n == 1 && key == 27 {
		return Action{Type: ActionEscape}, nil
	}
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
		return enterKillMode(tty, sessions)
	default:
		if idx, ok := IndexForKey(key); ok && idx < len(sessions) {
			fmt.Fprintf(tty, "\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, sessions[idx].Name, reset)
			return Action{Type: ActionAttach, Name: sessions[idx].Name}, nil
		}
	}

	return Action{Type: ActionEscape}, nil
}

func enterKillMode(tty *os.File, sessions []backend.Session) (Action, error) {
	if len(sessions) == 0 {
		fmt.Fprintf(tty, "  %sno sessions to kill%s\n", dim, reset)
		return Action{Type: ActionEscape}, nil
	}

	fmt.Fprintf(tty, "\n  %skill%s %swhich session?%s ", boldRed, reset, dim, reset)

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return Action{}, err
	}
	defer term.Restore(int(tty.Fd()), oldState)

	buf := make([]byte, 3)
	n, _ := tty.Read(buf)
	term.Restore(int(tty.Fd()), oldState)
	fmt.Fprintln(tty)

	if n == 1 && buf[0] == 27 {
		fmt.Fprintf(tty, "  %scancelled%s\n", dim, reset)
		return Action{Type: ActionEscape}, nil
	}

	if idx, ok := IndexForKey(buf[0]); ok && idx < len(sessions) {
		return Action{Type: ActionKill, Name: sessions[idx].Name}, nil
	}

	fmt.Fprintf(tty, "  %scancelled%s\n", dim, reset)
	return Action{Type: ActionEscape}, nil
}

func confirmAndKill(tty *os.File, b backend.Backend, name string) error {
	if os.Getenv("ZPICK_NO_CONFIRM") == "1" {
		return b.Kill(name)
	}

	fmt.Fprintf(tty, "  %skill %s%s%s?%s %sy/n%s ", boldRed, boldWht, name, boldRed, reset, dim, reset)

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(tty.Fd()), oldState)

	buf := make([]byte, 1)
	tty.Read(buf)
	term.Restore(int(tty.Fd()), oldState)
	fmt.Fprintln(tty)

	if buf[0] == 'y' || buf[0] == 'Y' {
		return b.Kill(name)
	}
	fmt.Fprintf(tty, "  %scancelled%s\n", dim, reset)
	return nil
}

func handleCustom(tty *os.File, b backend.Backend, sessions []backend.Session) (string, error) {
	fmt.Fprintf(tty, "\n  %sname:%s ", magenta, reset)

	reader := bufio.NewReader(tty)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", nil
	}
	customName := strings.TrimSpace(line)
	if customName == "" {
		return "", nil
	}

	fmt.Fprintf(tty, "\n  %senter%s %screate in ~%s  %sz%s %spick dir%s  %sesc%s %scancel%s\n\n",
		boldGrn, reset, dim, reset,
		magenta, reset, dim, reset,
		yellow, reset, dim, reset)
	fmt.Fprintf(tty, "  %s>%s ", boldCyan, reset)

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(tty.Fd()), oldState)

	buf := make([]byte, 3)
	n, _ := tty.Read(buf)
	term.Restore(int(tty.Fd()), oldState)
	fmt.Fprintln(tty)

	key := buf[0]

	if n == 1 && (key == 13 || key == 10) {
		fmt.Fprintf(tty, "\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, customName, reset)
		return "exec " + b.AttachCommand(customName, ""), nil
	}

	if key == 'z' {
		dir, err := runZoxide(tty)
		if err != nil || dir == "" {
			return "", nil
		}
		fmt.Fprintf(tty, "\n  %s>%s %s%s%s %s%s%s\n\n", boldGrn, reset, boldWht, customName, reset, dim, dir, reset)
		return fmt.Sprintf("cd %q && exec %s", dir, b.AttachCommand(customName, "")), nil
	}

	return "", nil
}

func runZoxide(tty *os.File) (string, error) {
	if _, err := exec.LookPath("zoxide"); err != nil {
		fmt.Fprintf(tty, "  %szoxide not installed%s\n", yellow, reset)
		return "", nil
	}
	fmt.Fprintln(tty)
	cmd := exec.Command("zoxide", "query", "-i")
	cmd.Stdin = tty
	cmd.Stderr = tty
	out, err := cmd.Output()
	if err != nil {
		return "", nil
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
