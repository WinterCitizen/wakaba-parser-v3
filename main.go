package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	url := flag.String("u", "", "thread url")
	images := flag.Bool("i", false, "save images")
	video := flag.Bool("v", false, "save video")
	directory := flag.String("d", time.Now().Format("2006-01-02 15-04-05"), "download directory")
	flag.Parse()
	resp, err := http.Get(strings.Replace(*url, ".html", ".json", 1))
	if err != nil {
		log.Fatal("Can't get required url")
	}
	defer resp.Body.Close()

	var page Page
	json.NewDecoder(resp.Body).Decode(&page)
	downloadMedia(&page, images, video, directory)
}

func downloadMedia(page *Page, images, video *bool, directory *string) {
	_ = os.Mkdir(*directory, 655)
	wg := sync.WaitGroup{}
	amount, count := 1, 0
	for _, post := range page.Threads[0].Posts {
		for _, file := range post.Files {
			if *images {
				if !strings.HasSuffix(file.Path, ".webm") {
					wg.Add(1)
					go storeImage(&file, directory, &wg, &amount, &count)
				}
			}
			if *video {
				if strings.HasSuffix(file.Path, ".webm") {
					wg.Add(1)
					go storeImage(&file, directory, &wg, &amount, &count)
				}
			}
			count++
		}
	}
	wg.Wait()
}

func storeImage(file *File, directory *string, wg *sync.WaitGroup, amount, count *int) {
	image := make(chan []byte)
	go getMediaData(file.Path, image)
	f, err := os.Create(*directory + `/` + file.Name)
	if err != nil {
		log.Println("Can't write file " + file.Name + " in " + *directory)
		wg.Done()
		return
	}
	defer wg.Done()
	defer f.Close()

	f.Write(<-image)
	log.Printf("Saved %s. %d/%d", file.Name, *amount, *count)
	*amount++

}

func getMediaData(path string, image chan []byte) {
	resp, err := http.Get("https://2ch.hk" + path)
	if err != nil {
		log.Println("Can't get image " + path)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Can't read response body")
			return
		}
		image <- body
	}

}

type Page struct {
	Threads []struct {
		Posts []struct {
			Files []File
		} `json:"posts"`
	} `json:"threads"`
}

type File struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
