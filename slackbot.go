package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"golang.org/x/net/websocket"
	"sync/atomic"
	"time"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

func ping(ws *websocket.Conn) {
	ticker := time.NewTicker(8 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				log.Println("sending ping to slack api")
				postMessage(ws, Message{atomic.AddUint64(&counter, 1), "json", "", ""})
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()

}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: slackbot slack-bot-token\n")
		os.Exit(1)
	}

	ws, id := slackConnect(os.Args[1])
	log.Println("slackbot ready")
	go ping(ws)
	for {
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			parts := strings.Fields(m.Text)
			if len(parts) == 2 {
				if parts[1] == "weather" {
					go func(m Message) {
						m.Text = weather()
						postMessage(ws, m)
					}(m)
				} else {
					m.Text = fmt.Sprintf("mumble... thought you had something for me ...\n")
					postMessage(ws, m)
				}

			} else {
				m.Text = fmt.Sprintf("sorry, that didn't make any sense to me\n")
				postMessage(ws, m)
			}
		}
	}
}

func weather() string {
	url := "http://api.openweathermap.org/data/2.5/weather?lon=24.93417&lat=60.17556&units=metric&mode=json"
	res, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	body, err := ioutil.ReadAll(res.Body)
	var weather WeatherMessage
	json.Unmarshal(body, &weather)
	return fmt.Sprintf("weather today: %s \n temperature is %.1f degrees \n wind speed is %.1f meters per hour\n", weather.Weather[0].Description, weather.Temperature.Temperature, weather.Wind.Wind)
}

type WeatherMessage struct {
	Weather []Weather
	Temperature Temperature `json:"main"`
	Wind Wind
}

type Weather struct {
	Description string `json:"description"`
}

type Temperature struct {
	Temperature float32 `json:"temp"`
}

type Wind struct {
	Wind float32 `json:"speed"`
}