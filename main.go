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

func main() {
	fmt.Print("hello sounds...")

	songs := listSongs(DIR)
	printSongs(songs)

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

	fmt.Println(songs[songIndex])
	f, err := os.Open(path.Join(DIR, songs[songIndex]))

	if err != nil {
		log.Fatal("cant open file")
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal("cant decode file")
	}

	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	loop := beep.Loop(-1, streamer)

	ctrl := &beep.Ctrl{Streamer: loop, Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   -1.0,
		Silent:   false,
	}

	//done := make(chan bool)
	speaker.Play(volume)

	for {
		fmt.Println("Press [p] to pause/resume.")

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
