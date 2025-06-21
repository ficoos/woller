package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/ficoos/woller/wol"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templateFiles embed.FS

type Config struct {
	Devices []DeviceConfig `json:"devices"`
}

func (c *Config) FindDevice(id string) *DeviceConfig {
	for _, device := range c.Devices {
		if device.ID == id {
			return &device
		}
	}

	return nil
}

type DeviceConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	MAC  string `json:"mac"`
}

func readConfiguration(path string) (*Config, error) {
	conf, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %s", err)
	}

	defer conf.Close()
	dec := json.NewDecoder(conf)
	var result Config
	err = dec.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decode: %s", err)
	}

	return &result, nil
}

func main() {
	confPath := os.Getenv("WOLLER_CONFIG")
	if len(confPath) == 0 {
		confPath = "./config.json"
	}
	conf, err := readConfiguration(confPath)
	if err != nil {
		log.Fatalf("read config: %s", err)
	}
	tpl, err := template.ParseFS(templateFiles, "templates/*.tmpl")
	if err != nil {
		log.Fatalf("parse templates: %s", err)
	}

	http.HandleFunc("POST /wakeup/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		dev := conf.FindDevice(id)
		if dev == nil {
			tpl.ExecuteTemplate(w, "result.html.tmpl", fmt.Sprintf("Could not find device %s", id))
			return
		}

		err := wol.Wakup(dev.MAC)
		if err != nil {
			tpl.ExecuteTemplate(w, "result.html.tmpl", fmt.Sprintf("Could wake up device %s", err))
			return
		}

		tpl.ExecuteTemplate(w, "result.html.tmpl", nil)
	})

	http.Handle("GET /{$}", http.RedirectHandler("/index.html", http.StatusPermanentRedirect))
	http.HandleFunc("GET /index.html", func(w http.ResponseWriter, r *http.Request) {
		tpl.ExecuteTemplate(w, "index.html.tmpl", conf)
	})
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("sub static: %s", err)
	}
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	http.ListenAndServe(":8080", nil)
}
