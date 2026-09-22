package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/djranoia/discforge/internal/archive"
	"github.com/djranoia/discforge/internal/jobs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type Model struct {
	root       string
	files      []string
	selected   map[int]bool
	jobs       []jobs.Job
	activeJob  int
	tab        int
	cursor     int
	width      int
	height     int
	help       help.Model
	keys       keyMap
	showHelp   bool
	search     bool
	searchTerm string
	status     string
	logText    string
	input      textarea.Model
}

var tabs = []string{"Library", "Queue", "Job", "Logs"}

var (
	accent = lipgloss.Color("205")
	muted  = lipgloss.Color("241")
	green  = lipgloss.Color("42")
	border = lipgloss.Color("238")
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)
	selectedStyle = lipgloss.NewStyle().Foreground(green)
	boxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Padding(1, 2)
)

func NewModel(root string) Model {
	h := help.New()
	t := textarea.New()
	t.Prompt = "/ "
	t.CharLimit = 128
	m := Model{root: root, selected: map[int]bool{}, activeJob: -1, help: h, keys: defaultKeyMap(), input: t}
	if files, err := archive.Scan(root); err == nil {
		m.files = files
		m.status = fmt.Sprintf("%d archive master(s) found", len(files))
	} else {
		m.status = fmt.Sprintf("archive unavailable: %v", err)
	}
	return m
}

func (m Model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd { return tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width
	case tickMsg:
		if m.activeJob >= 0 && m.activeJob < len(m.jobs) {
			m.jobs[m.activeJob] = m.jobs[m.activeJob].Advance(time.Time(msg))
			if m.jobs[m.activeJob].Status == jobs.Completed {
				m.status = "Restoration complete — ready for Sonarr"
			}
		}
		return m, tick()
	case tea.KeyMsg:
		if m.search {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Open) {
				m.search = false
				m.searchTerm = m.input.Value()
				m.input.Reset()
				return m, cmd
			}
			return m, cmd
		}
		if m.showHelp {
			if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Quit) {
				m.showHelp = false
			}
			return m, nil
		}
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.Help) {
			m.showHelp = true
			return m, nil
		}
		if key.Matches(msg, m.keys.Left) {
			m.tab = (m.tab + len(tabs) - 1) % len(tabs)
			m.cursor = 0
			return m, nil
		}
		if key.Matches(msg, m.keys.Right) {
			m.tab = (m.tab + 1) % len(tabs)
			m.cursor = 0
			return m, nil
		}
		if key.Matches(msg, m.keys.Up) { m.move(-1) }
		if key.Matches(msg, m.keys.Down) { m.move(1) }
		if key.Matches(msg, m.keys.Top) { m.cursor = 0 }
		if key.Matches(msg, m.keys.Bottom) { m.cursor = m.itemCount() - 1; if m.cursor < 0 { m.cursor = 0 } }
		if key.Matches(msg, m.keys.Open) { m.open() }
		if key.Matches(msg, m.keys.Select) && m.tab == 0 && len(m.files) > 0 { m.selected[m.cursor] = !m.selected[m.cursor] }
		if key.Matches(msg, m.keys.Restore) && m.tab == 0 { m.queueSelected() }
		if msg.String() == "/" && m.tab == 0 { m.search = true; m.input.Focus() }
	}
	return m, nil
}

func (m *Model) move(delta int) {
	count := m.itemCount()
	if count == 0 { return }
	m.cursor += delta
	if m.cursor < 0 { m.cursor = 0 }
	if m.cursor >= count { m.cursor = count - 1 }
}

func (m Model) itemCount() int {
	switch m.tab { case 0: return len(m.filteredFiles()); case 1: return len(m.jobs); case 2: return 1; case 3: return 1 }
	return 0
}

func (m Model) filteredFiles() []string {
	if m.searchTerm == "" { return m.files }
	var result []string
	for _, f := range m.files { if strings.Contains(strings.ToLower(f), strings.ToLower(m.searchTerm)) { result = append(result, f) } }
	return result
}

func (m *Model) open() {
	if m.tab == 0 && len(m.files) > 0 { m.status = "Selected archive master — press space, then r to queue" }
	if m.tab == 1 && len(m.jobs) > 0 { m.tab = 2; m.activeJob = m.cursor }
}

func (m *Model) queueSelected() {
	for i, path := range m.files {
		if !m.selected[i] { continue }
		id := fmt.Sprintf("job-%03d", len(m.jobs)+1)
		m.jobs = append(m.jobs, jobs.Job{ID: id, InputPath: path, OutputPath: filepath.Join("/mnt/work", filepath.Base(path)), Profile: "dvd-ntsc-qtgmc-hevc", Status: jobs.Queued, TotalFrames: 7192})
		m.selected[i] = false
	}
	if len(m.jobs) > 0 { m.activeJob = 0; m.tab = 1; m.status = "Queued fake restoration — no media commands are connected" }
}

func (m Model) View() string {
	if m.width == 0 { return "Starting DiscForge…" }
	header := titleStyle.Render("DiscForge") + "  " + lipgloss.NewStyle().Foreground(muted).Render("DVD restoration operator")
	nav := ""
	for i, tab := range tabs { label := fmt.Sprintf("%d %s", i+1, tab); if i == m.tab { label = lipgloss.NewStyle().Bold(true).Foreground(accent).Render("▸ " + label) }; nav += label + "   " }
	body := m.viewBody()
	footer := lipgloss.NewStyle().Foreground(muted).Render(m.status + "\n" + "h/l pane  j/k move  enter open  space select  r restore  / filter  ? help  q quit")
	view := header + "\n" + nav + "\n\n" + body + "\n\n" + footer
	if m.showHelp { view = boxStyle.Render("Keymap\n\n" + m.help.View(m.keys) + "\n\nPress ? or Esc to close") }
	if m.search { view = boxStyle.Render("Filter archive\n\n" + m.input.View() + "\n\nEnter apply · Esc cancel") }
	return lipgloss.NewStyle().Padding(1, 2).Width(m.width).Render(view)
}

func (m Model) viewBody() string {
	switch m.tab {
	case 0:
		files := m.filteredFiles()
		if len(files) == 0 { return boxStyle.Render("Library\n\nNo MKV masters found under " + m.root + ".") }
		var b strings.Builder
		b.WriteString("Library  " + lipgloss.NewStyle().Foreground(muted).Render(m.root) + "\n\n")
		for i, path := range files { marker := "[ ]"; if m.selected[i] { marker = selectedStyle.Render("[x]") }; line := fmt.Sprintf("%s %-3s %s", marker, fmt.Sprintf("%02d", i+1), path); if i == m.cursor { line = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(line) }; b.WriteString(line + "\n") }
		return boxStyle.Render(b.String())
	case 1:
		if len(m.jobs) == 0 { return boxStyle.Render("Restoration Queue\n\nQueue is empty. Select an archive master in Library and press r.") }
		var b strings.Builder; b.WriteString("Restoration Queue\n\n")
		for i, j := range m.jobs { line := fmt.Sprintf("%s %-12s %6.1f%%  %s", statusIcon(j.Status), j.ID, j.Percent(), filepath.Base(j.InputPath)); if i == m.cursor { line = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(line) }; b.WriteString(line + "\n") }
		return boxStyle.Render(b.String())
	case 2:
		if m.activeJob < 0 || m.activeJob >= len(m.jobs) { return boxStyle.Render("Current Job\n\nNo active job.") }
		j := m.jobs[m.activeJob]; return boxStyle.Render(fmt.Sprintf("Current Job\n\n%s\n\nProfile     %s\nStatus      %s\nProgress    %s %5.1f%%\nFrame       %d / %d\nFPS         %.1f\nOutput      %s", filepath.Base(j.InputPath), j.Profile, j.Status, progress(j.Percent(), 28), j.Percent(), j.Frame, j.TotalFrames, j.FPS, j.OutputPath))
	default:
		if m.logText == "" { return boxStyle.Render("Logs\n\nNo process logs yet. The fake runner keeps production side effects disabled.") }; return boxStyle.Render("Logs\n\n" + m.logText)
	}
}

func statusIcon(s jobs.Status) string { if s == jobs.Completed { return "✓" }; if s == jobs.Running { return "▶" }; return "○" }
func progress(percent float64, width int) string { filled := int(percent / 100 * float64(width)); if filled > width { filled = width }; return strings.Repeat("█", filled) + strings.Repeat("░", width-filled) }
