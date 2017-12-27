package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"tickets"
	"time"
	"user"
)

type myHandler struct {
	greeting string
}

func (mh myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%v", mh.greeting)))
}

func validateUser(r *http.Request) (response user.UserResponse) {
	r.ParseForm()
	response.Status = true
	if r.FormValue("email") == "" {
		response.Status = false
		response.Message = "Email missing"
	} else if r.FormValue("password") == "" {
		response.Status = false
		response.Message = "Password missing"
	}
	return response
}

func validateBet(r *http.Request) (response tickets.CreateTicketResponse) {
	r.ParseForm()
	response.Status = true
	quantity, errQuantity := strconv.ParseInt(r.FormValue("quantity"), 10, 0)
	price, errPrice := strconv.ParseFloat(r.FormValue("price"), 10)
	if r.FormValue("ticket") == "" {
		response.Status = false
		response.Message = "ticket missing"
	} else if r.FormValue("betType") == "" {
		response.Status = false
		response.Message = "betType missing"
	} else if r.FormValue("side") == "" {
		response.Status = false
		response.Message = "side missing"
	} else if r.FormValue("price") == "" {
		response.Status = false
		response.Message = "price missing"
	} else if r.FormValue("userId") == "" {
		response.Status = false
		response.Message = "userId missing"
	} else if r.FormValue("quantity") == "" {
		response.Status = false
		response.Message = "quantity missing"
	} else if errQuantity != nil {
		response.Status = false
		response.Message = "quantity not a number"
	} else if errPrice != nil {
		response.Status = false
		response.Message = "price not a number"
	} else if quantity > 10 || quantity < 1 {
		response.Status = false
		response.Message = "quantity must be between 1 and 10"
	} else if price > 10 || price < 0 {
		response.Status = false
		response.Message = "price must be between 1 and 10"
	}
	return response
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pwd, _ := os.Getwd()
		var filePath string
		filePath = pwd + "/../public" + r.URL.Path

		if r.URL.Path == "/" {
			filePath = pwd + "/../public/index.html"
		}
		http.ServeFile(w, r, filePath)
	})

	http.HandleFunc("/validateuser", func(w http.ResponseWriter, r *http.Request) {
		response := validateUser(r)
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.HandleFunc("/loginuser", func(w http.ResponseWriter, r *http.Request) {
		response := validateUser(r)
		if response.Status == true {
			response = user.LoginUser(r.FormValue("email"), r.FormValue("password"))
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.HandleFunc("/getuser", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		response := user.GetUser(r.FormValue("email"))
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.HandleFunc("/getallusers", func(w http.ResponseWriter, r *http.Request) {
		response := user.GetAllUsers()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.HandleFunc("/getalltickets", func(w http.ResponseWriter, r *http.Request) {
		response := tickets.GetAllTickets()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.HandleFunc("/createuser", func(w http.ResponseWriter, r *http.Request) {
		response := validateUser(r)
		if response.Status == true {
			response = user.PutUser(r.FormValue("email"), r.FormValue("password"))
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.HandleFunc("/createticket", func(w http.ResponseWriter, r *http.Request) {
		response := validateBet(r)
		if response.Status == true {
			padLength := 4
			priceParts := strings.Split(r.FormValue("price"), ".")
			price := strings.Repeat("0", padLength-len(priceParts[0])) + priceParts[0] + "." + priceParts[1] + strings.Repeat("0", padLength-len(priceParts[1]))

			quantity, _ := strconv.ParseInt(r.FormValue("quantity"), 10, 0)
			for j := int(1); j <= int(quantity); j++ {
				priceAndTimeAndQuantity := fmt.Sprintf("%v_%v_%v_%vOF%v", r.FormValue("side"), price, time.Now().UTC().String(), j, quantity)
				response = tickets.CreateTicket(priceAndTimeAndQuantity, r.FormValue("ticket"), r.FormValue("betType"), r.FormValue("side"), r.FormValue("price"), r.FormValue("userId"))
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.ListenAndServe(":4444", nil)
}
