package main

import (
	"fmt"
	"github.com/SevereCloud/vksdk/marusia"
	"math/rand"
	"net/http"
	"time"
)

const (
	Spade   int = 0 << 4
	Heart       = 1 << 4
	Diamond     = 2 << 4
	Club        = 3 << 4
)

const (
	Jack  int = 2
	Queen     = 3
	King      = 4
	Six       = 6
	Seven     = 7
	Eight     = 8
	Nine      = 9
	Ace       = 11
)

// 00 0000

type User struct {
	playing bool

	userCards  map[int]bool
	adminCards map[int]bool
}

var users map[string]*User

func main() {
	users = make(map[string]*User)
	type myPayload struct {
		Text string
		marusia.DefaultPayload
	}

	wh := marusia.NewWebhook()

	wh.OnEvent(func(r marusia.Request) (resp marusia.Response) {
		if _, ok := users[r.Session.UserID+r.Session.SessionID]; !ok {
			users[r.Session.UserID+r.Session.SessionID] = &User{playing: false, userCards: make(map[int]bool), adminCards: make(map[int]bool)}
		}

		user := users[r.Session.UserID+r.Session.SessionID]

		if user.playing && r.Request.Command == "еще" {
			userCard := getRandomCard(*user)
			user.userCards[userCard] = true

			if getSum(user.userCards) < 21 {
				resp.Text = fmt.Sprintf("Держите %s. Теперь ваши карты: %s. Сумма: %d. Ещё или вскрываемся?",
					CardToText(userCard), AllCardsToText(user.userCards), getSum(user.userCards))
				resp.TTS = resp.Text

				resp.AddButton("Еще", nil)
				resp.AddButton("Вскрываемся", nil)
			} else if getSum(user.userCards) == 21 {
				resp.Text = fmt.Sprintf("Держите %s. Теперь ваши карты: %s. Поздравляю, у вас очко. Хотите сыграть ещё раз?",
					CardToText(userCard), AllCardsToText(user.userCards))
				resp.TTS = resp.Text
				resp.AddButton("Да", nil)
				user.playing = false
			} else {
				resp.Text = fmt.Sprintf("Держите %s. Теперь ваши карты: %s. Сумма: %d. К сожалению, перебор. Сыграем ещё раз?",
					CardToText(userCard), AllCardsToText(user.userCards), getSum(user.userCards))
				resp.TTS = resp.Text
				resp.AddButton("Да", nil)
				user.playing = false
			}

			return
		}

		if user.playing && (r.Request.Command == "вскрываемся" || r.Request.Command == "скрываемся") {
			for {
				adminCard := getRandomCard(*user)
				user.adminCards[adminCard] = true

				sum := getSum(user.adminCards)
				if sum >= 17 {
					break
				}
			}

			sum := getSum(user.adminCards)
			playerSum := getSum(user.userCards)

			if sum > 21 {
				resp.Text = fmt.Sprintf("Мои карты: %s. Сумма: %d. У меня перебор, вы победили. Хотите сыграть ещё раз?", AllCardsToText(user.adminCards), sum)
				resp.TTS = resp.Text
				resp.AddButton("Да", nil)
				user.playing = false
			} else if sum == 21 {
				resp.Text = fmt.Sprintf("Мои карты: %s. У меня очко. Хотите реванш?", AllCardsToText(user.adminCards))
				resp.TTS = resp.Text
				resp.AddButton("Да", nil)
				user.playing = false
			} else if sum > playerSum {
				resp.Text = fmt.Sprintf("Мои карты: %s. Сумма: %d. Я победила. Хотите реванш?", AllCardsToText(user.adminCards), sum)
				resp.TTS = resp.Text
				resp.AddButton("Да", nil)
				user.playing = false
			} else if sum == playerSum {
				resp.Text = fmt.Sprintf("Мои карты: %s. Сумма: %d. Ничья. Ещё партейку?", AllCardsToText(user.adminCards), sum)
				resp.TTS = resp.Text
				resp.AddButton("Да", nil)
				user.playing = false
			} else {
				resp.Text = fmt.Sprintf("Мои карты: %s. Сумма: %d. Поздравляю, вы победили. Хотите сыграть ещё раз?", AllCardsToText(user.adminCards), sum)
				resp.TTS = resp.Text
				resp.AddButton("Да", nil)
				user.playing = false
			}

			return
		}

		if user.playing {
			resp.Text = "Не поняла вас. Ещё или вскрываемся?"
			resp.TTS = resp.Text
			resp.AddButton("Еще", nil)
			resp.AddButton("Вскрываемся", nil)
		}

		if !user.playing && r.Request.Command == "да" {
			user.userCards = make(map[int]bool)
			user.adminCards = make(map[int]bool)
			userCard := getRandomCard(*user)
			user.userCards[userCard] = true
			resp.Text = fmt.Sprintf("Держите %s. Ещё или вскрываемся?", CardToText(userCard))
			resp.TTS = resp.Text
			resp.AddButton("Еще", nil)
			resp.AddButton("Вскрываемся", nil)
			user.playing = true
			return
		}

		if !user.playing && r.Request.Command == "нет" {
			resp.Text = "А вот сейчас обидно было :("
			resp.TTS = "А вот сейчас обидно было"
			return
		}

		if !user.playing {
			resp.Text = "Хотите сыграть в очко?"
			resp.TTS = "Хотите сыграть в очко?"
			resp.AddButton("Да", nil)
			return
		}

		return
	})

	http.HandleFunc("/", wh.HandleFunc)

	http.ListenAndServe(":80", nil)
}

func getRandomCard(user User) int {
	rand.Seed(time.Now().UnixNano())

	allCards := make([]int, 0)
	for i := 0; i < 4; i++ {
		for j := 2; j < 12; j++ {
			if j == 10 || j == 5 {
				continue
			}

			if val, ok := user.userCards[i<<4+j]; ok && val {
				continue
			}

			if val, ok := user.adminCards[i<<4+j]; ok && val {
				continue
			}

			allCards = append(allCards, i<<4+j)
		}
	}

	card := allCards[rand.Intn(len(allCards))]

	return card
}

func getSum(cards map[int]bool) int {
	sum := 0
	for card := range cards {
		if !cards[card] {
			continue
		}
		sum += card & 0b1111
	}
	return sum
}

func AllCardsToText(cards map[int]bool) string {
	text := ""
	first := true
	for card := range cards {
		if !cards[card] {
			continue
		}
		if !first {
			text += ", "
		} else {
			first = false
		}
		text += CardToText(card)
	}
	return text
}

func CardToText(card int) string {
	suit := card & 0b110000
	cardType := card & 0b1111

	text := ""
	switch cardType {
	case Ace:
		text += "туз"
		break
	case Jack:
		text += "валет"
		break
	case Queen:
		text += "королева"
		break
	case King:
		text += "король"
		break
	case Six:
		text += "шесть"
		break
	case Seven:
		text += "семь"
		break
	case Eight:
		text += "восемь"
		break
	case Nine:
		text += "девять"
		break
	}
	text += " "

	switch suit {
	case Spade:
		text += "пик"
		break
	case Heart:
		text += "червей"
		break
	case Diamond:
		text += "бубей"
		break
	case Club:
		text += "треф"
		break
	}

	return text
}
