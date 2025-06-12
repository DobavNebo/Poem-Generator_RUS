package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
)

// Структура Токена. Необходима для того чтобы не обрабатывать слова каждый раз,
// а сразу иметь связанные с ними данные: само слово, его окончание и ударность каждого из его слов (+ их количество)
type Token struct {
	Word   string
	Suffix string
	Syll   []bool
}

// Структура описывающая конфигурацию самого стиха
type PoemSructure struct {
	Length    int      `json:"Length"`
	Endings   []string `json:"Endings"`
	Syllables []int    `json:"Syllables"`
	Accents   []bool   `json:"Accents"`
}

// Проверка руны на то, что она является гласной (кроме ё)
func isSyl(x rune) bool {
	return slices.Contains(SyllSimp, x)
}

// Функция которая превращает слова в структуру Token
func Tokenize(inword string) Token {
	runes := []rune(inword)
	if len(runes) == 0 {
		return Token{"", "", []bool{}}
	}
	var syllables []bool
	//Истинно пока не найдёт первую гласную
	first := true
	suff := ""
	for i := (len(runes) - 1); i >= 0; {
		// Собираем окончание, пока не наткнёмся на гласную. Когда наткнёмся - first станет false и сборка прекратится
		if first {
			if runes[i] == '\u0301' {
				str := string(runes[i-1])
				str += string(runes[i])
				suff = string(str) + suff
			} else {
				suff = string(runes[i]) + suff
			}
		}
		//"ё" нужно обработать отдельно, так как среди безударных она не подходит из-за ударения, а среди ударных из-за того, что записывается без отдельного акута
		if runes[i] == 'ё' {
			syllables = append(syllables, true)
			if first {
				first = false
			}
			i--
			continue
		}
		// Нашли акут
		if runes[i] == '\u0301' {
			// Проверка действительно ли акут образует с предыдущей буквой ударную гласную. Если да - добавляем слог с true
			if isSyl(runes[i-1]) {
				syllables = append(syllables, true)
				if first {
					first = false
				}
			}
			// -2 т.к. рассмотрели не только акут, но и идущую перед ним букву
			i -= 2
			continue
		}
		// Безударная гласная - слог с "false"
		if isSyl(runes[i]) {
			syllables = append(syllables, false)
			if first {
				first = false
			}
			i--
			continue
		}
		i--
	}

	// Если у слова нет слогов, то нам не нужно знать его окончание
	// Потому что такое слово все равно не подходит для генерации рифм
	// И в строке оно никогда не будет последним
	if len(syllables) == 0 {
		suff = ""
	}
	//Мы собирали слоги справа налево, поэтому массив нужно развернуть в другую сторону
	slices.Reverse(syllables)
	return Token{inword, suff, syllables}
}

// Подгрузка и токенизация словаря формата .txt
func LoadDictionary(chosen string) (tokens []Token) {
	file, err := os.Open("Словари/" + chosen + ".txt")
	if err != nil {
		fmt.Println("Словарь " + chosen + ".txt не может быть загружен, выключение программы.")
		log.Fatal(err)
	}
	defer file.Close()

	//Построчно читаем словарь
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := Tokenize(scanner.Text())
		if t.Word != "" {
			tokens = append(tokens, t)
		}
	}

	if err = scanner.Err(); err != nil {
		fmt.Println("Ошибка чтения словаря " + chosen + ".txt Выключение программы.")
		log.Fatal(err)
	} else {
		fmt.Println("Словарь " + chosen + ".txt успешно загружен")
	}

	return tokens
}

// Подгрузка .json конфигурации стиха и возвращение в читаемом для программы виде
func LoadStructure(chosen string) (ps PoemSructure) {
	fileContent, err := os.ReadFile("Структуры/" + chosen + ".json")
	if err != nil {
		fmt.Println("Структура " + chosen + ".json не может быть загружена, выключение программы.")
		log.Fatal(err)
	}

	err = json.Unmarshal(fileContent, &ps)
	if err != nil {
		fmt.Println("Ошибка чтения структуры " + chosen + ".json Выключение программы.")
		log.Fatal(err)
	} else {
		fmt.Println("Структура " + chosen + ".json успешно загружена")
	}

	// Проверка на ошибки настройки структуры
	if ps.Length <= 0 {
		ps.Length = 1
		fmt.Println("Структура содержит некорректный параметр Length. При чтении был изменён до 1.")
		fmt.Println("Рекомендуется исправить ошибку в файле структуры.")
	}
	if len(ps.Endings) == 0 {
		ps.Endings = append(ps.Endings, "")
		fmt.Println("Структура содержит некорректный параметр Endings. При чтении было добавлено игнорирование рифмы.")
		fmt.Println("Рекомендуется исправить ошибку в файле структуры.")
	}
	if len(ps.Syllables) == 0 {
		ps.Syllables = append(ps.Syllables, 0)
		fmt.Println("Структура содержит некорректный параметр Syllables. При чтении было добавлена генерация единственного слова.")
		fmt.Println("Рекомендуется исправить ошибку в файле структуры.")
	}
	for i := 0; i < len(ps.Syllables); i++ {
		if ps.Syllables[i] < 0 {
			ps.Syllables[i] = 0
		}
	}
	if len(ps.Accents) == 0 {
		ps.Syllables = append(ps.Syllables, 0)
		fmt.Println("Структура содержит некорректный параметр Accents. При чтении было добавлено игнорирование ударений.")
		fmt.Println("Рекомендуется исправить ошибку в файле структуры.")
	}

	return ps
}
