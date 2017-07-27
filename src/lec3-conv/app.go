package main

import (
	"flag"
	"image"
	"log"
	"path"
	"runtime"
	"sync"
)

//-----------------------------------------------------------------------------
// Work
//-----------------------------------------------------------------------------

// Work contains an information of image file to process
type Work struct {
	dir      string
	filename string
	quit     bool
}

// Worker is channel of Work
type Worker struct {
	workChan <-chan Work
}

func collectImages(workChan chan<- Work, finChan chan<- bool, srcDir string) {
	defer func() {
		finChan <- true
	}()

	// List image files
	files, err := ListImages(srcDir)
	if err != nil {
		log.Println(err)
		return
	}

	// add works
	for _, file := range files {
		workChan <- Work{srcDir, file.Name(), false}
	}
}

func work(worker Worker, config *Config, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	for {
		work := <-worker.workChan
		if work.quit {
			break
		}

		log.Printf("[R] %v\n", work.filename)

		src, err := LoadImage(path.Join(work.dir, work.filename))
		if err != nil {
			log.Printf("Error : %v : %v\n", work.filename, err)
			continue
		}

		// TODO: process image
		var destImg image.Image

		// resize
		destImg = ResizeImage(src, config.width, config.height)

		// save dest Image
		err = SaveJpeg(destImg, config.destDir, work.filename, 80)
		if err != nil {
			log.Printf("Error : %v : %v\n", work.filename, err)
			continue
		}
	}
}

func main() {
	cfgFilename := flag.String("cfg", "", "configuration filename")
	srcDir := flag.String("src", "", "source directory")
	destDir := flag.String("dest", "", "dest directory")
	flag.Parse()

	// Print usage
	if flag.NFlag() == 1 && flag.Arg(1) == "help" {
		flag.Usage()
		return
	}

	// create Config
	config := NewConfig(*cfgFilename, *srcDir, *destDir)
	config.Print()

	// set maxProcess
	runtime.GOMAXPROCS(config.maxProcess)

	// Create channels
	workChan := make(chan Work, 100)
	finChan := make(chan bool)

	// WaitGroup
	wg := sync.WaitGroup{}

	// start collector
	go collectImages(workChan, finChan, config.srcDir)

	// start workers
	for i := 0; i < config.maxProcess; i++ {
		worker := Worker{workChan}
		wg.Add(1)
		go work(worker, config, &wg)
	}

	// wait for collector finish
	<-finChan

	// finish workers
	for i := 0; i < config.maxProcess; i++ {
		workChan <- Work{"", "", true}
	}

	wg.Wait()
}
