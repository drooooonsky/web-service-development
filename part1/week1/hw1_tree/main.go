package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func PrintFile(out io.Writer, file os.FileInfo, level int, symbol string, with_tab_symbol *[]string) {
	// строка размера файла
	var size string
	if file.IsDir() {
		size = ""
	} else {
		var s string
		if file.Size() == 0 {
			s = "empty"
		} else {
			s = strconv.Itoa(int(file.Size())) + "b"
		}

		size = "(" + s + ")"
	}

	start_string1 := ""

	for idx, v := range *with_tab_symbol {

		if idx >= level {
			break
		}
		start_string1 += v + "\t"

	}
	start_string := start_string1 + symbol + "───"
	fmt.Fprintln(out, start_string+file.Name(), size)
	// fmt.Fprintf(out)
}

func runTree(out io.Writer, path string, printFiles bool, level int, with_tab_symbols *[]string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	// количество печатаемых файлов в уровне
	count_print_files_in_level := 0
	for _, file := range files {
		// fmt.Println(file.Name())
		if file.IsDir() {
			count_print_files_in_level++
		} else if printFiles && !file.IsDir() {
			count_print_files_in_level++
		}
	}

	// fmt.Println(level, count_print_files_in_level)

	for idx, file := range files {

		symbol := "├"
		if idx == count_print_files_in_level-1 {
			symbol = "└"
			// fmt.Println(level, count_print_files_in_level)
		}

		// если не папка и файлы печатаем то выводим файл
		if printFiles && !file.IsDir() {
			PrintFile(out, file, level, symbol, with_tab_symbols)
		}
		// если папка то выводим папку и пробегаемся по внутренностям
		if file.IsDir() {
			PrintFile(out, file, level, symbol, with_tab_symbols)
			level++
			new_path := filepath.Join(path, file.Name())

			if idx == count_print_files_in_level-1 {
				(*with_tab_symbols)[level-1] = " "
			}

			runTree(out, new_path, printFiles, level, with_tab_symbols)

			level--
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	var level int
	// TODO: fix this bad slice
	with_tab_symbols := []string{"│", "│", "│", "│", "│", "│", "│"}
	runTree(out, path, printFiles, level, &with_tab_symbols)
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

// проблемы
// не проходит тест с файлами хотя есть полное совпадение
// не правильно подбирается symbol в тесте без файлов
