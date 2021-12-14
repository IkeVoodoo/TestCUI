package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gdamore/tcell/v2"
	"os"
	"regexp"
	"strings"
	"time"
)

/*
	Made by Simone Coletti
	This is a test for the CUI
        The code is not the best, I plan on improving it
*/

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()

	screen.Clear()
	screen.Show()

	stop := false

	selected := 0

	lastWen := time.Now().UnixMilli()

	texts := []string{"", "", ""}
	textChanged := false
	lastTextChange := time.Now().UnixMilli()

	wrong := []bool{true, true, true}

	width, _ := screen.Size()

	warning := false
	warningMessage := ""
	warningTime := time.Now().UnixMilli()

	lastInput := time.Now().UnixMilli()
	blink := 0

	button(width, false, screen)
	for {
		if stop {
			break
		}
		if warning {
			if time.Now().UnixMilli()-warningTime > 3000 {
				warning = false
				fillBox(width / 2 - len(warningMessage) / 2, 6, len(warningMessage), 3, "", tcell.ColorBlack, tcell.ColorBlack, true, screen)
			} else {
				fillBox(width/2-len(warningMessage)/2, 6, len(warningMessage), 3, warningMessage, tcell.ColorRed, tcell.ColorBlack, true, screen)
			}
			screen.Show()
		}
		inputs(width, selected, texts, wrong, &lastInput, &blink, textChanged, screen)
		if time.Now().UnixMilli() - lastTextChange > 500 {
			textChanged = false
		}
		if !screen.HasPendingEvent() {
			continue
		}
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC:
				stop = true
				break
			case tcell.KeyTab:
				if lastWen + 50 < time.Now().UnixMilli() {
					lastWen = time.Now().UnixMilli()
					selected++
					if selected > 3 {
						selected = 0
					}
					inputs(width, selected, texts, wrong, &lastInput, &blink, false, screen)
					button(width, selected == 3, screen)
				}
				continue
			case tcell.KeyEnter:
				if selected == 3 {
					warningMessage = "Please fix the following errors: "
					if wrong[0] {
						warningMessage += "Username, "
					}
					if wrong[1] {
						warningMessage += "Password, "
					}
					if wrong[2] {
						warningMessage += "Host."
					}
					if warningMessage[len(warningMessage)-2:] == ", " {
						warningMessage = warningMessage[:len(warningMessage)-2] + "."
					}
					shouldStop := false
					for i := 0; i < len(wrong); i++ {
						if wrong[i] {
							warning = true
							warningTime = time.Now().UnixMilli()
							shouldStop = true
							continue
						}
					}
					if shouldStop {
						continue
					}
					stop = true
					writeJsonToFile(texts)
					break
				}
				continue
			case tcell.KeyUp:
				selected--
				if selected < 0 {
					selected = 3
				}
				inputs(width, selected, texts, wrong, &lastInput, &blink, false, screen)
				button(width, selected == 3, screen)
				continue
			case tcell.KeyDown:
				selected++
				if selected > 3 {
					selected = 0
				}
				inputs(width, selected, texts, wrong, &lastInput, &blink, false, screen)
				button(width, selected == 3, screen)
				continue
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if selected < 3 && len(texts[selected]) > 0 {
					texts[selected] = texts[selected][:len(texts[selected])-1]
					textChanged = true
					lastTextChange = time.Now().UnixMilli()
					checkWrong(selected, texts, wrong)
				}
				continue
			default:
				if selected < 3 {
					texts[selected] += string(ev.Rune())
					textChanged = true
					lastTextChange = time.Now().UnixMilli()
					checkWrong(selected, texts, wrong)
				}
				continue
			}
		case *tcell.EventResize:
			screen.Sync()
			screen.Clear()
			width, _ = screen.Size()
			inputs(width, selected, texts, wrong, &lastInput, &blink, false, screen)
			button(width, selected == 3, screen)
			screen.Show()
		}
	}
}

func checkWrong(selected int, texts []string, wrong []bool) {
	if selected == 3 {
		return
	}
	if len(texts[selected]) == 0 {
		wrong[selected] = true
		return
	}
	match, _ := regexp.MatchString("^[a-zA-Z0-9.$&/()\"@%]+$", texts[selected])
	wrong[selected] = !match
}

func writeJsonToFile(texts []string) {
	json := "{\n\t\"username\": \"" + texts[0] + "\",\n\t\"password\": \"" + encryptPass(texts[1]) + "\",\n\t\"hostname\": \"" + texts[2] + "\"\n}"
	bytes := []byte(json)
	err := os.WriteFile("output.json", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func encryptPass(text string) string {
	encrypted := sha256.Sum256([]byte(text))
	return hex.EncodeToString(encrypted[:])
}

func inputs(width, selected int, texts []string, wrong []bool, lastInput *int64, blink *int, sizeChanged bool, screen tcell.Screen) {
	colors := []tcell.Color{tcell.ColorGray, tcell.ColorGray, tcell.ColorGray}
	if selected < 3 {
		colors[selected] = tcell.ColorWhite
		if wrong[selected] {
			colors[selected] = tcell.ColorRed
		}
	}

	now := time.Now().UnixMilli()
	shouldBlink := now - *lastInput > 500
	if shouldBlink {
		*blink++
		if *blink == 3 {
			*blink = 0
		}
		*lastInput = now
	}

	localBlink := *blink

	if sizeChanged {
		localBlink = 0
		*lastInput = time.Now().UnixMilli()
	}

	writeBox(width / 2 - 25 / 2, 10, 25, 3, "Email", texts[0], colors[0], localBlink, selected == 0, screen)
	writeBox(width / 2 - 25 / 2, 13, 25, 3, "Password", strings.Repeat("*", len(texts[1])), colors[1], localBlink, selected == 1, screen)
	writeBox(width / 2 - 25 / 2, 16, 25, 3, "Host", texts[2], colors[2], localBlink, selected == 2, screen)
	screen.Show()
}

func button(width int, buttonSelected bool, screen tcell.Screen) {
	if !buttonSelected {
		writeButton(width / 2 - 5, 19, 10, 3, "OK", tcell.ColorGray, true, screen)
	} else {
		writeButton(width / 2 - 5, 19, 10, 3, "OK", tcell.ColorWhite, true, screen)
	}
	screen.Show()
}

func writeBox(x, y, width, height int, name, contents string, foreground tcell.Color, blink int, selected bool, screen tcell.Screen) {
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(foreground)
	for i := x; i < x + width; i++ {
		for j := y; j < y + height; j++ {
			screen.SetCell(i, j, style, ' ')
		}
	}
	for i := 0; i < width; i++ {
		screen.SetCell(x + i, y, style, '─')
		screen.SetCell(x + i, y + height - 1, style, '─')
	}
	for i := 0; i < height; i++ {
		screen.SetCell(x, y + i, style, '│')
		screen.SetCell(x + width - 1, y + i, style, '│')
	}
	screen.SetCell(x, y, style, '┌')
	screen.SetCell(x, y + height - 1, style, '└')
	screen.SetCell(x + width - 1, y, style, '┐')
	screen.SetCell(x + width - 1, y + height - 1, style, '┘')
	for i := 0; i < len(name); i++ {
		screen.SetCell(x + 1 + i, y, style, int32(name[i]))
	}

	if len(contents) > width - 3 {
		contents = contents[len(contents) - width + 3:]
	}

	for i := 0; i < len(contents); i++ {
		screen.SetCell(x + 1 + i, y + 1, style, int32(contents[i]))
	}

	if len(contents) > 0 || selected {
		if blink == 1 {
			screen.SetCell(x + 1 + len(contents), y + 1, tcell.StyleDefault, ' ')
		} else if blink == 0 {
			screen.SetCell(x + 1 + len(contents), y + 1, style, '|')
		}
	}
}

func fillBox(x int, y int, width int, height int, text string, color tcell.Color, foreground tcell.Color, centered bool, screen tcell.Screen) {
	style := tcell.StyleDefault.Background(color).Foreground(foreground)

	for i := x; i < x + width; i++ {
		for j := y; j < y + height; j++ {
			screen.SetCell(i, j, style, ' ')
		}
	}

	xOffset := 0
	if centered {
		xOffset = (width - len(text)) / 2
	} else {
		xOffset = 1
	}

	for i := 0; i < len(text); i++ {
		screen.SetCell(x + xOffset + i, y + (height / 2), style, int32(text[i]))
	}
}

func writeButton(x int, y int, width int, height int, text string, foreground tcell.Color, centered bool, screen tcell.Screen) {
	writeBox(x, y, width, height, "", "", foreground, 0, false, screen)

	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(foreground)

	xOffset := 0
	if centered {
		xOffset = (width - len(text)) / 2
	} else {
		xOffset = 1
	}

	for i := 0; i < len(text); i++ {
		screen.SetCell(x + xOffset + i, y + (height / 2), style, int32(text[i]))
	}
}
