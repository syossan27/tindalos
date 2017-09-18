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

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op.String() == "WRITE" {
					fileName := event.Name
					err := exec.Command("goimports", "-w "+fileName).Run()
					if err != nil {
						log.Println("Error: ", err)
					}
					fmt.Println("Success!!")
				}
			case err := <-watcher.Errors:
				log.Println("Error: ", err)
			}
		}
	}()

	var conf Config
	_, err = toml.DecodeFile("tindalos.toml", &conf)
	if err != nil {
		panic(err)
	}

	for _, confTarget := range conf.Target {
		baseDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		targets, err := filepath.Glob(baseDir + confTarget)
		if err != nil {
			panic(err)
		}
		for _, target := range targets {
			err = watcher.Add(target)
			if err != nil {
				panic(err)
			}
		}
	}

	<-done
}
