package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/npflan/steam-gameserver-token-api/steam"
)

// App contains references to global necessities
type App struct {
	Router *mux.Router
	Log    Log
}

// Log is a modifiable endpoint
type Log struct {
	Error *log.Logger
	Info  *log.Logger
}

const defaultLogFormat = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.LUTC

// Run server on specific interface
func (a *App) Run(addr string, enableWiping bool) {
	a.registerRoutes(enableWiping)
	// set default log format if no custom format present
	if a.Log.Info == nil {
		log.New(os.Stdout, "INFO: ", defaultLogFormat)
	}
	if a.Log.Error == nil {
		log.New(os.Stderr, "ERROR: ", defaultLogFormat)
	}
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) registerRoutes(enableWiping bool) {
	a.Router = mux.NewRouter().StrictSlash(true)
	a.Router.HandleFunc("/", a.getHome).Methods("GET")
	a.Router.HandleFunc("/token/{appID}/{memo}", a.pullToken).Methods("GET")

	if enableWiping {
		a.Router.HandleFunc("/wipe", a.wipeTokens).Methods("GET")
	}
}

// RespondWithJSON uses a struct, for a JSON response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithText returns text/plain.
func respondWithText(w http.ResponseWriter, code int, payload string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(payload))
}

// RespondWithError standardizes error messages, through the use of RespondWithJSON.
func respondWithError(w http.ResponseWriter, code int, message string) {
	fmt.Println(message)
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (a *App) getHome(w http.ResponseWriter, r *http.Request) {

}

func (a *App) pullToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if _, ok := vars["appID"]; !ok {
		respondWithError(w, http.StatusBadRequest, "Missing appID")
		return
	}
	if _, ok := vars["memo"]; !ok {
		respondWithError(w, http.StatusBadRequest, "Missing memo")
		return
	}

	appID, err := strconv.Atoi(vars["appID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("bad appID: %s", err))
		return
	}

	accounts, err := steam.GetAccountList()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to list existing tokens: %s", err))
		return
	}

	// Check for existing account
	var account steam.Account
	for _, acct := range accounts {
		if acct.Memo == vars["memo"] && int(acct.AppID) == appID {
			account = acct
			break
		}
	}

	// Create new if not found
	if account.SteamID == "" {
		account, err = steam.CreateAccount(appID, vars["memo"])
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh token if found and expired
	if account.IsExpired == true {
		account, err = steam.ResetLoginToken(account.SteamID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithText(w, http.StatusOK, account.LoginToken)
}

func (a *App) wipeTokens(w http.ResponseWriter, r *http.Request) {
	accounts, err := steam.GetAccountList()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to list existing tokens: %s", err))
		return
	}

	for _, acct := range accounts {
		err = steam.DeleteAccount(acct.SteamID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithText(w, http.StatusOK, "OK")
}
