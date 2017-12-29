package main

import (
	"encoding/json"
	"net/http"
	"os"
	"stats"
	"strconv"
	"tickets"
	"user"
)

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
	price, errPrice := strconv.ParseInt(r.FormValue("price"), 10, 0)
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
	} else if price > 100 || price < 0 {
		response.Status = false
		response.Message = "price must be between 1 and 100"
	}
	return response
}

func validateDeleteTicket(r *http.Request) (response tickets.CreateTicketResponse) {
	r.ParseForm()
	response.Status = true
	if r.FormValue("ticket") == "" {
		response.Status = false
		response.Message = "ticket missing"
	} else if r.FormValue("priceAndTime") == "" {
		response.Status = false
		response.Message = "priceAndTime missing"
	}
	return response
}

// APIErrorResponse is for sending error messages back to the client
type APIErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func respondWithError(w http.ResponseWriter, errorMessage string) {
	response := APIErrorResponse{Error: true, Message: errorMessage}
	enc := json.NewEncoder(w)
	enc.Encode(response)
}

func respondWithMessage(w http.ResponseWriter, hasError bool, message string) {
	response := APIErrorResponse{Error: hasError, Message: message}
	enc := json.NewEncoder(w)
	enc.Encode(response)
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

	// http.HandleFunc("/validateuser", func(w http.ResponseWriter, r *http.Request) {
	// 	response := validateUser(r)
	// 	enc := json.NewEncoder(w)
	// 	enc.Encode(response)
	// })

	http.HandleFunc("/loginuser", func(w http.ResponseWriter, r *http.Request) {
		response := validateUser(r)
		if response.Status == true {
			response = user.LoginUser(r.FormValue("email"), r.FormValue("password"))
		}
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

	http.HandleFunc("/getuser", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		response := user.GetUser(r.FormValue("email"))
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})

	http.HandleFunc("/getallusers", func(w http.ResponseWriter, r *http.Request) {
		response := user.GetAllUsers()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if response == nil {
			respondWithError(w, "No users")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/createticket", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		response := validateBet(r)
		if response.Status != true {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		} else {
			_, err := tickets.CreateTicketsFromFormData(r.FormValue("ticket"), r.FormValue("betType"), r.FormValue("side"), r.FormValue("price"), r.FormValue("userId"), r.FormValue("quantity"))

			if err != nil {
				respondWithError(w, err.Error())
			} else {
				respondWithMessage(w, false, "Ticket created")
			}
		}
	})

	http.HandleFunc("/getticket", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		response, err := tickets.GetTicket(r.FormValue("ticket"), r.FormValue("priceAndTime"))

		if err != nil {
			respondWithError(w, err.Error())
		} else if response.Ticket == "" {
			respondWithError(w, "Ticket not found")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/getalltickets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		response, err := tickets.GetAllTickets()
		if err != nil {
			respondWithError(w, err.Error())
		} else if len(response) == 0 {
			respondWithError(w, "No tickets found.")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/getticketcommon", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		response, err := tickets.GetTicketCommon(r.FormValue("ticket"))

		if err != nil {
			respondWithError(w, err.Error())
		} else if len(response) == 0 {
			respondWithError(w, "No tickets found.")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/getticketoppose", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		ticket := r.FormValue("ticket")

		response, err := tickets.GetTicketOppose(ticket)

		if err != nil {
			respondWithError(w, err.Error())
		} else if len(response) == 0 {
			respondWithError(w, "No tickets found.")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	type BetResponse struct {
		Ticket  tickets.TicketObj `json:"ticket"`
		Oppose  tickets.TicketObj `json:"opposet"`
		Status  bool              `json:"status"`
		Message string            `json:"message"`
	}

	http.HandleFunc("/deleteticket", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		response := validateDeleteTicket(r)

		if response.Status != true {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		} else {
			err := tickets.DeleteTicket(r.FormValue("ticket"), r.FormValue("priceAndTime"))
			if err != nil {
				respondWithError(w, err.Error())
			} else {
				respondWithError(w, "DELETE WORKED")
			}
		}
	})

	http.HandleFunc("/updateuser", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		_, err := user.UpdateUser(r.FormValue("email"), r.FormValue("bankroll"))
		if err != nil {
			respondWithError(w, err.Error())
		} else {
			respondWithMessage(w, false, "Update Worked")
		}
	})

	http.HandleFunc("/createbet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		response, err := tickets.CreateBetFromTicketID(r.FormValue("ticket"), r.FormValue("priceAndTime"))
		if err != nil {
			respondWithError(w, err.Error())
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/getallbets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		response, err := tickets.GetAllBets()
		if err != nil {
			respondWithError(w, err.Error())
		} else if len(response) == 0 {
			respondWithError(w, "No bets found.")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/updatestat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		response, err := stats.UpdateStat(r.FormValue("ticket"), r.FormValue("type"), r.FormValue("value"))
		if err != nil {
			respondWithError(w, err.Error())
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/updatecounter", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()

		response, err := stats.UpdateCounter(r.FormValue("ticket"), r.FormValue("type"), r.FormValue("amount"))
		if err != nil {
			respondWithError(w, err.Error())
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.HandleFunc("/getallstats", func(w http.ResponseWriter, r *http.Request) {
		response := stats.GetAllStats()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if response == nil {
			respondWithError(w, "No stats")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.ListenAndServe(":4444", nil)
}
