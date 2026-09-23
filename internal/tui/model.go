package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DjTuner13/DiscForge/internal/archive"
	"github.com/DjTuner13/DiscForge/internal/jobs"
	"github.com/DjTuner13/DiscForge/internal/logs"
	"github.com/DjTuner13/DiscForge/internal/pipeline"
	"github.com/DjTuner13/DiscForge/internal/probe"
	"github.com/DjTuner13/DiscForge/internal/state"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time
type probeMsg struct {
	path  string
	media probe.Media
	err   error
}
type jobDoneMsg struct {
	index int
	err   error
}
type jobProgressMsg struct {
	index    int
	progress jobs.Progress
}

type Model struct {
	root          string
	files         []string
	selected      map[string]bool
	jobs          []jobs.Job
	activeJob     int
	tab           int
	cursor        int
	width         int
	height        int
	help          help.Model
	keys          keyMap
	showHelp      bool
	search        bool
	searchTerm    string
	status        string
	logText       string
	input         textarea.Model
	store         state.Store
	media         map[string]probe.Media
	executor      jobs.Executor
	logStore      logs.Store
	running       bool
	progressCh    chan jobProgressMsg
	libraryFilter int
	confirmDelete bool
}

var tabs = []string{"Library", "Queue", "Job", "Logs / History"}
var libraryFilters = []string{"All", "Not processed", "Active", "Completed"}

var (
	accent        = lipgloss.Color("205")
	muted         = lipgloss.Color("241")
	green         = lipgloss.Color("42")
	border        = lipgloss.Color("238")
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	selectedStyle = lipgloss.NewStyle().Foreground(green)
	boxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Padding(1, 2)
)

func NewModel(root string, executor jobs.Executor) Model {
	h := help.New()
	t := textarea.New()
	t.Prompt = "/ "
	t.CharLimit = 128
	store := state.Default()
	home, _ := os.UserHomeDir()
	m := Model{root: root, selected: map[string]bool{}, activeJob: -1, help: h, keys: defaultKeyMap(), input: t, store: store, media: map[string]probe.Media{}, executor: executor, logStore: logs.Store{Root: filepath.Join(home, ".local", "state", "discforge", "logs")}, progressCh: make(chan jobProgressMsg, 32)}
	if snapshot, err := store.Load(); err == nil {
		m.jobs = snapshot.Jobs
		for _, job := range m.jobs {
			if job.Status == jobs.Interrupted {
				pipeline.CleanupArtifacts(job)
			}
		}
		if len(m.jobs) > 0 {
			m.activeJob = 0
		}
	} else {
		m.status = fmt.Sprintf("state unavailable: %v", err)
	}
	if files, err := archive.Scan(root); err == nil {
		m.files = files
		m.status = fmt.Sprintf("%d archive master(s) found", len(files))
	} else {
		m.status = fmt.Sprintf("archive unavailable: %v", err)
	}
	return m
}

func (m Model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width
	case tickMsg:
		for {
			select {
			case update := <-m.progressCh:
				if update.index >= 0 && update.index < len(m.jobs) {
					m.jobs[update.index].Frame = update.progress.Frame
					m.jobs[update.index].FPS = update.progress.FPS
				}
			default:
				goto progressDrained
			}
		}
	progressDrained:
		if m.executor == nil && m.activeJob >= 0 && m.activeJob < len(m.jobs) {
			m.jobs[m.activeJob] = m.jobs[m.activeJob].Advance(time.Time(msg))
			if m.jobs[m.activeJob].Status == jobs.Completed {
				m.status = "Restoration complete — ready for Sonarr"
			}
			if err := m.store.Save(state.Snapshot{Jobs: m.jobs}); err != nil {
				m.status = fmt.Sprintf("state save failed: %v", err)
			}
		}
		return m, tick()
	case jobDoneMsg:
		m.running = false
		if msg.index >= 0 && msg.index < len(m.jobs) {
			if msg.err != nil {
				m.jobs[msg.index].Status = jobs.Failed
				m.jobs[msg.index].Error = msg.err.Error()
				m.status = fmt.Sprintf("%s failed: %v", m.jobs[msg.index].ID, msg.err)
			} else {
				m.jobs[msg.index].Status = jobs.Completed
				m.jobs[msg.index].FinishedAt = time.Now()
				m.status = fmt.Sprintf("%s completed", m.jobs[msg.index].ID)
			}
			_ = m.store.Save(state.Snapshot{Jobs: m.jobs})
		}
		return m, m.runNext()
	case probeMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("ffprobe failed: %v", msg.err)
			return m, nil
		}
		m.media[msg.path] = msg.media
		audio, subtitles := msg.media.StreamCounts()
		m.status = fmt.Sprintf("%s · %.1fs · %d audio · %d subtitles", filepath.Base(msg.path), msg.media.DurationSeconds(), audio, subtitles)
		return m, nil
	case tea.KeyMsg:
		if m.search {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			if key.Matches(msg, m.keys.Back) {
				m.search = false
				m.input.Reset()
				return m, cmd
			}
			if key.Matches(msg, m.keys.Open) {
				m.search = false
				m.searchTerm = strings.TrimSpace(m.input.Value())
				m.input.Reset()
				m.cursor = 0
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
		if key.Matches(msg, m.keys.Back) && m.tab == 2 {
			m.tab = 1
			m.cursor = m.activeJob
			return m, nil
		}
		if key.Matches(msg, m.keys.Delete) && m.tab == 1 {
			indices := m.queueIndices()
			if m.cursor >= len(indices) {
				return m, nil
			}
			index := indices[m.cursor]
			if m.jobs[index].Status != jobs.Queued {
				m.confirmDelete = false
				m.status = "Only queued jobs can be removed"
				return m, nil
			}
			if !m.confirmDelete {
				m.confirmDelete = true
				m.status = fmt.Sprintf("Press d again to remove %s", m.jobs[index].ID)
				return m, nil
			}
			m.jobs = append(m.jobs[:index], m.jobs[index+1:]...)
			m.confirmDelete = false
			if m.cursor >= len(m.queueIndices()) && m.cursor > 0 {
				m.cursor--
			}
			m.status = "Queued job removed"
			if err := m.store.Save(state.Snapshot{Jobs: m.jobs}); err != nil {
				m.status = fmt.Sprintf("state save failed: %v", err)
			}
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
		if key.Matches(msg, m.keys.Up) {
			m.move(-1)
		}
		if key.Matches(msg, m.keys.Down) {
			m.move(1)
		}
		if key.Matches(msg, m.keys.Top) {
			m.cursor = 0
		}
		if key.Matches(msg, m.keys.Bottom) {
			m.cursor = m.itemCount() - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
		}
		if key.Matches(msg, m.keys.Open) {
			return m, m.open()
		}
		if key.Matches(msg, m.keys.Select) && m.tab == 0 && len(m.files) > 0 {
			files := m.filteredFiles()
			if m.cursor < len(files) {
				m.selected[files[m.cursor]] = !m.selected[files[m.cursor]]
			}
		}
		if key.Matches(msg, m.keys.Restore) && m.tab == 0 {
			m.queueSelected()
			return m, m.runNext()
		}
		if msg.String() == "/" && m.tab == 0 {
			m.search = true
			m.input.Focus()
		}
		if key.Matches(msg, m.keys.Filter) && m.tab == 0 {
			m.libraryFilter = (m.libraryFilter + 1) % len(libraryFilters)
			m.cursor = 0
			m.status = "Library filter: " + libraryFilters[m.libraryFilter]
		}
	}
	return m, nil
}

func (m *Model) move(delta int) {
	count := m.itemCount()
	if count == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= count {
		m.cursor = count - 1
	}
}

func (m Model) itemCount() int {
	switch m.tab {
	case 0:
		return len(m.filteredFiles())
	case 1:
		return len(m.queueIndices())
	case 2:
		return 1
	case 3:
		return len(m.historyIndices())
	}
	return 0
}

func (m Model) queueIndices() []int {
	var indices []int
	for i, job := range m.jobs {
		switch job.Status {
		case jobs.Queued, jobs.Preparing, jobs.Running, jobs.Muxing, jobs.Validating:
			indices = append(indices, i)
		}
	}
	return indices
}

func (m Model) historyIndices() []int {
	var indices []int
	for i, job := range m.jobs {
		switch job.Status {
		case jobs.Completed, jobs.Failed, jobs.Cancelled, jobs.Interrupted:
			indices = append(indices, i)
		}
	}
	return indices
}

func (m Model) filteredFiles() []string {
	var result []string
	term := strings.ToLower(strings.TrimSpace(m.searchTerm))
	for _, f := range m.files {
		if !m.matchesLibraryFilter(f) {
			continue
		}
		if term == "" || strings.Contains(strings.ToLower(f), term) {
			result = append(result, f)
		}
	}
	return result
}

func (m Model) matchesLibraryFilter(path string) bool {
	if m.libraryFilter == 0 {
		return true
	}
	active, completed := false, false
	for _, job := range m.jobs {
		if job.InputPath != path {
			continue
		}
		switch job.Status {
		case jobs.Queued, jobs.Preparing, jobs.Running, jobs.Muxing, jobs.Validating:
			active = true
		case jobs.Completed:
			completed = true
		}
	}
	switch m.libraryFilter {
	case 1:
		return !active && !completed
	case 2:
		return active
	case 3:
		return completed
	default:
		return true
	}
}

func (m *Model) open() tea.Cmd {
	if m.tab == 0 {
		files := m.filteredFiles()
		if len(files) == 0 {
			m.status = "No archive titles match the current filter"
			return nil
		}
		m.status = "Selected archive master — press space, then r to queue"
		if m.cursor >= len(files) {
			m.cursor = len(files) - 1
		}
		path := files[m.cursor]
		if _, ok := m.media[path]; !ok {
			return probeFile(path)
		}
	}
	if m.tab == 1 && len(m.jobs) > 0 {
		m.tab = 2
		indices := m.queueIndices()
		if m.cursor < len(indices) {
			m.activeJob = indices[m.cursor]
		}
	}
	if m.tab == 3 {
		indices := m.historyIndices()
		if m.cursor < len(indices) {
			m.activeJob = indices[m.cursor]
			if data, err := m.logStore.Read(m.jobs[m.activeJob].ID); err == nil {
				m.logText = string(data)
			} else {
				m.logText = "No process log was recorded for this job."
			}
		}
	}
	return nil
}

func probeFile(path string) tea.Cmd {
	return func() tea.Msg {
		media, err := (probe.Runner{}).Inspect(context.Background(), path)
		return probeMsg{path: path, media: media, err: err}
	}
}

func (m *Model) queueSelected() {
	for _, path := range m.files {
		if !m.selected[path] {
			continue
		}
		id := fmt.Sprintf("job-%03d", len(m.jobs)+1)
		totalFrames := int64(0)
		if media, ok := m.media[path]; ok {
			totalFrames = estimateFrames(media)
		}
		m.jobs = append(m.jobs, jobs.Job{ID: id, InputPath: path, OutputPath: filepath.Join("/mnt/work", filepath.Base(path)), Profile: "dvd-ntsc-qtgmc-hevc", Status: jobs.Queued, TotalFrames: totalFrames})
		m.selected[path] = false
	}
	if len(m.jobs) > 0 {
		m.activeJob = 0
		m.tab = 1
		if m.executor == nil {
			m.status = "Queued simulated restoration — use --live to run media commands"
		} else {
			m.status = "Queued live restoration"
		}
		if err := m.store.Save(state.Snapshot{Jobs: m.jobs}); err != nil {
			m.status = fmt.Sprintf("state save failed: %v", err)
		}
	}
}

func (m *Model) runNext() tea.Cmd {
	if m.executor == nil || m.running {
		return nil
	}
	for i := range m.jobs {
		if m.jobs[i].Status != jobs.Queued {
			continue
		}
		if m.jobs[i].TotalFrames == 0 {
			if media, err := (probe.Runner{}).Inspect(context.Background(), m.jobs[i].InputPath); err == nil {
				m.media[m.jobs[i].InputPath] = media
				m.jobs[i].TotalFrames = estimateFrames(media)
			}
		}
		m.activeJob = i
		m.running = true
		m.jobs[i].Status = jobs.Running
		m.jobs[i].StartedAt = time.Now()
		_ = m.store.Save(state.Snapshot{Jobs: m.jobs})
		index := i
		job := m.jobs[i]
		return func() tea.Msg {
			logFile, err := m.logStore.Open(job.ID)
			if err != nil {
				return jobDoneMsg{index: index, err: err}
			}
			defer logFile.Close()
			if writer, ok := m.executor.(interface {
				ExecuteWithProgress(context.Context, jobs.Job, io.Writer, func(jobs.Progress)) error
			}); ok {
				callback := func(progress jobs.Progress) {
					select {
					case m.progressCh <- jobProgressMsg{index: index, progress: progress}:
					default:
					}
				}
				return jobDoneMsg{index: index, err: writer.ExecuteWithProgress(context.Background(), job, logFile, callback)}
			}
			if writer, ok := m.executor.(interface {
				ExecuteWithLog(context.Context, jobs.Job, io.Writer) error
			}); ok {
				return jobDoneMsg{index: index, err: writer.ExecuteWithLog(context.Background(), job, logFile)}
			}
			return jobDoneMsg{index: index, err: m.executor.Execute(context.Background(), job)}
		}
	}
	return nil
}

func estimateFrames(media probe.Media) int64 {
	for _, stream := range media.Streams {
		if stream.CodecType != "video" {
			continue
		}
		parts := strings.SplitN(stream.RFrameRate, "/", 2)
		if len(parts) != 2 {
			return 0
		}
		numerator, errN := strconv.ParseFloat(parts[0], 64)
		denominator, errD := strconv.ParseFloat(parts[1], 64)
		if errN != nil || errD != nil || denominator == 0 {
			return 0
		}
		return int64(media.DurationSeconds() * numerator / denominator * 2)
	}
	return 0
}

func (m Model) View() string {
	if m.width == 0 {
		return "Starting DiscForge…"
	}
	header := titleStyle.Render("DiscForge") + "  " + lipgloss.NewStyle().Foreground(muted).Render("DVD restoration operator")
	nav := ""
	for i, tab := range tabs {
		label := fmt.Sprintf("%d %s", i+1, tab)
		if i == m.tab {
			label = lipgloss.NewStyle().Bold(true).Foreground(accent).Render("▸ " + label)
		}
		nav += label + "   "
	}
	body := m.viewBody()
	footer := lipgloss.NewStyle().Foreground(muted).Render(m.status + "\n" + "h/l pane  j/k move  enter open  space select  r restore  d remove queued  f library filter  / search  ? help  q quit")
	view := header + "\n" + nav + "\n\n" + body + "\n\n" + footer
	if m.showHelp {
		view = boxStyle.Render("Keymap\n\n" + m.help.View(m.keys) + "\n\nPress ? or Esc to close")
	}
	if m.search {
		view = boxStyle.Render("Filter archive\n\n" + m.input.View() + "\n\nEnter apply · Esc cancel")
	}
	return lipgloss.NewStyle().Padding(1, 2).Width(m.width).Render(view)
}

func (m Model) viewBody() string {
	switch m.tab {
	case 0:
		files := m.filteredFiles()
		if len(files) == 0 {
			return boxStyle.Render("Library\n\nNo MKV masters found under " + m.root + ".")
		}
		var b strings.Builder
		b.WriteString("Library  " + lipgloss.NewStyle().Foreground(muted).Render(m.root) + "  [" + libraryFilters[m.libraryFilter] + "]\n\n")
		for i, path := range files {
			marker := "[ ]"
			if m.selected[path] {
				marker = selectedStyle.Render("[x]")
			}
			line := fmt.Sprintf("%s %-3s %s", marker, fmt.Sprintf("%02d", i+1), path)
			if i == m.cursor {
				line = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(line)
			}
			b.WriteString(line + "\n")
		}
		if m.cursor < len(files) {
			if media, ok := m.media[files[m.cursor]]; ok {
				for _, stream := range media.Streams {
					if stream.CodecType == "video" {
						b.WriteString(fmt.Sprintf("\nVideo  %s  %dx%d  %s  %s", stream.CodecName, stream.Width, stream.Height, stream.RFrameRate, stream.FieldOrder))
						break
					}
				}
			}
		}
		return boxStyle.Render(b.String())
	case 1:
		indices := m.queueIndices()
		if len(indices) == 0 {
			return boxStyle.Render("Restoration Queue\n\nQueue is empty. Select an archive master in Library and press r.")
		}
		var b strings.Builder
		b.WriteString("Restoration Queue\n\n")
		for i, index := range indices {
			j := m.jobs[index]
			line := fmt.Sprintf("%s %-12s %6.1f%%  %s", statusIcon(j.Status), j.ID, j.Percent(), filepath.Base(j.InputPath))
			if i == m.cursor {
				line = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(line)
			}
			b.WriteString(line + "\n")
		}
		return boxStyle.Render(b.String())
	case 2:
		if m.activeJob < 0 || m.activeJob >= len(m.jobs) {
			return boxStyle.Render("Current Job\n\nNo active job.")
		}
		j := m.jobs[m.activeJob]
		return boxStyle.Render(fmt.Sprintf("Current Job\n\n%s\n\nProfile     %s\nStatus      %s\nProgress    %s %5.1f%%\nFrame       %d / %d\nFPS         %.1f\nOutput      %s", filepath.Base(j.InputPath), j.Profile, j.Status, progress(j.Percent(), 28), j.Percent(), j.Frame, j.TotalFrames, j.FPS, j.OutputPath))
	default:
		indices := m.historyIndices()
		if len(indices) == 0 {
			return boxStyle.Render("Logs / History\n\nNo completed or interrupted jobs yet.")
		}
		var b strings.Builder
		b.WriteString("Logs / History\n\n")
		for i, index := range indices {
			j := m.jobs[index]
			line := fmt.Sprintf("%s %-12s %-11s %s", statusIcon(j.Status), j.ID, j.Status, filepath.Base(j.InputPath))
			if i == m.cursor {
				line = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(line)
			}
			b.WriteString(line + "\n")
		}
		if m.logText != "" {
			b.WriteString("\nSelected log\n\n" + m.logText)
		}
		return boxStyle.Render(b.String())
	}
}

func statusIcon(s jobs.Status) string {
	if s == jobs.Completed {
		return "✓"
	}
	if s == jobs.Running {
		return "▶"
	}
	return "○"
}
func progress(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
