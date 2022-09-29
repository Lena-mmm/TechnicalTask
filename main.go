package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // Символы, из которых создаётся короткая ссылка
	connStr     = "user=postgres password=123654 dbname=DB sslmode=disable"        // Подключение к базе данных
)

func shorting() string { // Создаёт строку в 32 символа, взятых из letterBytes
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func isValidUrl(token string) bool { // Проверка на валидность ссылки
	_, err := url.ParseRequestURI(token)
	if err != nil {
		return false
	}
	u, err := url.Parse(token)
	if err != nil || u.Host == "" {
		return false
	}
	return true
}

type Result struct {
	Link   string
	Short  string
	Status string
}

var arr []Result

func indexPage(w http.ResponseWriter, r *http.Request) { // Связываем короткую и длинную ссылки
	templ, _ := template.ParseFiles("templates/index.html")
	result := Result{}
	if r.Method == "POST" {
		if !isValidUrl(r.FormValue("s")) {
			fmt.Println("Что-то не так")
			result.Status = "Ссылка имеет неправильный формат!"
			result.Link = ""
		} else {
			result.Link = r.FormValue("s")
			result.Short = shorting()

			if os.Args[len(os.Args)-1] == "-d" { // Если нужно сохранить в postgres
				db, err := sql.Open("postgres", connStr) // Устанавливаем соединение с базой данных
				if err != nil {
					panic(err)
				}
				defer db.Close()
				db.Exec(`INSERT INTO public."Table1" (link, short) values ($1, $2)`, result.Link, result.Short) // Помещаем в БД короткую и длинную ссылки
			} else { // Если должно храниться в памяти
				arr = append(arr, result)
				arr[len(arr)-1] = Result{result.Link, result.Short, "Сокращение было выполнено успешно"}
			}

			result.Status = "Сокращение было выполнено успешно"
		}
	}
	templ.Execute(w, result)
}

func redirectTo(w http.ResponseWriter, r *http.Request) {
	var link string
	vars := mux.Vars(r) // Получаем маршрут
	if os.Args[len(os.Args)-1] == "-d" {
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		rows := db.QueryRow(`select link from public."Table1" where short=$1 limit 1`, vars["key"]) // Берём длинную ссылку из БД, которая соответствует ключу
		rows.Scan(&link)
	} else {
		for i := range arr {
			if arr[i].Short == vars["key"] {
				link = arr[i].Link // Берём из среза длинную ссылку, которая соответствует ключу
			}
		}

	}
	fmt.Fprintf(w, "<script>location='%s';</script>", link) // Вставляем длинную ссылку в поисковый запрос
}

func main() {
	arr = make([]Result, 1)                         // Формируем срез структур для хранения ссылок в памяти
	router := mux.NewRouter()                       // Для определения маршрутов
	router.HandleFunc("/", indexPage)               // Мультиплексор маршрута
	router.HandleFunc("/{key}", redirectTo)         //Продолжаем маршрут по /{key}
	log.Fatal(http.ListenAndServe(":8080", router)) //Запускаем веб-сервер
}
