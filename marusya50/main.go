package main

import (
	"fmt"
	"github.com/SevereCloud/vksdk/marusia"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

const (
	UP    int = 0
	DOWN      = 1
	LEFT      = 2
	RIGHT     = 3
)

type User struct {
	playing bool
	field   [][]int
	score   int
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
			field := make([][]int, 4)
			for i := 0; i < 4; i++ {
				field[i] = make([]int, 4)
			}
			users[r.Session.UserID+r.Session.SessionID] = &User{field: field, playing: false}
		}

		user := users[r.Session.UserID+r.Session.SessionID]

		needToGen := false

		if !user.playing {
			resp.Text = "Привет! Не хотите сыграть в 2048?"
			resp.TTS = resp.Text

			if r.Request.Command == "хочу" || r.Request.Command == "да" {

			} else if r.Request.Command == "нет" {
				resp.Text = "А вот сейчас обидно было :("
				resp.TTS = "А вот сейчас обидно было :("
				return
			} else {
				resp.AddButton("Хочу", nil)
				return
			}
			user.score = 0

			user.playing = true
			r.Request.Command = ""

			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					user.field[i][j] = 0
				}
			}

			needToGen = true
		}

		scoreAdded := 0
		moveFailed := false

		if r.Request.Command == "вверх" {
			resp.TTS = "Двигаю вверх"
			moveFailed, scoreAdded = moveField(UP, &user.field, true)
			needToGen = true
		} else if r.Request.Command == "вниз" {
			resp.TTS = "Двигаю вниз"
			moveFailed, scoreAdded = moveField(DOWN, &user.field, true)
			needToGen = true
		} else if r.Request.Command == "вправо" || r.Request.Command == "справа" || r.Request.Command == "право" {
			resp.TTS = "Двигаю вправо"
			moveFailed, scoreAdded = moveField(RIGHT, &user.field, true)
			needToGen = true
		} else if r.Request.Command == "влево" || r.Request.Command == "слева" || r.Request.Command == "лево" {
			resp.TTS = "Двигаю влево"
			moveFailed, scoreAdded = moveField(LEFT, &user.field, true)
			needToGen = true
		} else {
			resp.TTS = "Вывожу поле"
		}

		if moveFailed {
			resp.TTS = "Ход в этом направлении сделать не выйдет"
			needToGen = false
		}

		user.score += scoreAdded

		x, y, ok := getRandomEmptyCell(user.field)
		if ok && needToGen {
			rand.Seed(time.Now().UnixNano())
			if rand.Float64() < 0.9 {
				user.field[x][y] = 2
			} else {
				user.field[x][y] = 4
			}
		}

		lose := checkLose(user.field)
		if lose {
			resp.Text = fmt.Sprintf("Вы набрали %d очков! Игра будет перезапущена автоматически после любого ответа!", user.score)
			resp.TTS = "Игра окончена!"
			user.playing = false
			return
		}

		resp.Text = fmt.Sprintf("Счёт: %d\n", user.score) + printField(user.field)
		resp.AddButton("Вверх", nil)
		resp.AddButton("Вниз", nil)
		resp.AddButton("Влево", nil)
		resp.AddButton("Вправо", nil)

		return
	})

	http.HandleFunc("/", wh.HandleFunc)

	http.ListenAndServe(":80", nil)
}

func checkLose(field [][]int) bool {
	copied := make([][]int, 4)

	for i := 0; i < 4; i++ {
		copied[i] = make([]int, 4)

		for j := 0; j < 4; j++ {
			copied[i][j] = field[i][j]
		}
	}

	tryDown, _ := moveField(DOWN, &copied, false)

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			copied[i][j] = field[i][j]
		}
	}

	tryUp, _ := moveField(UP, &copied, false)

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			copied[i][j] = field[i][j]
		}
	}

	tryLeft, _ := moveField(LEFT, &copied, false)

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			copied[i][j] = field[i][j]
		}
	}

	tryRight, _ := moveField(RIGHT, &copied, false)

	return tryRight && tryLeft && tryDown && tryUp
}

func printField(field [][]int) string {
	result := ""
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			num := field[i][j]
			size := int(math.Log10(float64(num))) + 1
			if num == 0 {
				size = 0
			}
			cell := ""
			if j != 0 {
				cell += "  "
			}
			cell += "["
			if num != 0 {
				cell += fmt.Sprintf("%d", num)
			}
			for k := 0; k < 4-size; k++ {
				cell += "  "
			}

			cell += "]"
			result += cell
		}
		result += "\n"
	}
	return result
}

func getRandomEmptyCell(field [][]int) (int, int, bool) {
	rand.Seed(time.Now().UnixNano())
	candidates := make([]int, 0)

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if field[i][j] == 0 {
				candidates = append(candidates, i+j<<2)
			}
		}
	}

	if len(candidates) == 0 {
		return 0, 0, false
	}

	result := candidates[rand.Intn(len(candidates))]

	return result & 0b11, (result & 0b1100) >> 2, true
}

func moveField(direction int, field *[][]int, logs bool) (loser bool, scoreAdded int) {
	printLog := log.Printf
	if !logs {
		printLog = func(format string, v ...any) {

		}
	}

	scoreAdded = 0

	moved := 0
	if direction == UP {
		printLog("moving up\n")
		for i := 1; i < 4; i++ {
			for j := 0; j < 4; j++ {
				printLog("checking %d %d\n", i, j)
				if (*field)[i][j] == 0 {
					continue
				}

				value := (*field)[i][j]

				lastGoodK := -1
				for k := i - 1; k >= 0; k-- {
					if (*field)[k][j] == 0 {
						lastGoodK = k
						printLog("new lastgoodk reason 0 value %d\n", lastGoodK)
						continue
					} else if (*field)[k][j] == value {
						lastGoodK = k
						printLog("new lastgoodk reason equals value %d\n", lastGoodK)
						break
					} else {
						break
					}
				}

				if lastGoodK == -1 {
					printLog("no lgk, skipping cell\n")
					continue
				} else if (*field)[lastGoodK][j] == 0 {
					(*field)[i][j] = 0
					(*field)[lastGoodK][j] = value
					printLog("found lgk at %d %d, moving from %d %d\n", lastGoodK, j, i, j)
					moved++
				} else {
					(*field)[i][j] = 0
					(*field)[lastGoodK][j] = value * 2
					scoreAdded += value * 2
					printLog("found lgk at %d %d, moving (dbl) from %d %d\n", lastGoodK, j, i, j)
					moved++
				}
			}
		}
	} else if direction == DOWN {
		printLog("moving down\n")
		for i := 2; i >= 0; i-- {
			for j := 0; j < 4; j++ {
				printLog("checking %d %d\n", i, j)
				if (*field)[i][j] == 0 {
					continue
				}

				value := (*field)[i][j]

				lastGoodK := -1
				for k := i + 1; k <= 3; k++ {
					if (*field)[k][j] == 0 {
						lastGoodK = k
						printLog("new lastgoodk reason 0 value %d\n", lastGoodK)
						continue
					} else if (*field)[k][j] == value {
						lastGoodK = k
						printLog("new lastgoodk reason equals value %d\n", lastGoodK)
						break
					} else {
						break
					}
				}

				if lastGoodK == -1 {
					printLog("no lgk, skipping cell\n")
					continue
				} else if (*field)[lastGoodK][j] == 0 {
					(*field)[i][j] = 0
					(*field)[lastGoodK][j] = value
					printLog("found lgk at %d %d, moving from %d %d\n", lastGoodK, j, i, j)
					moved++
				} else {
					(*field)[i][j] = 0
					(*field)[lastGoodK][j] = value * 2
					scoreAdded += value * 2
					printLog("found lgk at %d %d, moving (dbl) from %d %d\n", lastGoodK, j, i, j)
					moved++
				}
			}
		}
	} else if direction == LEFT {
		printLog("moving left\n")
		for j := 1; j < 4; j++ {
			for i := 0; i < 4; i++ {
				printLog("checking %d %d\n", i, j)
				if (*field)[i][j] == 0 {
					continue
				}

				value := (*field)[i][j]

				lastGoodK := -1
				for k := j - 1; k >= 0; k-- {
					if (*field)[i][k] == 0 {
						lastGoodK = k
						printLog("new lastgoodk reason 0 value %d\n", lastGoodK)
						continue
					} else if (*field)[i][k] == value {
						lastGoodK = k
						printLog("new lastgoodk reason equals value %d\n", lastGoodK)
						break
					} else {
						break
					}
				}

				if lastGoodK == -1 {
					printLog("no lgk, skipping cell\n")
					continue
				} else if (*field)[i][lastGoodK] == 0 {
					(*field)[i][j] = 0
					(*field)[i][lastGoodK] = value
					printLog("found lgk at %d %d, moving from %d %d\n", lastGoodK, j, i, j)
					moved++
				} else {
					(*field)[i][j] = 0
					(*field)[i][lastGoodK] = value * 2
					scoreAdded += value * 2
					printLog("found lgk at %d %d, moving (dbl) from %d %d\n", lastGoodK, j, i, j)
					moved++
				}
			}
		}
	} else if direction == RIGHT {
		printLog("moving right\n")
		for j := 2; j >= 0; j-- {
			for i := 0; i < 4; i++ {
				printLog("checking %d %d\n", i, j)
				if (*field)[i][j] == 0 {
					continue
				}

				value := (*field)[i][j]

				lastGoodK := -1
				for k := j + 1; k <= 3; k++ {
					if (*field)[i][k] == 0 {
						lastGoodK = k
						printLog("new lastgoodk reason 0 value %d\n", lastGoodK)
						continue
					} else if (*field)[i][k] == value {
						lastGoodK = k
						printLog("new lastgoodk reason equals value %d\n", lastGoodK)
						break
					} else {
						break
					}
				}

				if lastGoodK == -1 {
					printLog("no lgk, skipping cell\n")
					continue
				} else if (*field)[i][lastGoodK] == 0 {
					(*field)[i][j] = 0
					(*field)[i][lastGoodK] = value
					printLog("found lgk at %d %d, moving from %d %d\n", lastGoodK, j, i, j)
					moved++
				} else {
					(*field)[i][j] = 0
					(*field)[i][lastGoodK] = value * 2
					scoreAdded += value * 2
					printLog("found lgk at %d %d, moving (dbl) from %d %d\n", lastGoodK, j, i, j)
					moved++
				}
			}
		}
	}

	if !logs {
		fmt.Printf("Moved %d dir %d", moved, direction)
	}
	return moved == 0, scoreAdded
}
