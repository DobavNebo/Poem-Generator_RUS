package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"slices"
)

type Settings struct {
	Dictionary string `json:"Default dictionary"`
	Structure  string `json:"Default structure"`
}

// Конфигурация для строки, которую мы хотим получить от GenLine и ID этой строки
type Seed struct {
	Id     int
	Syll   []bool
	Ending string
}

// Возвращаемая от GenLine строка со своим id
type MadeLine struct {
	Id   int
	Line string
}

// Возвращает случайный токен из указанного массива
func GenWord(words []Token) Token {
	return words[rand.Intn(len(words))]
}

// Генерирует слово и возвращает его окончание, а также то, является ли это окончание ударным
func GenSyll(words []Token) (string, bool) {
	for {
		cur := GenWord(words)
		if cur.Suffix != "" {
			return cur.Suffix, cur.Syll[len(cur.Syll)-1]
		}
	}
}

// Генерирует строку из указанного массива токенов, настроенного "семени" и передаёт её в указанный канал
func GenLine(words []Token, plan Seed, stream chan MadeLine) {
	for g := 0; g < 100; g++ {
		// Во благо повышения скорости генерации, сначала создаётся подходящий конец строки,
		// и его удобнее хранить в lword и lsyll, а не в общей строке
		var lword string
		var line string
		var syllables []bool
		var lsyl []bool

		// Сначала подбираем последнее слово строки
		for z := 0; ; z++ {
			cur := GenWord(words)
			if EndCheck(cur.Suffix, plan.Ending) && IsEquit(cur.Syll, plan.Syll[(len(plan.Syll)-len(cur.Syll)):]) {
				lword = cur.Word
				lsyl = append(lsyl, cur.Syll...)
				break
			}
			if z >= 1000 {
				stream <- MadeLine{plan.Id, "Ошибка генерации 1 (окончание строки)"}
				return
			}
		}
		// Далее достраиваем остальную часть строки
		// Вся конструкция касающаяся переменной last связана с тем,
		// чтобы программа не использовала бесслоговые предлоги типо "в" в сочетаниях вроде "в в в"
		last := -1
		for j := 0; j < 100; j++ {
			cur := GenWord(words)
			// Подбираем слова, проверяя подходят ли они текущей строке по своим ударным слогам
			if (last != 0 || len(cur.Syll) != 0) && IsEquit(cur.Syll, plan.Syll[len(syllables):]) {
				syllables = append(syllables, cur.Syll...)
				line += cur.Word + " "
				last = len(cur.Syll)
			}
			// Если строка нам подходит - отправляем её в канал
			test := append(syllables, lsyl...)
			if len(test) == len(plan.Syll) && IsEquit(test, plan.Syll) {
				line += lword
				stream <- MadeLine{plan.Id, line}
				return
			}
			// Если строка содержит больше слогов, чем требуется - нет смысла продолжать работать с ней.
			// Прерывание текущего цикла отправит цикл более высокого уровня на следующую итерацию
			if len(test) > len(plan.Syll) {
				break
			}
		}
	}
	// Возврат ошибки чтобы не заблокировать работу программы.
	stream <- MadeLine{plan.Id, "Ошибка генерации 2 (генерация строки)"}
}

func GenPoem(words []Token, structure PoemSructure) {
	var seeds []Seed
	var poem []string
	// Формируем "план" для каждой строки и создаём ей дублёр формата string
	for i := 0; i < structure.Length; i++ {
		var expsyl []bool
		for g := 0; g < structure.Syllables[i%len(structure.Syllables)]; g++ {
			expsyl = append(expsyl, structure.Accents[g%len(structure.Accents)])
		}
		seeds = append(seeds, Seed{i, expsyl, structure.Endings[i%len(structure.Endings)]})
		poem = append(poem, "")
	}

	// Сводим планы строк по окончаниям и задаём окончание, которое ожидается от всех этих строк
	var overs []string
	var accents []bool
	for _, cur := range seeds {
		if (cur.Ending != "") && (!slices.Contains(overs, cur.Ending)) {
			overs = append(overs, cur.Ending)
			accents = append(accents, cur.Syll[(len(cur.Syll)-1)])
		}
	}
	for i, val := range overs {
		var suf string
		var bul bool
		for j := 0; ; j++ {
			suf, bul = GenSyll(words)
			if bul == accents[i] {
				break
			}
			if j >= 100 {
				fmt.Println("Ошибка подбора концов строк")
				return
			}
		}
		for g := 0; g < len(seeds); g++ {
			if seeds[g].Ending == val {
				seeds[g].Ending = suf
			}
		}
	}

	// Запускаем Goroutines
	stream := make(chan MadeLine)
	for _, val := range seeds {
		go GenLine(words, val, stream)
	}

	// Собираем данные из канала и размещаем в соответствующих им местах
	cnt := 0
	leng := len(poem)
	for {
		if cnt >= leng {
			break
		}
		unit := <-stream
		poem[unit.Id] = unit.Line
		cnt++
	}

	speaker(poem)
}

// Проверка соответствия ударности слогов
func IsEquit(x []bool, y []bool) bool {
	// Если мы сверяемся со списком где нет ударных слогов, то считаем,
	// Что при этой генерации нам безразличны ударения
	if !slices.Contains(y, true) {
		return true
	}
	// Если первый список меньше второго - они однозначно не соответствуют
	if len(x) > len(y) {
		return false
	}
	// Здесь роверяется только кусочек по длине x,
	// но не сравнивается что x вообще равен y.
	// Это не совсем корректно в общем случае, но именно в этой программе
	// позволяет упростить работу с подбором слов в строку.
	for i, val := range x {
		if val != y[i] {
			return false
		}
	}
	return true
}

// Проверка сходится ли строка с необходимым окончанием
func EndCheck(x string, y string) bool {
	// Если окончание не имеет значения, проверка выдаст истину
	if y == "" {
		return true
	}
	if x == y {
		return true
	}
	return false
}

func main() {
	//Инициализация настроек
	fileContent, err := os.ReadFile("Настройки.json")
	if err != nil {
		fmt.Println("Стаднартные настройки не могут быть загружены, выключение программы.")
		log.Fatal(err)
	}
	var base Settings
	err = json.Unmarshal(fileContent, &base)
	if err != nil {
		fmt.Println("Ошибка чтения стандартных настроек, выключение программы.")
		log.Fatal(err)
	} else {
		fmt.Println("Стандартные настройки успешно загружены")
	}

	MainMenu(base.Dictionary, base.Structure)
}

// Функция выводит список строк в виде нескольких строк на экране.
func speaker(lines []string) {
	for _, v := range lines {
		fmt.Println(v)
	}
}
