package main

import (
	"github.com/xuebing1110/broadlink"
	"time"
)

func main() {
	_, err := broadlink.Discover(5 * time.Second)
	if err != nil {
		panic(err)
	}
}
