package utils

import "strings"

// GetTicketWithoutSide strips the home and away
func GetTicketWithoutSide(ticket string) (retstring string) {
	retstring = strings.Replace(ticket, "_home", "", 1)
	retstring = strings.Replace(retstring, "_away", "", 1)
	return retstring
}
