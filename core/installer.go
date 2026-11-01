package main

import (
	"fmt"
	"strings"

	"ninixw-core/structs"
)

type installer struct {
}

func (i installer) SendEvent(event ...structs.Event) string {
	for _, e := range event {
		switch e.Event {
		case "OnInstallFailure":
			fmt.Println(fmt.Sprintf("failed to install: %s", e.Reference[0]))
		case "OnInstallCompleteEx":
			for _, v := range strings.Split(e.Reference[1], "\x01") {
				fmt.Println(fmt.Sprintf("install successful: %s", v))
			}
		}
		break
	}
	return ""
}
