package main

import (
	"fmt"
	"os"

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
	a := App{}
	a.Run(mayEnv("STEAM_WEB_API_BIND_ADDRESS", ":8000"))
}
