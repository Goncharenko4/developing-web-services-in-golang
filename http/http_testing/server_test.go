package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type TestCase struct {
	ID      string
	Result  *CheckoutResult
	IsError bool
}

type CheckoutResult struct {
	Status  int
	Balance int
	Err     string
}

func CheckoutDummy(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("id")
	switch key {
	case "42":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": 200, "balance": 100500}`)
	case "100500":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": 400, "err": "bad_balance"}`)
	case "__broken_json":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": 400`) // ошибка с Json
	case "__internal_error":
		fallthrough // остальные ошибки
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type Cart struct {
	PaymentApiURL string
}

func (c *Cart) Checkout(id string) (*CheckoutResult, error) {
	url := c.PaymentApiURL + "?id=" + id // нет хардкода (нет url конкретного сайта)
	resp, err := http.Get(url)           // идем в банк
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil { // ошибка при чтении данных
		return nil, err
	}

	result := &CheckoutResult{}
	err = json.Unmarshal(data, result)
	if err != nil { // ошибка при распаковке Json
		return nil, err
	}
	return result, nil
}

func TestCartCheckout(t *testing.T) {
	cases := []TestCase{
		TestCase{
			ID: "42", // то, что передаем
			Result: &CheckoutResult{ // то, что должно вернуться
				Status:  200,
				Balance: 100500,
				Err:     "",
			},
			IsError: false, // вернется ли ошибка
		},
		TestCase{
			ID: "100500",
			Result: &CheckoutResult{
				Status:  400,
				Balance: 0,
				Err:     "bad_balance",
			},
			IsError: false,
		},
		TestCase{
			ID:      "__broken_json",
			Result:  nil,
			IsError: true,
		},
		TestCase{
			ID:      "__internal_error",
			Result:  nil,
			IsError: true,
		},
	}

	// эмуляция банковских операций:
	ts := httptest.NewServer(http.HandlerFunc(CheckoutDummy)) // создаем новый тестовый сервер

	for caseNum, item := range cases {
		c := &Cart{
			PaymentApiURL: ts.URL, // передаем адрес платежного шлюза
		}
		result, err := c.Checkout(item.ID)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		}
	}
	ts.Close() // когда много тест-кейсов - необходимо потушить тестовый сервер
}
