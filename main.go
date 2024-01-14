package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

var userMap = map[string]UserInfo{} //作为临时的缓存

const (
	githubUri         = "https://github.com/login/oauth/authorize"
	githubAccessToken = "https://github.com/login/oauth/access_token"
	githubUserApi     = "https://api.github.com/user"
	redirectUri       = "http://localhost:8080/token" //地址必须注册到github的配置中
	clientID          = ""                            //TODO 填写自己的clientID
	clientSecret      = ""                            //TODO 填写自己的clientSecret
	sessionKey        = "test"
)

func main() {
	userMap = make(map[string]UserInfo)

	//前端静态文件地址
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	//请求登录接口
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		uri := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s", githubUri, clientID, url.QueryEscape(redirectUri))
		http.Redirect(w, r, uri, http.StatusFound)
	})

	//重定向时根据code获取token接口
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		httpClient := http.Client{}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")

		reqURL := fmt.Sprintf("%s?client_id=%s&client_secret=%s&code=%s", githubAccessToken, clientID, clientSecret, code)
		req, err := http.NewRequest(http.MethodPost, reqURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		req.Header.Set("accept", "application/json")
		res, err := httpClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			_ = res.Body.Close()
		}()

		var t OAuthAccessResponse
		if err = json.NewDecoder(res.Body).Decode(&t); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//set cookie
		cookie, err := genCookie(t.AccessToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: sessionKey, Value: cookie, Path: "/", Domain: "localhost", Expires: time.Now().Add(time.Second * 3600)})

		w.Header().Set("Location", "/index.html")
		w.WriteHeader(http.StatusFound)
	})

	http.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionKey)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bytes, err := json.Marshal(userMap[cookie.Value])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = w.Write(bytes)
	})
	_ = http.ListenAndServe(":8080", nil)
}

//根据token获取userInfo，置换出自定义的cookie
func genCookie(token string) (string, error) {
	httpClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, githubUserApi, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+token)
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	cookie := uuid.NewString()
	var userInfo UserInfo
	if err = json.Unmarshal(bytes, &userInfo); err != nil {
		return "", err
	}

	userMap[cookie] = userInfo
	return cookie, nil
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

type UserInfo struct {
	AvatarUrl         string `json:"avatar_url,omitempty"`
	Bio               string `json:"bio,omitempty"`
	Blog              string `json:"blog,omitempty"`
	Company           string `json:"company,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	Email             string `json:"email,omitempty"`
	EventsUrl         string `json:"events_url,omitempty"`
	Followers         int32  `json:"followers,omitempty"`
	FollowersUrl      string `json:"followers_url,omitempty"`
	Following         int32  `json:"following,omitempty"`
	FollowingUrl      string `json:"following_url,omitempty"`
	GistsUrl          string `json:"gists_url,omitempty"`
	GravatarId        string `json:"gravatar_id,omitempty"`
	HtmlUrl           string `json:"html_url,omitempty"`
	Id                int32  `json:"id,omitempty"`
	Location          string `json:"location,omitempty"`
	Login             string `json:"login,omitempty"`
	Name              string `json:"name,omitempty"`
	NodeId            string `json:"node_id,omitempty"`
	OrganizationsUrl  string `json:"organizations_url,omitempty"`
	PublicGists       int32  `json:"public_gists,omitempty"`
	PublicRepos       int32  `json:"public_repos,omitempty"`
	ReceivedEventsUrl string `json:"received_events_url,omitempty"`
	ReposUrl          string `json:"repos_url,omitempty"`
	SiteAdmin         bool   `json:"site_admin,omitempty"`
	StarredUrl        string `json:"starred_url,omitempty"`
	SubscriptionsUrl  string `json:"subscriptions_url,omitempty"`
	TwitterUsername   string `json:"twitter_username,omitempty"`
	Type              string `json:"type,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
	Url               string `json:"url,omitempty"`
}
