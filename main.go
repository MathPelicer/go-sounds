package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type model struct {
	list     list.Model
	choice   string
	cursor   int
	selected map[int]struct{}
	quitting bool
}

type Song struct {
	streamer beep.StreamSeekCloser
	format   beep.Format
	name     string
}

type SongsList struct {
	songs []Song
}

type Control string

const (
	START   Control = "start"
	PLAYING Control = "playing"
	NEXT    Control = "n"
	PAUSE   Control = "p"
)

func verifyOS() string {
	if runtime.GOOS == "windows" {
		return "C:/Users/mathe/Music/System Of A Down/Hypnotize"
	} else {
		return "/home/mathe/Music"
	}
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

func initialModel() model {
	dir := verifyOS()
	songs := listSongs(dir)
	playlist := &SongsList{}
	playlist.addAllSongsToPlaylist(songs, dir)

	items := []list.Item{}
	for _, song := range playlist.songs {
		items = append(items, item(song.name))
	}

	l := list.New(items, itemDelegate{}, 20, 20)
	return model{
		list:     l,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			if m.list.Cursor() < len(m.list.Items())-1 {
				m.list.CursorDown()
			}
		case "k", "up":
			if m.list.Cursor() > 0 {
				m.list.CursorUp()
			}
		case " ", "enter":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("lets go", m.choice))
	}
	return "\n" + m.list.View()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error %v", err)
		os.Exit(1)
	}
}

func (q *SongsList) Add(songs ...Song) {
	q.songs = append(q.songs, songs...)
}

func (q *SongsList) addAllSongsToPlaylist(songs []string, dir string) {
	for songIndex := range songs {
		f, err := os.Open(path.Join(dir, songs[songIndex]))
		if err != nil {
			log.Fatal("cant open file")
		}

		streamer, format, err := mp3.Decode(f)
		if err != nil {
			log.Fatal("cant decode file")
		}

		song := Song{
			streamer: streamer,
			format:   format,
			name:     songs[songIndex],
		}

		q.Add(song)
	}
}

func listSongs(dir string) []string {
	root := os.DirFS(dir)

	songFiles, err := fs.Glob(root, "*.mp3")

	if err != nil {
		log.Fatal(err)
	}

	var songs []string
	songs = append(songs, songFiles...)

	return songs
}

func printSongs(songs []string) {
	for index, song := range songs {
		fmt.Printf("%d: %s\n", index, song)
	}
}

func selectSong(songs []string) int {
	var songIndex int
	for {
		fmt.Println("Select a song number: ")
		_, err := fmt.Scanf("%d", &songIndex)

		if err != nil {
			log.Fatal(err)
		}

		if songIndex <= len(songs) && songIndex > 0 {
			break
		}

		fmt.Println("Thats not a valid song")
	}

	return songIndex
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

func startSong(control Control, playlist *SongsList, playlistIndex int) *beep.Ctrl {
	if control == START || control == NEXT {
		sr := playlist.songs[playlistIndex].format
		speaker.Init(sr.SampleRate, sr.SampleRate.N(time.Second/10))
	}

	ctrl := &beep.Ctrl{Streamer: playlist.songs[playlistIndex].streamer, Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   -4.0,
		Silent:   false,
	}

	speaker.Play(beep.Seq(volume))

	return ctrl
}

func goToNextSong(playlist *SongsList, playlistIndex *int) {
	speaker.Lock()
	playlist.songs[*playlistIndex].streamer.Close()
	*playlistIndex += 1
	speaker.Unlock()
}

func goToSong(playlist *SongsList, playlistIndex *int, songIndex int) {
	speaker.Lock()
	playlist.songs[*playlistIndex].streamer.Close()
	*playlistIndex = songIndex
	speaker.Unlock()
}

func goToRandomSong(playlist *SongsList, playlistIndex *int) {
	speaker.Lock()
	playlist.songs[*playlistIndex].streamer.Close()
	*playlistIndex = rand.Intn(len(playlist.songs))
	speaker.Unlock()
}

func waitForUserInput(c chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		char, _ := reader.ReadString('\n')

		c <- strings.TrimSuffix(char, "\n")
	}
}

func isSongFinished(playlist *SongsList, playlistIndex int) bool {
	songLen := playlist.songs[playlistIndex].streamer.Len()
	songPos := playlist.songs[playlistIndex].streamer.Position()
	return songLen == songPos
}

func songProgress(playlist *SongsList, playlistIndex int) {
	songLen := playlist.songs[playlistIndex].streamer.Len()
	chunkSize := songLen / 30

	chunksListened := playlist.songs[playlistIndex].streamer.Position() / chunkSize

	progressString := "\r" + playlist.songs[playlistIndex].name + " ["

	for i := 0; i < 30; i++ {
		if i <= chunksListened {
			progressString += "#"
		} else {
			progressString += " "
		}
	}
	progressString += "]"

	fmt.Printf("\r%s", progressString)
}
