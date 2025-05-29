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
)

type model struct {
	list        list.Model
	choice      string
	cursor      int
	selected    map[int]struct{}
	quitting    bool
	playlist    utils.SongsList
	songPlaying int
	sub         chan struct{} // where we'll receive activity notifications
	spinner     spinner.Model
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
	songLen := m.playlist.Songs[m.songPlaying].Streamer.Len()
	songPos := m.playlist.Songs[m.songPlaying].Streamer.Position()
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
		case " ", "enter":
			control := utils.START
			m.songPlaying = m.list.Index()
			utils.StartSong(control, &m.playlist, m.list.Index())
			return m, waitForActivity(m.sub)
		}
	case responseMsg:
		utils.GoToNextSong(&m.playlist, m.songPlaying)
		control := utils.NEXT
		m.songPlaying += 1
		utils.StartSong(control, &m.playlist, m.songPlaying)
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

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("lets go", m.choice))
	}
	return "\n" + m.list.View() + "\n\n" + "playing now index: " + strconv.Itoa(m.songPlaying) + " " + m.playlist.Songs[m.songPlaying].Name + strconv.Itoa(m.playlist.Songs[m.songPlaying].Streamer.Len()) + " - " + strconv.Itoa(m.playlist.Songs[m.songPlaying].Streamer.Position())
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error %v", err)
		os.Exit(1)
	}
}

//func main() {
//
//	fmt.Println("###    Welcome to go-sounds playlist =)  ###")
//	fmt.Println()
//	fmt.Println("### Commands:                            ###")
//	fmt.Println("| -> Enter [p] to pause/resume the song.")
//	fmt.Println("| -> Enter [n] to play the next song.")
//	fmt.Println("| -> Enter [l] to list all songs.")
//	fmt.Println("| -> Enter [r] to try your luck ;).")
//	fmt.Println("| -> Enter [song-number] to play the specific song")
//	fmt.Println("############################################")
//
//	songs := listSongs(DIR)
//
//	playlist := &SongsList{}
//	playlist.addAllSongsToPlaylist(songs)
//
//	control := START
//	c := make(chan string)
//	playlistIndex := 0
//	go waitForUserInput(c)
//
//	ctrl := startSong(control, playlist, playlistIndex)
//
//	for {
//		select {
//		case controlCommand := <-c:
//			if controlCommand == "n" {
//				goToNextSong(playlist, &playlistIndex)
//				control = NEXT
//				ctrl = startSong(control, playlist, playlistIndex)
//			}
//			if controlCommand == "p" {
//				speaker.Lock()
//				ctrl.Paused = !ctrl.Paused
//				speaker.Unlock()
//			}
//			if controlCommand == "l" {
//				printSongs(songs)
//			}
//			if controlCommand == "r" {
//				goToRandomSong(playlist, &playlistIndex)
//				control = NEXT
//				ctrl = startSong(control, playlist, playlistIndex)
//			}
//
//			songIndex, convErr := strconv.Atoi(controlCommand)
//			if convErr == nil {
//				goToSong(playlist, &playlistIndex, songIndex)
//				control = NEXT
//				ctrl = startSong(control, playlist, songIndex)
//			}
//
//		case <-time.After(time.Millisecond * 500):
//			//fmt.Print("\033[H\033[2J")
//			songProgress(playlist, playlistIndex)
//		}
//
//		if isSongFinished(playlist, playlistIndex) {
//			goToNextSong(playlist, &playlistIndex)
//			control = NEXT
//			ctrl = startSong(control, playlist, playlistIndex)
//			fmt.Println()
//		}
//	}
//}
