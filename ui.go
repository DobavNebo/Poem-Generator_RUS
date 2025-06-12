package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Глобальный массив гласных.
// Инициализируется глобально, а не в соотетствующей функции, в связи с тем, что
// если его инициализировать при каждой проверке, то это займёт куда больше времени,
// чем работать с существующим массивом.
// В то же время, пересылать его между функциями - нецелесообразно усложняет читаемость кода.
var SyllSimp []rune

// Главное меню
func MainMenu(dictionary string, instructure string) {
	//Инициализация программы + главное меню
	SyllSimp = []rune{'а', 'у', 'о', 'и', 'э', 'ы', 'я', 'ю', 'е'}

	words := LoadDictionary(dictionary)
	structure := LoadStructure(instructure)

	for {
		menuItems := []string{"Сгенерировать стихотворение",
			("Выбрать словарь | Текущий: " + dictionary + ".txt"),
			("Выбрать конструкцию стихотворения | Текущая: " + instructure + ".json"),
			"Закончить работу"}
		printMenu(menuItems)

		choice := getUserInput("Выберите нужное действие: ")
		index, err := strconv.Atoi(choice)

		if err != nil || index < 1 || index > len(menuItems) {
			fmt.Println("Невозможный выбор, попробуйте еще раз")
			continue
		}

		switch index {
		case 1:
			fmt.Println("")
			fmt.Println("Стихотворение:")
			fmt.Println("")
			GenPoem(words, structure)
		case 2:
			dictionary = getUserInput("Введите название словаря (без .txt): ")
			words = LoadDictionary(dictionary)
		case 3:
			instructure = getUserInput("Введите название структуры (без .json): ")
			structure = LoadStructure(instructure)
		case 4:
			fmt.Println("Завершение работы...")
			return
		}
	}
}

// Выводит функции меню на экран
func printMenu(items []string) {
	fmt.Println("\nМеню:")
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item)
	}
}

// Получение ввода от пользователя
func getUserInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
