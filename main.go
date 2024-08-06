package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	DIR = "../../car-pendrive/"
)

type Song struct {
	streamer beep.Streamer
	format   beep.Format
}

type Queue struct {
	songs []Song
}

type Control string

const (
	START   Control = "start"
	PLAYING Control = "playing"
	NEXT    Control = "next"
	PAUSE   Control = "p"
)

func (q *Queue) Add(songs ...Song) {
	q.songs = append(q.songs, songs...)
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

func (q *Queue) addAllSongsToPlaylist(songs []string) {
	for songIndex, _ := range songs {
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
		}

		q.Add(song)
	}
}

func main() {

	fmt.Println("go-sounds playlist =)")

	songs := listSongs(DIR)
	printSongs(songs)

	playlist := &Queue{}
	playlist.addAllSongsToPlaylist(songs)
	//songIndex := selectSong(songs)

	//fmt.Println(songs[songIndex])
	//f, err := os.Open(path.Join(DIR, songs[songIndex]))

	//if err != nil {
	//	log.Fatal("cant open file")
	//}

	//streamer, format, err := mp3.Decode(f)
	//if err != nil {
	//	log.Fatal("cant decode file")
	//}

	//defer playlist.streamers[0].Close()
	control := START
	playlistIndex := 0
	for len(playlist.songs) > 0 {
		if control == START || control == NEXT {
			sr := playlist.songs[playlistIndex].format
			speaker.Init(sr.SampleRate, sr.SampleRate.N(time.Second/10))
		}

		ctrl := &beep.Ctrl{Streamer: playlist.songs[playlistIndex].streamer, Paused: false}
		volume := &effects.Volume{
			Streamer: ctrl,
			Base:     2,
			Volume:   -1.0,
			Silent:   false,
		}

		fmt.Printf("Now playing: %s\n", songs[playlistIndex])
		speaker.Play(volume)

		control = PLAYING

		fmt.Println("Enter [p] to pause/resume.")
		fmt.Println("Enter [next] to play the next song.")

		for {
			fmt.Scan(&control)

			if control == NEXT {
				speaker.Lock()
				playlistIndex += 1
				speaker.Unlock()
				break
			}
			if control == PAUSE {
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
			}
		}

		//useKeyboardLib(*ctrl)
		//select {
		//case <-done:
		//	return
		//case <-time.After(time.Second):
		//	speaker.Lock()
		//	fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
		//	speaker.Unlock()
		//}
	}
}

func useKeyboardLib(ctrl beep.Ctrl) {
	control, key, err := keyboard.GetSingleKey()
	if err != nil {
		panic(err)
	}

	//fmt.Scanln()
	if control == rune(112) {
		speaker.Lock()
		ctrl.Paused = !ctrl.Paused
		speaker.Unlock()
	}

	if key == keyboard.KeyEsc {
		return
	}
}
