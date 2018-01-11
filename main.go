package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	"github.com/urfave/cli"
)

// baseURL/interface/method/version?parameters
const location = "https://api.steampowered.com/IGameServersService/"
const version = "v1"

var apiKey string
var requireAPIKey = func(c *cli.Context) error {
	if len(apiKey) == 0 {
		return cli.NewExitError("API key not provided", 1)
	}
	return nil
}

// Steam returns a JSON { response: } object, which wraps all return values.
type steamResponse struct {
	Response json.RawMessage `json:"response"`
}

type response struct {
	Servers []accountEntity `json:"servers,omitempty"`
}

type accountEntity struct {
	SteamID    string `json:"steamid,omitempty"`
	AppID      uint16 `json:"appid,omitempty"`
	LoginToken string `json:"login_token,omitempty"`
	Memo       string `json:"memo,omitempty"`
	IsDeleted  bool   `json:"is_deleted,omitempty"`
	IsExpired  bool   `json:"is_expired,omitempty"`
	LastLogon  int    `json:"rt_last_logon,omitempty"`
}

// Wrap requests for Steam Web API, to generalize insertion of API key,
// and handling of Response Header.
func querySteam(command string, method string, params map[string]string) (data []byte, err error) {
	// ready request for location
	req, err := http.NewRequest(method, location+command+"/"+version, nil)
	if err != nil {
		return nil, err
	}
	// Add API Key and extra parameters
	q := url.Values{}
	q.Add("key", apiKey)
	for key, value := range params {
		q.Add(key, value)
	}
	// Encode parameters and append them to the url
	req.URL.RawQuery = q.Encode()
	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Drop if Error Header present
	if respErrState := resp.Header.Get("X-error_message"); respErrState != "" {
		log.Fatal(respErrState)
	}
	// Decode response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Remove the { response: data } wrapper, and return inner json as byte array.
func unwrapResponse(response *[]byte) error {
	resp := steamResponse{}
	if err := json.Unmarshal(*response, &resp); err != nil {
		log.Fatal(err)
	}
	*response = ([]byte)(resp.Response)
	return nil
}

// NotYetImplemented
func printCSV(arr interface{}) error {
	arrValue := reflect.ValueOf(arr)
	values := make([]interface{}, arrValue.NumField())
	for key, value := range values {
		fmt.Printf("%v;%v", key, value)
	}

	/*v := reflect.ValueOf(elements[0])
	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%s", v.Field(i).Interface())
	}*/
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "Steam Gameserver Token Manager"
	app.Usage = "making server management a little bit easier"
	app.UsageText = "main [global options] command [command options]"
	app.Version = "0.1.0"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Kristian Dahl Kærgaard",
			Email: "hcand.dk@gmail.com",
		},
	}

	// Customize version printer to include compile timestamp
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s %s\nCompiled at %s\n", app.Name, app.Version, app.Compiled)
	}

	// Define Global Flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "key, k",
			Value:       "",
			Usage:       "API key (optionally from environment variable)",
			EnvVar:      "STEAM_WEB_API_KEY",
			Destination: &apiKey,
		},
		/*cli.StringFlag{
			Name:  "format, f",
			Value: "csv",
			Usage: "Output format for returned data (csv, json)",
		},*/
	}

	// Define Subcommands, each with its own set of flags
	app.Commands = []cli.Command{
		{
			Name:    "GetAccountList",
			Aliases: []string{"gal"},
			Usage:   "Gets a list of game server accounts with their logon tokens",
			Before:  requireAPIKey,
			Action: func(c *cli.Context) error {
				params := make(map[string]string)
				data, err := querySteam("GetAccountList", "GET", params)
				if err != nil {
					log.Fatal(err)
				}
				unwrapResponse(&data)
				var response response
				if err := json.Unmarshal(data, &response); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%s;%s;%s;%s;%s;%s;%s\n", "AppID", "IsDeleted", "IsExpired", "LastLogon", "LoginToken", "Memo", "SteamID")
				for _, server := range response.Servers {
					fmt.Printf("%d;%t;%t;%d;%s;%s;%s\n", server.AppID, server.IsDeleted, server.IsExpired, server.LastLogon, server.LoginToken, server.Memo, server.SteamID)
				}
				return nil
			},
		},
		{
			Name:    "CreateAccount",
			Aliases: []string{"ca"},
			Usage:   "Creates a persistent game server account",
			Before:  requireAPIKey,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, appid",
					Value: "730",
					Usage: "The app to use the account for",
				},
				cli.StringFlag{
					Name:  "m, memo",
					Value: "",
					Usage: "The memo to set on the new account",
				},
			},
			Action: func(c *cli.Context) error {
				params := make(map[string]string)
				params["appid"] = c.String("appid")
				params["memo"] = c.String("memo")
				data, err := querySteam("CreateAccount", "POST", params)
				if err != nil {
					log.Fatal(err)
				}
				unwrapResponse(&data)
				var account accountEntity
				if err := json.Unmarshal(data, &account); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%s;%s\n", "SteamID", "LoginToken")
				fmt.Printf("%s;%s\n", account.SteamID, account.LoginToken)
				return nil
			},
		},
		/*{
			Name:    "SetMemo",
			Aliases: []string{"sm"},
			Usage:   "Change the memo associated with the game server account. Memos do not affect the account in any way. The memo shows up in the GetAccountList response and serves only as a reminder of what the account is used for.",
			Before:  requireAPIKey,
			Flags: []cli.Flag{
				cli.Uint64Flag{
					Name:  "s, steamid",
					Value: 0,
					Usage: "SteamID of the game server to set the memo on",
				},
				cli.StringFlag{
					Name:  "m, memo",
					Value: "",
					Usage: "Memo to set on the account",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Uint64("steamid") == 0 {
					return cli.NewExitError("steamid not provided", 1)
				}
				return nil
			},
		},
		{
			Name:    "ResetLoginToken",
			Aliases: []string{"rlt"},
			Usage:   "Generate a new login token for the specified game server",
			Before:  requireAPIKey,
			Flags: []cli.Flag{
				cli.Uint64Flag{
					Name:  "s, steamid",
					Value: 0,
					Usage: "SteamID of the game server to reset the login token of",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Uint64("steamid") == 0 {
					return cli.NewExitError("steamid not provided", 1)
				}
				return nil
			},
		},*/
		{
			Name:    "DeleteAccount",
			Aliases: []string{"da"},
			Usage:   "Delete a persistent game server account",
			Before:  requireAPIKey,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "s, steamid",
					Value: "0",
					Usage: "SteamID of the game server account to delete",
				},
			},
			Action: func(c *cli.Context) error {
				params := make(map[string]string)
				params["steamid"] = c.String("steamid")
				_, err := querySteam("DeleteAccount", "POST", params)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Deleted steamid %s\n", c.String("steamid"))
				return nil
			},
		},
		/*{
			Name:    "GetAccountPublicInfo",
			Aliases: []string{"gapi"},
			Usage:   "Get public information about a given game server account",
			Flags: []cli.Flag{
				cli.Uint64Flag{
					Name:  "s, steamid",
					Value: 0,
					Usage: "SteamID of the game server to get info on",
				},
			},
		},
		{
			Name:    "QueryLoginToken",
			Aliases: []string{"qlt"},
			Usage:   "Query the status of the specified token, which must be owned by you",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "t, token",
					Value: "",
					Usage: "Login token to query",
				},
			},
		},
		{
			Name:    "GetServerSteamIDsByIP",
			Aliases: []string{"gssibi"},
			Usage:   "Get a list of server SteamIDs given a list of IPs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "s, servers",
					Value: "",
					Usage: "List of server IPs to query",
				},
			},
		},
		{
			Name:    "GetServerIPsBySteamID",
			Aliases: []string{"gsibsi"},
			Usage:   "Get a list of server IP addresses given a list of SteamIDs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "s, steamids",
					Value: "",
					Usage: "List of Steam IDs to query",
				},
			},
		},*/
	}

	app.Run(os.Args)
	//parseFlags()
	/*
		q := url.Values{}
		q.Add("key", apiKey)
		q.Add("appid", "730")

		tokenURL, err := url.Parse(location)
		if err != nil {
			log.Fatal(err)
		}

		client := &http.Client{}

		var data map[string]interface{}

		for server := 1; server <= serverCount; server++ {
			for instance := 1; instance <= instanceCount; instance++ {
				// build
				q.Set("memo", fmt.Sprintf("%d:%d", server, instance))
				req, err := http.NewRequest("POST", tokenURL.String(), nil)
				if err != nil {
					log.Fatal(err)
				}
				req.URL.RawQuery = q.Encode()
				// execute
				resp, err := client.Do(req)
				if err != nil {
					log.Fatal(err)
				}
				// decode
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				err = json.Unmarshal(body, &data)
				// print
				fmt.Printf("%d:%d %s\n", server, instance, data["response"].(map[string]interface{})["login_token"])
				//fmt.Printf("%v\n", resp)
			}
		}
	*/
}

/*
	GetAccountList
		key
	CreateAccount
		key
		appid
		memo
	SetMemo
		key
		steamid
		memo
	ResetLoginToken
		key
		steamid
	DeleteAccount
		key
		steamid
	GetAccountPublicInfo
		key
		steamid
	QueryLoginToken
		key
		login_token
*/
