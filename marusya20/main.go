package main

import (
	"encoding/json"
	"fmt"
	"github.com/SevereCloud/vksdk/marusia"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type User struct {
	row              int
	playing          bool
	isLastWordEdible bool
	lastWord         string
}

var users map[string]*User

func main() {
	plan, _ := ioutil.ReadFile("words.json")
	var data []struct {
		Word   string `json:"word"`
		Edible bool   `json:"edible"`
	}
	json.Unmarshal(plan, &data)

	edible := 0
	inedible := 0
	for _, datum := range data {
		if datum.Edible {
			edible++
		} else {
			inedible++
		}
	}

	fmt.Printf("%d %d", edible, inedible)

	users = make(map[string]*User)
	type myPayload struct {
		Text string
		marusia.DefaultPayload
	}

	wh := marusia.NewWebhook()

	wh.OnEvent(func(r marusia.Request) (resp marusia.Response) {
		if _, ok := users[r.Session.UserID+r.Session.SessionID]; !ok {
			users[r.Session.UserID+r.Session.SessionID] = &User{playing: false, row: 0}
		}

		user := users[r.Session.UserID+r.Session.SessionID]

		if user.playing {
			if r.Request.Command == "съем" || r.Request.Command == "выброшу" {
				if r.Request.Command == "съем" && user.isLastWordEdible || r.Request.Command == "выброшу" && !user.isLastWordEdible {
					user.row++
				} else {
					resp.Text = fmt.Sprintf("Неправильно! Ваш счет: %d. Не хотите ли сыграть ещё раз?", user.row)
					resp.TTS = resp.Text
					resp.AddButton("Хочу", nil)
					resp.AddButton("Нет", nil)
					user.row = 0
					user.playing = false
					user.lastWord = ""
					return
				}

				rand.Seed(time.Now().UnixNano())

				randomWord := data[rand.Intn(len(data))]
				user.isLastWordEdible = randomWord.Edible
				user.lastWord = randomWord.Word

				resp.Text += fmt.Sprintf("Правильно! Съешь или выбросишь %s?", randomWord.Word)
				resp.TTS = resp.Text
				resp.Text += fmt.Sprintf(" Текущий счет: %d", user.row)
				resp.AddButton("Съем", nil)
				resp.AddButton("Выброшу", nil)
			} else {
				resp.Text = fmt.Sprintf("Не поняла вас. Съедите или выбросите %s?", user.lastWord)
				resp.TTS = resp.Text
			}
		} else {
			if r.Request.Command == "да" || r.Request.Command == "хочу" {
				rand.Seed(time.Now().UnixNano())
				randomWord := data[rand.Intn(len(data))]
				user.isLastWordEdible = randomWord.Edible
				user.lastWord = randomWord.Word
				user.playing = true

				resp.Text = fmt.Sprintf("Съешь или выбросишь %s?", randomWord.Word)
				resp.AddButton("Съем", nil)
				resp.AddButton("Выброшу", nil)
				resp.TTS = resp.Text
			} else if r.Request.Command == "нет" || r.Request.Command == "не хочу" {
				resp.Text = "А вот сейчас обидно было :("
				resp.TTS = resp.Text
			} else {
				resp.Text = "Не хотите ли сыграть в съедобное-несъедобное?"
				resp.TTS = resp.Text
				resp.AddButton("Хочу", nil)
				resp.AddButton("Нет", nil)
			}
		}

		return
	})

	http.HandleFunc("/", wh.HandleFunc)

	http.ListenAndServe(":80", nil)
}
