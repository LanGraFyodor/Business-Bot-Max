package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func validateInit(raw, token string, at time.Time) (User, error) {
	fail := errors.New("Откройте приложение из MAX: данные входа недействительны")
	q, e := url.ParseQuery(raw)
	if e != nil || token == "" {
		return User{}, fail
	}
	for _, v := range q {
		if len(v) != 1 {
			return User{}, fail
		}
	}
	signature, e := hex.DecodeString(q.Get("hash"))
	if e != nil || len(signature) != 32 {
		return User{}, fail
	}
	q.Del("hash")
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := []string{}
	for _, k := range keys {
		parts = append(parts, k+"="+q.Get(k))
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(token))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(strings.Join(parts, "\n")))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return User{}, fail
	}
	timestamp, e := strconv.ParseInt(q.Get("auth_date"), 10, 64)
	if e != nil || at.Unix()-timestamp > 3600 || timestamp > at.Unix()+60 {
		return User{}, fail
	}
	var u struct {
		ID       int64  `json:"id"`
		First    string `json:"first_name"`
		Last     string `json:"last_name"`
		Username string `json:"username"`
	}
	if json.Unmarshal([]byte(q.Get("user")), &u) != nil || u.ID == 0 {
		return User{}, fail
	}
	uid := strconv.FormatInt(u.ID, 10)
	name := strings.TrimSpace(u.First + " " + u.Last)
	if name == "" && u.Username != "" {
		name = "@" + u.Username
	}
	if name == "" {
		name = "MAX " + uid
	}
	return User{uid, name, "employee"}, nil
}
func (a *App) registerAccount(u User) error {
	for i, account := range a.state.Accounts {
		if account.ID != u.ID {
			continue
		}
		if account.Name == u.Name {
			return nil
		}
		before := a.snapshot()
		a.state.Accounts[i].Name = u.Name
		for ti := range a.state.Teams {
			for mi := range a.state.Teams[ti].Members {
				if a.state.Teams[ti].Members[mi].ID == u.ID {
					a.state.Teams[ti].Members[mi].Name = u.Name
				}
			}
		}
		return a.commit(before)
	}
	before := a.snapshot()
	a.state.Accounts = append(a.state.Accounts, Account{ID: u.ID, Name: u.Name, Registered: now()})
	return a.commit(before)
}
func (a *App) user(r *http.Request) (User, error) {
	if a.demo {
		u, ok := demoUsers[r.Header.Get("X-Demo-User")]
		if !ok {
			u = demoUsers["1"]
		}
		return u, nil
	}
	u, e := validateInit(r.Header.Get("X-Max-Init-Data"), a.token, time.Now())
	if e != nil {
		return u, e
	}
	return u, nil
}
