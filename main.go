package main

import (
	"files_server/dir"
	"log"
	"net/http"
)

func main() {
	http.Handle("/", dir.New())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//тесты на pwd
//почитать про обьекты и глобальные состояния
