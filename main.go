package main

import (
	// "errors"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var lastID int
var items = make(map[string]Item)

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}


func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/items", handleCreateItem)
	mux.HandleFunc("GET /items/{id}", handleGetItem)
	return mux
}

func handleCreateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req Item
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "некорректное тело запроса", http.StatusBadRequest)
		return
	}

	if req.Price < 0 {
		http.Error(w, "поле price должно быть >= 0", http.StatusUnprocessableEntity)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "поле name не может быть пустым", http.StatusUnprocessableEntity)
		return
	}

	item := createItem(req.Name, req.Price)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func handleGetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	item, err := getItem(id)

	if err != nil {
		http.Error(w, "объявление не найдено", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

func resetStorage() {
	items = make(map[string]Item)
	lastID = 0
}

func createItem(name string, price int) Item {
	lastID++

	item := Item{ID: strconv.Itoa(lastID), Name: name, Price: price}
	items[item.ID] = item
	return item
}

func getItem(id string) (Item, error) {
	value, ok := items[id]
	if !ok {
		return Item{}, fmt.Errorf("объявление с id %s не найдено", id)
	}

	return value, nil

}

func main() {
	http.ListenAndServe(":8080", newRouter())
}
