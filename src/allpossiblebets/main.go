package main

import (
	"apblogger"
	"bets"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"tickets"
	"time"
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
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func respondWithError(w http.ResponseWriter, errorMessage string) {
	response := APIErrorResponse{Error: true, Message: errorMessage}
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

		fmt.Printf("getuser_____________\n")
		fmt.Printf("%v\n", response)

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
			quantity, _ := strconv.ParseInt(r.FormValue("quantity"), 10, 0)
			// TODO: don't hardcode the max Value
			price, _ := strconv.ParseInt(r.FormValue("price"), 10, 0)
			sortKeyPrice := 100 - price //need to sort by largest number and dynamodb sorts by smallest so have to take invese, should be max - price, hard coded in 100
			for j := int(1); j <= int(quantity); j++ {
				sortKey := fmt.Sprintf("%v_%v_%vOF%v", sortKeyPrice, time.Now().UnixNano(), j, quantity)
				_, err := tickets.CreateTicket(sortKey, r.FormValue("ticket"), r.FormValue("betType"), r.FormValue("side"), r.FormValue("price"), r.FormValue("userId"))
				if err != nil {
					respondWithError(w, err.Error())
					return
				}
			}
			enc := json.NewEncoder(w)
			enc.Encode(APIErrorResponse{Error: false, Message: "Tickets Created"})
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
		apblogger.LogForm("getticketcommon", r)

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
		response, err := user.UpdateUser(r.FormValue("email"), r.FormValue("bankroll"))
		apblogger.LogMessage("updateuser")
		apblogger.LogVar("response", fmt.Sprintf("%v", response))
		if err != nil {
			respondWithError(w, err.Error())
		} else {
			respondWithError(w, "Update Worked")
		}
	})

	http.HandleFunc("/getbet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		ticket, err := tickets.GetTicket(r.FormValue("ticket"), r.FormValue("priceAndTime"))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err != nil {
			respondWithError(w, err.Error())
		} else {
			if ticket.Ticket != "" {
				// ticket found
				ticketsJSON, _ := json.Marshal(ticket)
				ticketsOppose, err := tickets.GetTicketOppose(ticket.Ticket)
				ticketOpposeTop := ticketsOppose[0]
				if err != nil {
					respondWithError(w, err.Error())
				} else {
					ticketPrice, err := strconv.ParseInt(ticket.Price, 10, 0)
					ticketPriceOppose, err := strconv.ParseInt(ticketOpposeTop.Price, 10, 0)
					maxPrice, err := strconv.ParseInt("100", 10, 0)
					if err != nil {
						respondWithError(w, err.Error())
					}

					if len(ticketsOppose) > 0 && (ticketPrice >= (100 - ticketPriceOppose)) {
						if maxPrice-ticketPrice-ticketPriceOppose > 0 {
							respondWithError(w, "The prices didn't equal")
						} else {
							overPaid, err := strconv.ParseInt("0", 10, 0)
							if maxPrice != ticketPrice+ticketPriceOppose {
								overPaid = maxPrice - ticketPrice - ticketPriceOppose
							}
							apblogger.LogVar("overPaid", fmt.Sprintf("%v", overPaid))
							ticketPriceAdjusted := int64(ticketPrice + (overPaid / 2))
							ticketPriceOpposeAdjusted := maxPrice - ticketPriceAdjusted
							ticketToWrite := ticket
							ticketToWriteOpposeTop := ticketOpposeTop

							ticketToWrite.PriceAdjusted = fmt.Sprintf("%v", ticketPriceAdjusted)
							ticketToWriteOpposeTop.PriceAdjusted = fmt.Sprintf("%v", ticketPriceOpposeAdjusted)

							userTicket := user.GetUser(ticket.UserID)
							userTicketBankroll := int64(userTicket.Bankroll)
							userTicketOppose := user.GetUser(ticketOpposeTop.UserID)
							userTicketOpposeBankroll := int64(userTicketOppose.Bankroll)
							if userTicketBankroll < ticketPriceAdjusted {
								respondWithError(w, "User doesn't have enough bankroll")
							}
							if userTicketOpposeBankroll < ticketPriceOpposeAdjusted {
								respondWithError(w, "User 2 doesn't have enough bankroll")
							}

							apblogger.LogVar("ticketToWrite.Price >> NEW >>", fmt.Sprintf("%v", ticketToWrite.Price))
							apblogger.LogVar("ticketToWriteOpposeTop.Price >> NEW >>", fmt.Sprintf("%v", ticketToWriteOpposeTop.Price))

							apblogger.LogVar("ticketToWrite", fmt.Sprintf("%v", ticketToWrite))
							apblogger.LogVar("ticketToWriteOpposeTop", fmt.Sprintf("%v", ticketToWriteOpposeTop))

							history := BetResponse{Ticket: ticketToWrite, Oppose: ticketToWriteOpposeTop}
							historyJSON, _ := json.Marshal(history)
							timestamp := fmt.Sprintf("%v", time.Now().UnixNano())

							response, err := bets.CreateBet(ticket.Ticket, timestamp, ticket.BetType, ticket.UserID, ticketOpposeTop.UserID, string(historyJSON))
							if err != nil {
								respondWithError(w, err.Error())
							} else {
								user.UpdateUser(ticket.UserID, fmt.Sprintf("%v", userTicketBankroll-ticketPriceAdjusted))
								user.UpdateUser(ticketOpposeTop.UserID, fmt.Sprintf("%v", userTicketOpposeBankroll-ticketPriceOpposeAdjusted))
								err = tickets.DeleteTicket(ticket.Ticket, ticket.PriceAndTime)
								err = tickets.DeleteTicket(ticketOpposeTop.Ticket, ticketOpposeTop.PriceAndTime)
							}
							enc := json.NewEncoder(w)
							enc.Encode(response)
						}
					} else {
						apblogger.LogVar("Match Not Found", string(ticketsJSON))
						respondWithError(w, "Match not found")
					}
				}
			}

		}
	})

	http.HandleFunc("/getallbets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		response, err := bets.GetAllBets()
		if err != nil {
			respondWithError(w, err.Error())
		} else if len(response) == 0 {
			respondWithError(w, "No bets found.")
		} else {
			enc := json.NewEncoder(w)
			enc.Encode(response)
		}
	})

	http.ListenAndServe(":4444", nil)
}
