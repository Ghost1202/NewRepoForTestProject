package handler

import (
	"net/http"

	"github.com/turtlepavlo/event-booking/internal/domain"
)

func walletEnum(status domain.WalletChargeStatus) int {
	switch status {
	case domain.WalletChargeStatusSuccess:
		return http.StatusCreated
	case domain.WalletChargeStatusTicketTaken:
		return http.StatusConflict
	case domain.WalletChargeStatusInsufficientFunds:
		return http.StatusPaymentRequired
	default:
		return http.StatusInternalServerError
	}
}
