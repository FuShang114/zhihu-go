package main

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	users  = map[string]string{}
	nextId int64
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/register", register)
	http.ListenAndServe(":8001", nil)
	//TIP <p>Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined text
	// to see how GoLand suggests fixing the warning.</p><p>Alternatively, if available, click the lightbulb to view possible fixes.</p>

}

func ping(w http.ResponseWriter, r *http.Request) {
	var resp response
	resp.Code = 0
	resp.Msg = "pong"
	resp.Data = "pong"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(response{Code: 10001, Msg: "method not allowed"})
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(response{Code: 10001, Msg: "参数错误"})
		return
	}

	if req.Username == "" || req.Password == "" {
		json.NewEncoder(w).Encode(response{Code: 10001, Msg: "username or password is empty"})
		return
	}

	if _, ok := users[req.Username]; ok {
		json.NewEncoder(w).Encode(response{Code: 20001, Msg: "username already exists"})
		return
	}

	nextId++
	users[req.Username] = req.Password
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Code: 0, Msg: "success", Data: map[string]int64{"id": nextId}})
}
