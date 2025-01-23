package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	DIR = "../../car-pendrive/"
)

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

func (q *SongsList) Add(songs ...Song) {
	q.songs = append(q.songs, songs...)
}

func (q *SongsList) addAllSongsToPlaylist(songs []string) {
	for songIndex := range songs {
		f, err := os.Open(path.Join(DIR, songs[songIndex]))

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

func main() {

	fmt.Println("###    Welcome to go-sounds playlist =)  ###")
	fmt.Println()
	fmt.Println("### Commands:                            ###")
	fmt.Println("| -> Enter [p] to pause/resume the song.")
	fmt.Println("| -> Enter [n] to play the next song.")
	fmt.Println("| -> Enter [l] to list all songs.")
	fmt.Println("| -> Enter [r] to try your luck ;).")
	fmt.Println("| -> Enter [song-number] to play the specific song")
	fmt.Println("############################################")

	songs := listSongs(DIR)

	playlist := &SongsList{}
	playlist.addAllSongsToPlaylist(songs)

	control := START
	c := make(chan string)
	playlistIndex := 0
	go waitForUserInput(c)

	ctrl := startSong(control, playlist, playlistIndex)

	for {
		select {
		case controlCommand := <-c:
			if controlCommand == "n" {
				goToNextSong(playlist, &playlistIndex)
				control = NEXT
				ctrl = startSong(control, playlist, playlistIndex)
			}
			if controlCommand == "p" {
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
			}
			if controlCommand == "l" {
				printSongs(songs)
			}
			if controlCommand == "r" {
				goToRandomSong(playlist, &playlistIndex)
				control = NEXT
				ctrl = startSong(control, playlist, playlistIndex)
			}

			songIndex, convErr := strconv.Atoi(controlCommand)
			if convErr == nil {
				goToSong(playlist, &playlistIndex, songIndex)
				control = NEXT
				ctrl = startSong(control, playlist, songIndex)
			}

		case <-time.After(time.Millisecond * 500):
			//fmt.Print("\033[H\033[2J")
			songProgress(playlist, playlistIndex)
		}

		if isSongFinished(playlist, playlistIndex) {
			goToNextSong(playlist, &playlistIndex)
			control = NEXT
			ctrl = startSong(control, playlist, playlistIndex)
			fmt.Println()
		}
	}
}

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
