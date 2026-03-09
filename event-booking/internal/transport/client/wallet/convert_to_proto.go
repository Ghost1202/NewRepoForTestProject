package wallet

import (
	"github.com/turtlepavlo/event-booking/internal/domain"
	walletv1 "github.com/turtlepavlo/proto-contract/gen/go/wallet/v1"
)

type protoConverter struct{}

func newProtoConverter() protoConverter { return protoConverter{} }

func (c protoConverter) ToChargeRequest(input domain.WalletCharge) *walletv1.ChargeWalletRequest {
	return &walletv1.ChargeWalletRequest{
		BookingId: input.BookingID,
		UserId:    input.UserID,
		EventId:   input.EventID,
		TicketIds: input.TicketIDs,
		Amount:    input.Amount,
		Currency:  input.Currency,
	}
}

func (c protoConverter) ToChargeStatus(status walletv1.ChargeStatus) domain.WalletChargeStatus {
	switch status {
	case walletv1.ChargeStatus_CHARGE_STATUS_SUCCESS:
		return domain.WalletChargeStatusSuccess
	case walletv1.ChargeStatus_CHARGE_STATUS_TICKET_TAKEN:
		return domain.WalletChargeStatusTicketTaken
	case walletv1.ChargeStatus_CHARGE_STATUS_INSUFFICIENT_FUNDS:
		return domain.WalletChargeStatusInsufficientFunds
	default:
		return domain.WalletChargeStatusUnspecified
	}
}
