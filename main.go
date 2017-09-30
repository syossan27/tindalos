package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Target []string
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	var conf Config
	_, err = toml.DecodeFile("tindalos.toml", &conf)
	if err != nil {
		panic(err)
	}

	var targets []string
	for _, target := range conf.Target {
		baseDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		ts, err := filepath.Glob(baseDir + target)
		for _, t := range ts {
			targets = append(targets, t)
			if err != nil {
				panic(err)
			}
		}
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}

				fileName := event.Name
				for _, t := range targets {
					if fileName == t {
						err := exec.Command("goimports", "-w", fileName).Run()
						if err != nil {
							log.Println("Error: ", err)
						}
					} else {
						swpFileName := filepath.Dir(t) + "/." + filepath.Base(t) + ".swp"
						if swpFileName == fileName {
							fmt.Println(t)
							err := exec.Command("goimports", "-w", t).Run()
							if err != nil {
								log.Println("Error: ", err)
							}
						}
					}
				}
				fmt.Println("Success!!")
			case err := <-watcher.Errors:
				log.Println("Error: ", err)
			}
		}
	}()

	for _, target := range conf.Target {
		baseDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		err = watcher.Add(filepath.Dir(baseDir + target))
		if err != nil {
			panic(err)
		}
	}

	<-done
}
