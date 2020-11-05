/*
 * Copyright (c) 2020  InterDigital Communications, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the \"License\");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an \"AS IS\" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * AdvantEDGE Platform Controller REST API
 *
 * This API is the main Platform Controller API for scenario configuration & sandbox management <p>**Micro-service**<br>[meep-pfm-ctrl](https://github.com/InterDigitalInc/AdvantEDGE/tree/master/go-apps/meep-platform-ctrl) <p>**Type & Usage**<br>Platform main interface used by controller software to configure scenarios and manage sandboxes in the AdvantEDGE platform <p>**Details**<br>API details available at _your-AdvantEDGE-ip-address/api_
 *
 * API version: 1.0.0
 * Contact: AdvantEDGE@InterDigital.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	dataModel "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-data-model"
	log "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-logger"
	sm "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-sessions"
	users "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-users"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	gitlaboauth "golang.org/x/oauth2/gitlab"
)

const OAUTH_PROVIDER_GITHUB = "github"
const OAUTH_PROVIDER_GITLAB = "gitlab"

var mutex sync.Mutex

func initOAuth() {

	// Get default platform URI
	pfmCtrl.uri = strings.TrimSpace(os.Getenv("MEEP_HOST_URL"))

	// Initialize OAuth
	pfmCtrl.oauthConfigs = make(map[string]*oauth2.Config)
	pfmCtrl.loginRequests = make(map[string]*LoginRequest)

	// Initialize Github config
	githubClientId := strings.TrimSpace(os.Getenv("MEEP_OAUTH_GITHUB_CLIENT_ID"))
	githubSecret := strings.TrimSpace(os.Getenv("MEEP_OAUTH_GITHUB_SECRET"))
	githubRedirectUri := strings.TrimSpace(os.Getenv("MEEP_OAUTH_GITHUB_REDIRECT_URI"))
	githubOauthConfig := &oauth2.Config{
		ClientID:     githubClientId,
		ClientSecret: githubSecret,
		RedirectURL:  githubRedirectUri,
		Scopes:       []string{},
		Endpoint:     githuboauth.Endpoint,
	}
	pfmCtrl.oauthConfigs[OAUTH_PROVIDER_GITHUB] = githubOauthConfig

	// Initialize Gitlab config
	gitlabClientId := strings.TrimSpace(os.Getenv("MEEP_OAUTH_GITLAB_CLIENT_ID"))
	gitlabSecret := strings.TrimSpace(os.Getenv("MEEP_OAUTH_GITLAB_SECRET"))
	gitlabRedirectUri := strings.TrimSpace(os.Getenv("MEEP_OAUTH_GITLAB_REDIRECT_URI"))
	gitlabOauthConfig := &oauth2.Config{
		ClientID:     gitlabClientId,
		ClientSecret: gitlabSecret,
		RedirectURL:  gitlabRedirectUri,
		Scopes:       []string{"read_user"},
		Endpoint:     gitlaboauth.Endpoint,
	}
	pfmCtrl.oauthConfigs[OAUTH_PROVIDER_GITLAB] = gitlabOauthConfig
}

// Generate a random state string
func generateState(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func getUniqueState() (state string, err error) {
	for i := 0; i < 3; i++ {
		// Get random state
		randState, err := generateState(20)
		if err != nil {
			log.Error(err.Error())
			return "", err
		}

		// Make sure state is unique
		if _, found := pfmCtrl.loginRequests[randState]; !found {
			return randState, nil
		}
	}
	return "", errors.New("Failed to generate a random state string")
}

func getLoginRequest(state string) *LoginRequest {
	mutex.Lock()
	defer mutex.Unlock()
	request, found := pfmCtrl.loginRequests[state]
	if !found {
		return nil
	}
	return request
}

func setLoginRequest(state string, request *LoginRequest) {
	mutex.Lock()
	defer mutex.Unlock()
	pfmCtrl.loginRequests[state] = request
}

func delLoginRequest(state string) {
	mutex.Lock()
	defer mutex.Unlock()
	request, found := pfmCtrl.loginRequests[state]
	if !found {
		return
	}
	if request.timer != nil {
		request.timer.Stop()
	}
	delete(pfmCtrl.loginRequests, state)
}

func getErrUrl(err string) string {
	return pfmCtrl.uri + "?err=" + strings.ReplaceAll(err, " ", "_")
}

func uaLoginOAuth(w http.ResponseWriter, r *http.Request) {
	log.Info("----- OAUTH LOGIN -----")

	// Retrieve query parameters
	query := r.URL.Query()
	provider := query.Get("provider")

	// Get provider-specific OAuth config
	config, found := pfmCtrl.oauthConfigs[provider]
	if !found {
		err := errors.New("Provider config not found for: " + provider)
		log.Error(err.Error())
		http.Redirect(w, r, getErrUrl(err.Error()), http.StatusFound)
		return
	}

	// Generate unique random state string
	state, err := getUniqueState()
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, getErrUrl(err.Error()), http.StatusFound)
		return
	}

	// Track oauth request & handle
	request := &LoginRequest{
		provider: provider,
		timer:    time.NewTimer(10 * time.Minute),
	}
	setLoginRequest(state, request)

	// Start timer to remove request from map
	go func() {
		<-request.timer.C
		delLoginRequest(state)
	}()

	// Generate provider-specific oauth redirect
	uri := config.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, uri, http.StatusFound)
}

func uaAuthorize(w http.ResponseWriter, r *http.Request) {

	// Retrieve query parameters
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")

	// Validate request state
	request := getLoginRequest(state)
	if request == nil {
		err := errors.New("Invalid OAuth state")
		log.Error(err.Error())
		http.Redirect(w, r, getErrUrl(err.Error()), http.StatusFound)
		return
	}

	// Delete login request & timer
	delLoginRequest(state)

	// Get provider-specific OAuth config
	config := pfmCtrl.oauthConfigs[request.provider]

	// Retrieve access token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, getErrUrl(err.Error()), http.StatusFound)
		return
	}

	// Retrieve User ID
	var userId string
	switch request.provider {
	case OAUTH_PROVIDER_GITHUB:
		oauthClient := config.Client(context.Background(), token)
		client := github.NewClient(oauthClient)
		if client == nil {
			err = errors.New("Failed to create new GitHub client")
			log.Error(err.Error())
			http.Redirect(w, r, getErrUrl(err.Error()), http.StatusFound)
			return
		}
		user, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			log.Error(err.Error())
			http.Redirect(w, r, getErrUrl("Failed to retrieve GitHub user ID"), http.StatusFound)
			return
		}
		userId = *user.Login
	case OAUTH_PROVIDER_GITLAB:
		client, err := gitlab.NewOAuthClient(token.AccessToken)
		if err != nil {
			err = errors.New("Failed to create new GitLab client")
			log.Error(err.Error())
			http.Redirect(w, r, getErrUrl(err.Error()), http.StatusFound)
			return
		}
		user, _, err := client.Users.CurrentUser()
		if err != nil {
			log.Error(err.Error())
			http.Redirect(w, r, getErrUrl("Failed to retrieve GitLab user ID"), http.StatusFound)
			return
		}
		userId = user.Username
	default:
	}

	// Start user session
	sandboxName, err, errCode := startSession(userId, w, r)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, getErrUrl(err.Error()), errCode)
		return
	}

	// Redirect user to sandbox
	http.Redirect(w, r, pfmCtrl.uri+"?sbox="+sandboxName+"&user="+userId, http.StatusFound)
}

func uaLoginUser(w http.ResponseWriter, r *http.Request) {
	log.Info("----- LOGIN -----")

	// Get form data
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate user credentials
	authenticated, err := pfmCtrl.userStore.AuthenticateUser(username, password)
	if err != nil || !authenticated {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Start user session
	sandboxName, err, errCode := startSession(username, w, r)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), errCode)
		return
	}

	// Prepare response
	var sandbox dataModel.Sandbox
	sandbox.Name = sandboxName

	// Format response
	jsonResponse, err := json.Marshal(sandbox)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonResponse))
}

// Retrieve existing user session or create a new one
func startSession(username string, w http.ResponseWriter, r *http.Request) (sandboxName string, err error, code int) {

	// Get existing session by user name, if any
	sessionStore := pfmCtrl.sessionMgr.GetSessionStore()
	session, err := sessionStore.GetByName(username)
	if err != nil {
		// Check if max session count is reached before creating a new one
		count := sessionStore.GetCount()
		if count >= pfmCtrl.maxSessions {
			err = errors.New("Maximum session count exceeded")
			return "", err, http.StatusServiceUnavailable
		}

		// Get requested sandbox name & role from user profile, if any
		role := users.RoleUser
		user, err := pfmCtrl.userStore.GetUser(username)
		if err == nil {
			sandboxName = user.Sboxname
			role = user.Role
		}

		// Get a new unique sanbox name if not configured in user profile
		if sandboxName == "" {
			sandboxName = getUniqueSandboxName()
			if sandboxName == "" {
				err = errors.New("Failed to generate a unique sandbox name")
				return "", err, http.StatusInternalServerError
			}
		}

		// Create sandbox in DB
		var sandboxConfig dataModel.SandboxConfig
		err = createSandbox(sandboxName, &sandboxConfig)
		if err != nil {
			return "", err, http.StatusInternalServerError
		}

		// Create new session
		session = new(sm.Session)
		session.ID = ""
		session.Username = username
		session.Sandbox = sandboxName
		session.Role = role
	} else {
		sandboxName = session.Sandbox
	}

	// Set session
	err, code = sessionStore.Set(session, w, r)
	if err != nil {
		log.Error("Failed to set session with err: ", err.Error())
		// Remove newly created sandbox on failure
		if session.ID == "" {
			deleteSandbox(sandboxName)
		}
		return "", err, code
	}
	return sandboxName, nil, http.StatusOK
}

func uaLogoutUser(w http.ResponseWriter, r *http.Request) {
	log.Info("----- LOGOUT -----")

	// Get existing session
	sessionStore := pfmCtrl.sessionMgr.GetSessionStore()
	session, err := sessionStore.Get(r)
	if err == nil {
		// Delete sandbox
		deleteSandbox(session.Sandbox)
	}

	// Delete session
	err, code := sessionStore.Del(w, r)
	if err != nil {
		log.Error("Failed to delete session with err: ", err.Error())
		http.Error(w, err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func uaTriggerWatchdog(w http.ResponseWriter, r *http.Request) {
	// Refresh session
	sessionStore := pfmCtrl.sessionMgr.GetSessionStore()
	err, code := sessionStore.Refresh(w, r)
	if err != nil {
		log.Error("Failed to refresh session with err: ", err.Error())
		http.Error(w, err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
