package utils

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Song struct {
	Streamer beep.StreamSeekCloser
	format   beep.Format
	Name     string
}

type SongsList struct {
	Songs []Song
}

type Control string

const (
	START   Control = "start"
	PLAYING Control = "playing"
	NEXT    Control = "n"
	PAUSE   Control = "p"
)

func VerifyOS() string {
	if runtime.GOOS == "windows" {
		return "C:/Users/mathe/Music/System Of A Down/Hypnotize"
	} else {
		return "/home/mathe/Music"
	}
}

func (q *SongsList) Add(songs ...Song) {
	q.Songs = append(q.Songs, songs...)
}

func (q *SongsList) AddAllSongsToPlaylist(songs []string, dir string) {
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
			Streamer: streamer,
			format:   format,
			Name:     songs[songIndex],
		}

		q.Add(song)
	}
}

func ListSongs(dir string) []string {
	root := os.DirFS(dir)

	songFiles, err := fs.Glob(root, "*.mp3")

	if err != nil {
		log.Fatal(err)
	}

	var songs []string
	songs = append(songs, songFiles...)

	return songs
}

func PrintSongs(songs []string) {
	for index, song := range songs {
		fmt.Printf("%d: %s\n", index, song)
	}
}

func SelectSong(songs []string) int {
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

func StartSong(control Control, playlist *SongsList, playlistIndex int) *beep.Ctrl {
	if control == START || control == NEXT {
		sr := playlist.Songs[playlistIndex].format
		speaker.Init(sr.SampleRate, sr.SampleRate.N(time.Second/10))
	}

	ctrl := &beep.Ctrl{Streamer: playlist.Songs[playlistIndex].Streamer, Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   -4.0,
		Silent:   false,
	}

	speaker.Play(beep.Seq(volume))

	return ctrl
}

func GoToNextSong(playlist *SongsList, playlistIndex int) {
	speaker.Lock()
	playlist.Songs[playlistIndex].Streamer.Close()
	playlistIndex += 1
	speaker.Unlock()
}

func goToSong(playlist *SongsList, playlistIndex *int, songIndex int) {
	speaker.Lock()
	playlist.Songs[*playlistIndex].Streamer.Close()
	*playlistIndex = songIndex
	speaker.Unlock()
}

func goToRandomSong(playlist *SongsList, playlistIndex *int) {
	speaker.Lock()
	playlist.Songs[*playlistIndex].Streamer.Close()
	*playlistIndex = rand.Intn(len(playlist.Songs))
	speaker.Unlock()
}

func waitForUserInput(c chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		char, _ := reader.ReadString('\n')

		c <- strings.TrimSuffix(char, "\n")
	}
}

func IsSongFinished(playlist *SongsList, playlistIndex int) bool {
	songLen := playlist.Songs[playlistIndex].Streamer.Len()
	songPos := playlist.Songs[playlistIndex].Streamer.Position()
	return songLen == songPos
}

func songProgress(playlist *SongsList, playlistIndex int) {
	songLen := playlist.Songs[playlistIndex].Streamer.Len()
	chunkSize := songLen / 30

	chunksListened := playlist.Songs[playlistIndex].Streamer.Position() / chunkSize

	progressString := "\r" + playlist.Songs[playlistIndex].Name + " ["

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
