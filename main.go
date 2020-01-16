package main

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/npflan/steam-gameserver-token-api/steam"
)

func mayEnv(envVar string, deflt string) string {
	env := os.Getenv(envVar)
	if env == "" {
		fmt.Printf("Environment variable '%s' is empty / undefined. Defaulting to '%s'", envVar, deflt)
		return deflt
	}
	return env
}

func main() {
	enableWiping, err := strconv.ParseBool(mayEnv("ENABLE_WIPING", "false"))
	if err != nil {
		fmt.Println("Invalid value specified for ENABLE_WIPING")
		os.Exit(1)
	}

	a := App{}
	a.Run(mayEnv("STEAM_WEB_API_BIND_ADDRESS", ":8000"), enableWiping)
}
