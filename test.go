package main

import (
	"fmt"

	"github.com/damonto/telegram-sms/esim"
)

func main() {
	esim, err := esim.New("/dev/ttyUSB6")
	fmt.Println(esim, err)
}
