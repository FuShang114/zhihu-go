package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
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

type userAccount struct {
	Username string `json:"username"`
	Password string `json:"-"`
	UserId   int64  `json:"userId"`
}

var nextId int64 = 0

var users = map[string]userAccount{}

// 本地锁
var mu sync.Mutex

var tokenSecret string = "dev-secret-change-me"

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/me", me)
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
	//上本地锁
	mu.Lock()
	defer mu.Unlock()
	if _, ok := users[req.Username]; ok {
		json.NewEncoder(w).Encode(response{Code: 20001, Msg: "username already exists"})
		return
	}

	nextId++
	var user userAccount
	user.UserId = nextId
	user.Username = req.Username
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		json.NewEncoder(w).Encode(response{Code: 10500, Msg: "服务内部错误"})
		return
	}
	user.Password = string(hash)
	users[req.Username] = user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Code: 0, Msg: "success", Data: map[string]int64{"id": user.UserId}})
}

func login(w http.ResponseWriter, r *http.Request) {
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
	mu.Lock()
	defer mu.Unlock()
	account, ok := users[req.Username]
	if !ok {
		json.NewEncoder(w).Encode(response{Code: 20002, Msg: "username or password is wrong"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(req.Password)); err != nil {
		json.NewEncoder(w).Encode(response{Code: 20002, Msg: "username or password is wrong"})
		return
	}
	token, err := signToken(account.UserId, account.Username)
	if err != nil {
		json.NewEncoder(w).Encode(response{Code: 10500, Msg: "服务内部错误"})
		return
	}
	json.NewEncoder(w).Encode(response{Code: 0, Msg: "success", Data: map[string]any{"token": token, "id": account.UserId}})

}

func me(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		json.NewEncoder(w).Encode(response{Code: 10001, Msg: "method not allowed"})
		return
	}
	auth := r.Header.Get("Authorization")
	const Prefix = "Bearer "
	if !strings.HasPrefix(auth, Prefix) {
		json.NewEncoder(w).Encode(response{Code: 10002, Msg: "未登录或令牌失效"})
		return
	}
	_, userName, ok := verifyToken(strings.TrimPrefix(auth, Prefix))
	if !ok {
		json.NewEncoder(w).Encode(response{Code: 10002, Msg: "未登录或令牌失效"})
		return
	}
	user, ok := users[userName]
	if !ok {
		json.NewEncoder(w).Encode(response{Code: 20002, Msg: "未登录或令牌失效"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Code: 0, Msg: "success", Data: user})

}

func signToken(userId int64, userName string) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(map[string]any{"id": userId, "username": userName, "exp": time.Now().Add(24 * time.Hour).Unix()})
	if err != nil {
		return "", err
	}
	hB64 := base64.RawURLEncoding.EncodeToString(header)
	pB64 := base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(sha256.New, []byte(tokenSecret))
	mac.Write([]byte(hB64 + "." + pB64))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return hB64 + "." + pB64 + "." + sig, nil
}

func verifyToken(token string) (int64, string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, "", false
	}

	hB64, pB64, sig := parts[0], parts[1], parts[2]

	mac := hmac.New(sha256.New, []byte(tokenSecret))
	mac.Write([]byte(hB64 + "." + pB64))
	expectedSig := mac.Sum(nil)

	gotSig, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil || !hmac.Equal(gotSig, expectedSig) {
		return 0, "", false
	}

	payload, err := base64.RawURLEncoding.DecodeString(pB64)
	if err != nil {
		return 0, "", false
	}
	var claims struct {
		Username string `json:"username"`
		UserId   int64  `json:"id"`
		Exp      int64  `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, "", false
	}
	if time.Unix(claims.Exp, 0).Before(time.Now()) {
		return 0, "", false
	}
	return claims.UserId, claims.Username, true
}
