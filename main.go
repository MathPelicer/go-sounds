package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"go-sounds/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep"
)

type model struct {
	list         list.Model
	choice       string
	cursor       int
	selected     map[int]struct{}
	quitting     bool
	playlist     utils.SongsList
	songPlaying  int
	isSongPaused bool
	sub          chan struct{} // where we'll receive activity notifications
	spinner      spinner.Model
	streamer     beep.StreamSeekCloser
	format       beep.Format
}

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type item string

func (i item) FilterValue() string { return "" }

type responseMsg struct{}

func listenForActivity(m model) tea.Cmd {
	return func() tea.Msg {
		for {
			if m.isSongFinished() {
				m.sub <- struct{}{}

			}
		}
	}
}

func (m model) isSongFinished() bool {
	if m.streamer == nil {
		return false
	}
	songLen := m.streamer.Len()
	songPos := m.streamer.Position()
	return songLen == songPos
}

// A command that waits for the activity on a channel.
func waitForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

func initialModel() model {
	dir := utils.VerifyOS()
	songs := utils.ListSongs(dir)
	playlist := &utils.SongsList{}
	playlist.AddAllSongsToPlaylist(songs, dir)

	items := []list.Item{}
	for _, song := range playlist.Songs {
		items = append(items, item(song.Name))
	}

	l := list.New(items, itemDelegate{}, 20, 20)
	return model{
		list:     l,
		selected: make(map[int]struct{}),
		playlist: *playlist,
		spinner:  spinner.New(),
		sub:      make(chan struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		listenForActivity(m),   // generate activity
		waitForActivity(m.sub), // wait for activity
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c", "q":
			return &m, tea.Quit
		case "j", "down":
			if m.list.Cursor() < len(m.list.Items())-1 {
				m.list.CursorDown()
			}
		case "k", "up":
			if m.list.Cursor() > 0 {
				m.list.CursorUp()
			}
		// case "v":
		// 	ChangeViews()
		case "p":
			m.isSongPaused = utils.PauseSong(m.isSongPaused)
		case " ", "enter":
			m.songPlaying = m.list.Index()
			m.streamer, m.format = utils.OpenSong(m.playlist.Songs, utils.VerifyOS(), m.songPlaying)
			utils.StartSong(m.streamer, m.format)
			return m, waitForActivity(m.sub)
		}
	case responseMsg:
		utils.GoToNextSong(m.streamer, m.songPlaying)
		m.songPlaying += 1
		m.streamer, m.format = utils.OpenSong(m.playlist.Songs, utils.VerifyOS(), m.songPlaying)
		utils.StartSong(m.streamer, m.format)
		return m, waitForActivity(m.sub)
	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.isSongFinished() {
			m.sub <- struct{}{}
		}
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:

	}

	return &m, nil
}

// func (m model) ChangeViews() string {

// }

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("lets go", m.choice))
	}
	return "\n" + m.list.View() + "\n\n" + "playing now index: " + strconv.Itoa(m.songPlaying) + " " + m.playlist.Songs[m.songPlaying].Name
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error %v", err)
		os.Exit(1)
	}
}
