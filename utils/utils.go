package utils

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Song struct {
	Name string
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
		return "C:/Users/mathe/Music/System Of A Down/Toxicity"
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

		_, _, err = mp3.Decode(f)
		if err != nil {
			log.Fatal("cant decode file")
		}

		song := Song{
			Name: songs[songIndex],
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

func OpenSong(songs []Song, dir string, songIndex int) (streamer beep.StreamSeekCloser, format beep.Format) {
	f, err := os.Open(path.Join(dir, songs[songIndex].Name))
	if err != nil {
		log.Fatal("cant open file")
	}

	streamer, format, err = mp3.Decode(f)
	if err != nil {
		log.Fatal("cant decode file")
	}

	return streamer, format
}

func StartSong(streamer beep.StreamSeekCloser, format beep.Format) *beep.Ctrl {
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   -4.0,
		Silent:   false,
	}

	speaker.Play(beep.Seq(volume))

	return ctrl
}

func PauseSong(isSongPaused bool) bool {
	if !isSongPaused {
		speaker.Lock()
		return true
	} else {
		speaker.Unlock()
		return false
	}
}

func GoToNextSong(streamer beep.StreamSeekCloser, playlistIndex int) {
	speaker.Lock()
	streamer.Close()
	speaker.Unlock()
}
