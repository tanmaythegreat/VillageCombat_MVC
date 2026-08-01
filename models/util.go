package models

import "log"

func logErr(context string, err error) {
	if err != nil {
		log.Printf("%s: %v", context, err)
	}
}
