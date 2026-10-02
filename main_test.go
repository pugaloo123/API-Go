package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSuccessCreateItem(t *testing.T) {
	resetStorage()

	data := strings.NewReader(`{"name":"Стул","price":1000}`)
	request := httptest.NewRequest(http.MethodPost, "/items", data)
	w := httptest.NewRecorder()
	handleCreateItem(w, request)

	if w.Code != http.StatusCreated {
		t.Fatalf("Неверный статус %d, должен быть %d. Тело: %s",
			w.Code, http.StatusCreated, w.Body.String())
	}

	var item Item
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatalf("Не удалось разобрать тело ответа: %v", err)
	}

	if item.Name != "Стул" {
		t.Errorf("Неверное имя элемента: %s, должно быть: Стул", item.Name)
	}
	if item.Price != 1000 {
		t.Errorf("Неверная цена элемента: %d, должна быть: 1000", item.Price)
	}
	if item.ID == "" {
		t.Errorf("Ожидалось, что у элемента будет ID, но его нет")
	}
}

func TestCreateItemWrongMethod(t *testing.T) {
	resetStorage()

	data := strings.NewReader(`{"name":"Стул","price":1000}`)
	request := httptest.NewRequest(http.MethodGet, "/items", data)
	w := httptest.NewRecorder()
	handleCreateItem(w, request)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("неверный статус %d, должен быть %d. Тело: %s",
			w.Code, http.StatusMethodNotAllowed, w.Body.String())
	}

	if len(items) != 0 {
		t.Errorf("объявление создано, хотя метод не поддерживается: %+v", items)
	}

	if got := w.Header().Get("Allow"); got != http.MethodPost {
		t.Errorf("заголовок Allow: ожидали %q, получили %q", http.MethodPost, got)
	}

}

func TestCreateItemBadBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"не json", "не json"},
		{"Пустое тело", ""},
		{"Массив вместо объекта", "[1, 2, 3]"},
		{"Число вместо объекта", "123"},
		{"Цена строкой", `{"name":"Стул","price":"1000"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStorage()

			data := strings.NewReader(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/items", data)
			w := httptest.NewRecorder()
			handleCreateItem(w, request)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("неверный статус %d, должен быть %d. Тело: %s",
					w.Code, http.StatusBadRequest, w.Body.String())
			}

			if len(items) != 0 {
				t.Errorf("объявление создано, хотя тело некорректно: %+v", items)
			}

		})
	}
}

func TestCreateItemInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantMSG string
	}{
		{"пустой объект", "{}", "name"},
		{"пустое имя", `{"name":"","price":1000}`, "name"},
		{"имя из пробелов", `{"name":"   ","price":1000}`, "name"},
		{"цена у гранницы", `{"name":"Стул","price":-1}`, "price"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStorage()

			data := strings.NewReader(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/items", data)
			w := httptest.NewRecorder()
			handleCreateItem(w, request)

			if w.Code != http.StatusUnprocessableEntity {
				t.Errorf("неверный статус %d, должен быть %d. Тело: %s",
					w.Code, http.StatusUnprocessableEntity, w.Body.String())
			}

			if len(items) != 0 {
				t.Errorf("объявление создано, хотя тело имеет невалидные данные: %+v", items)
			}

			if !strings.Contains(w.Body.String(), tt.wantMSG) {
				t.Errorf("неверное сообщение об ошибке: ожидали упоминание %q, получили %q", tt.wantMSG, w.Body.String())
			}

		})
	}

}

func TestCreateItemFreePrice(t *testing.T) {
	resetStorage()

	data := strings.NewReader(`{"name":"Стул","price":0}`)
	request := httptest.NewRequest(http.MethodPost, "/items", data)
	w := httptest.NewRecorder()
	handleCreateItem(w, request)

	if w.Code != http.StatusCreated {
		t.Fatalf("Неверный статус %d, должен быть %d. Тело: %s",
			w.Code, http.StatusCreated, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), `"price":0`) {
		t.Errorf("в ответе нет поля price со значением 0: %s", w.Body.String())
	}

	var item Item
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatalf("Не удалось разобрать тело ответа: %v", err)
	}

	if item.Name != "Стул" {
		t.Errorf("Неверное имя элемента: %s, должно быть: Стул", item.Name)
	}
	if item.Price != 0 {
		t.Errorf("Неверная цена элемента: %d, должна быть: 0", item.Price)
	}
	if item.ID == "" {
		t.Errorf("Ожидалось, что у элемента будет ID, но его нет")
	}

	stored, err := getItem(item.ID)
	if err != nil {
		t.Fatalf("объявление не попало в хранилище: %v", err)
	}
	if stored.Price != 0 {
		t.Errorf("в хранилище цена %d, ожидали 0", stored.Price)
	}
}

func TestGetItemSuccess(t *testing.T) {
	resetStorage()

	createItem("Кресло", 3500) // помеха: ручка должна найти второе, а не отдать первое попавшееся
	second := createItem("Кресло 2", 4100)
	request := httptest.NewRequest(http.MethodGet, "/items/"+second.ID, nil)
	w := httptest.NewRecorder()
	newRouter().ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Fatalf("неверный статус %d, должен быть %d. Тело: %s",
			w.Code, http.StatusOK, w.Body.String())
	}

	var got Item
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("не удалось разобрать тело ответа: %v", err)
	}

	if got != second {
		t.Errorf("не совпадают данные, получили: %+v, ожидали: %+v", got, second)
	}
}

func TestGetItemNotFound(t *testing.T) {
	resetStorage()
	createItem("Кресло", 3500)
	request := httptest.NewRequest(http.MethodGet, "/items/999", nil)
	w := httptest.NewRecorder()
	newRouter().ServeHTTP(w, request)

	if w.Code != http.StatusNotFound {
		t.Fatalf("неверный статус %d, должен быть %d. Тело: %s",
			w.Code, http.StatusNotFound, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "не найдено") {
		t.Errorf("неверное сообщение об ошибке: ожидали упоминание 'не найдено', получили %q", w.Body.String())
	}

}

func TestGetItemWrongMethod(t *testing.T) {
	resetStorage()
	item := createItem("Кресло", 3500)
	request := httptest.NewRequest(http.MethodDelete, "/items/"+item.ID, nil)
	w := httptest.NewRecorder()
	newRouter().ServeHTTP(w, request)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("неверный статус %d, должен быть %d. Тело: %s",
			w.Code, http.StatusMethodNotAllowed, w.Body.String())
	}

	if !strings.Contains(w.Header().Get("Allow"), "GET") {
		t.Errorf("в заголовке ожидали получить разрешенный метод: GET, получили: %q", w.Header().Get("Allow"))
	}

}
