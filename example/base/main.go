package main

import (
	"fmt"
	"log"

	"mlib.com/confy"
)

func main() {
	confy.SetConfigFile("config-win.yaml")
	err := confy.ReadInConfig()
	if err != nil {
		log.Println("read config file failed:", err)
		return
	}
	fmt.Println(confy.Get[map[string]any]("mysql"))
}
