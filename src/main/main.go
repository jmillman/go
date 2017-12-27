package main

import "fmt"

func main(){
	message := "go message"
	var greeting *string = & message
	fmt.Println(message, *greeting)
}
