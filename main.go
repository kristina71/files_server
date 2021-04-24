package main

import (
	"files_server/auth"
	"files_server/dir"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	http.Handle("/", dir.New())
	http.Handle("/auth", auth.New())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//тесты на pwd
//почитать про обьекты и глобальные состояния
