package main

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/lrstanley/go-ytdlp"
	"gopkg.in/yaml.v3"
)

//go:embed channels.yaml
var file []byte

type Channel struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Directory string `yaml:"directory"`
}

type Targets struct {
	Channels []Channel
}

func (t *Targets) getChannels() []Channel {
	var config []Channel
	err := yaml.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func (c *Channel) existsDir() bool {
	info, err := os.Stat(c.Directory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	if !info.IsDir() {
		fmt.Fprintln(os.Stderr, c.Directory+" is not directory")
		return false
	}
	fmt.Fprintln(os.Stdout, c.Directory+" is existed")
	return true
}

func (c *Channel) createDir() {
	error := os.Mkdir(c.Directory, 0750)
	if error != nil && !os.IsExist(error) {
		fmt.Fprintln(os.Stderr, error)
	}
	fmt.Fprintln(os.Stdout, "Created "+c.Directory)
}

func (c *Channel) download(wg *sync.WaitGroup) {
	if !c.existsDir() {
		c.createDir()
	}

	dl := ytdlp.New().
		//FormatSort("asr,abr,res,ext:webm:ogg").
		Format("bestvideo+251/140/139").
		EmbedThumbnail().
		EmbedMetadata().
		CookiesFromBrowser("chrome").
		Paths(c.Directory).
		DownloadArchive(filepath.Join(c.Directory, "downloaded")).
		Output("%(upload_date)s_%(title)s.%(ext)s").
		SetExecutable("/usr/local/bin/yt-dlp")

	res, err := dl.Run(context.TODO(), c.URL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, res.Stderr)
	} else {
		fmt.Fprintln(os.Stdout, res.Stdout)
	}

	wg.Done()
}

func main() {
	var wg sync.WaitGroup

	t := Targets{}
	channels := t.getChannels()
	for _, channel := range channels {
		wg.Add(1)
		go channel.download(&wg)
	}
	wg.Wait()
}
