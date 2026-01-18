package main

import (
	"flag"
	"log"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/go-vgo/robotgo"
	"golang.design/x/hotkey"
)

// ====== Настройки ======
type Preset struct {
	Name      string
	Speed     float64 // множитель скорости (меньше = медленнее)
	ErrorRate float64 // вероятность ошибки на символ
	Misclicks bool
}

var presets = []Preset{
	{"slow", 1.4, 0.03, true},
	{"normal", 1.0, 0.05, true},
	{"fast", 0.7, 0.08, true},
	{"careful", 1.2, 0.01, true},
}

var (
	presetName string
	speedMul   float64
	errorRate  float64
)

func main() {
	flag.StringVar(&presetName, "preset", "normal", "Preset: slow|normal|fast|careful")
	flag.Float64Var(&speedMul, "speed", 1.0, "Speed multiplier (0.5..2.0). Lower = slower")
	flag.Float64Var(&errorRate, "errors", -1.0, "Error rate (0..0.3). If set overrides preset")
	flag.Parse()

	preset := getPreset(presetName)
	if preset == nil {
		log.Fatalf("Unknown preset: %s", presetName)
	}

	if errorRate >= 0 {
		preset.ErrorRate = clamp(errorRate, 0, 0.3)
	}
	if speedMul > 0 {
		preset.Speed *= speedMul
	}

	log.Printf("Preset: %s | speed=%.2f | errors=%.3f\n", preset.Name, preset.Speed, preset.ErrorRate)

	// hotkey Shift+Space
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModShift}, hotkey.KeySpace)
	if err := hk.Register(); err != nil {
		log.Fatal(err)
	}

	log.Println("Ready. Press Shift+Space to type clipboard content.")

	for {
		<-hk.Keydown()
		text, _ := clipboard.ReadAll()
		text = strings.TrimRight(text, "\r\n")
		if text == "" {
			continue
		}

		// небольшая "человеческая" реакция
		sleep(150, 400, preset.Speed)
		humanType(text, preset)
	}
}

// ====== Human typing engine ======

func humanType(text string, p *Preset) {
	for _, r := range text {
		char := string(r)

		// Если символ не печатается — просто пропускаем
		if !canType(char) {
			continue
		}

		// шанс мисклика
		if p.Misclicks && shouldMisclick(char, p.ErrorRate) {
			if wrong := randomNeighbor(char); wrong != "" {
				// печатаем неправильный
				typeChar(wrong, p.Speed)
				sleep(80, 180, p.Speed)
				// стираем
				robotgo.KeyTap("backspace")
				sleep(60, 140, p.Speed)
				// печатаем правильный
				typeChar(char, p.Speed)
				applyDelay(char, p.Speed)
				continue
			}
		}

		typeChar(char, p.Speed)
		applyDelay(char, p.Speed)
	}
}

// ====== typing ======

func typeChar(char string, speed float64) {
	// robotgo.TypeStr умеет печатать Unicode
	robotgo.TypeStr(char)
	sleep(20, 60, speed)
}

// ====== delays ======

func applyDelay(char string, speed float64) {
	switch char {
	case " ", "\n":
		sleep(80, 180, speed)
	case ".", ",", "!", "?", ";", ":":
		sleep(180, 450, speed)
	default:
		sleep(35, 95, speed)
	}
}

func sleep(min, max int, speed float64) {
	// speed >1 = медленнее
	d := time.Duration(min+rand.Intn(max-min)) * time.Millisecond
	time.Sleep(time.Duration(float64(d) * speed))
}

// ====== errors ======

func shouldMisclick(char string, errRate float64) bool {
	if len(char) != 1 {
		return false
	}
	r := rune(char[0])
	if !unicode.IsLetter(r) {
		return false
	}
	return rand.Float64() < errRate
}

// ====== neighbors ======

var neighbors = map[string][]string{
	// латиница
	"q": {"w", "a"},
	"w": {"q", "e", "s"},
	"e": {"w", "r", "d"},
	"r": {"e", "t", "f"},
	"t": {"r", "y", "g"},
	"y": {"t", "u", "h"},
	"u": {"y", "i", "j"},
	"i": {"u", "o", "k"},
	"o": {"i", "p", "l"},
	"p": {"o", "l"},
	"a": {"q", "s", "z"},
	"s": {"a", "d", "w", "x"},
	"d": {"s", "f", "e", "c"},
	"f": {"d", "g", "r", "v"},
	"g": {"f", "h", "t", "b"},
	"h": {"g", "j", "y", "n"},
	"j": {"h", "k", "u", "m"},
	"k": {"j", "l", "i"},
	"l": {"k", "o", "p"},
	"z": {"a", "x"},
	"x": {"z", "c", "s"},
	"c": {"x", "v", "d"},
	"v": {"c", "b", "f"},
	"b": {"v", "n", "g"},
	"n": {"b", "m", "h"},
	"m": {"n", "j"},

	// кириллица (русская раскладка)
	"й": {"ц", "ф"},
	"ц": {"й", "у", "ы"},
	"у": {"ц", "к", "ы", "г"},
	"к": {"у", "е", "г", "н"},
	"е": {"к", "н", "р"},
	"н": {"е", "г", "р", "м"},
	"г": {"у", "к", "н", "ш"},
	"ш": {"г", "щ", "з"},
	"щ": {"ш", "з", "х"},
	"з": {"щ", "х", "ъ"},
	"х": {"щ", "з", "ъ"},
	"ъ": {"х", "з"},
	"ф": {"й", "ы", "в"},
	"ы": {"ф", "в", "у", "ц"},
	"в": {"ф", "ы", "а", "п"},
	"а": {"в", "п", "с", "я"},
	"п": {"в", "а", "р"},
	"р": {"п", "а", "о", "л"},
	"о": {"р", "л", "д"},
	"л": {"о", "д", "ж"},
	"д": {"л", "ж", "э"},
	"ж": {"д", "э"},
	"э": {"ж", "д"},
	"я": {"а", "с"},
	"с": {"я", "а", "м", "и"},
	"м": {"с", "и", "т"},
	"и": {"м", "т", "ь"},
	"т": {"и", "ь", "б"},
	"ь": {"т", "б", "ю"},
	"б": {"ь", "ю"},
	"ю": {"б", "ь"},
}

func randomNeighbor(char string) string {
	char = strings.ToLower(char)
	if arr, ok := neighbors[char]; ok {
		return arr[rand.Intn(len(arr))]
	}
	return ""
}

// ====== misc ======

func canType(char string) bool {
	// robotgo печатает Unicode, поэтому можно печатать всё
	return len(char) > 0
}

func getPreset(name string) *Preset {
	for _, p := range presets {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
