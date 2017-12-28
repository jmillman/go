package apblogger

import (
	"fmt"
	"net/http"
	"strings"
)

// LogRequest print out the Request Info
func LogRequest(functionName string, r *http.Request) {
	fmt.Printf("____%v____\n", functionName)
	fmt.Printf("%v\n", r)
}

//LogMessage Print out a message
func LogMessage(message string) {
	fmt.Printf("%v\n", message)
}

//LogVar Print out a variable
func LogVar(varName string, variable string) {
	fmt.Printf("%v: %v\n", varName, variable)
}

//LogForm Print out the Form Info
func LogForm(functionName string, r *http.Request) {
	r.ParseForm()
	fmt.Printf("____%v____\n", functionName)
	fmt.Println("path", r.URL.Path)
	fmt.Println("__Form Values")
	for k, v := range r.Form {
		fmt.Printf("[%v]:\t%v\n", k, strings.Join(v, ","))
	}
}
