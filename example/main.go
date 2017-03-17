package main

import (
	"github.com/xuebing1110/broadlink"
	"time"
)

func main() {
	devs, err := broadlink.Discover(5 * time.Second)
	if err != nil {
		panic(err)
	}

	for _, dev := range devs {
		rmdev := dev.(*broadlink.RmDevice)
		err = rmdev.BaseDevice.Auth()
		if err != nil {
			panic(err)
		}
	}
}
