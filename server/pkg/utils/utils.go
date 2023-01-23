package utils

import (
	"fmt"
	"runtime"
	"time"
)


func PrintGos() {
	for {
		fmt.Printf("no of gos: %d\n", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}