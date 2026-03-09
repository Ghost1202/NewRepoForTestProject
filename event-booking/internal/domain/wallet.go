package domain

type WalletChargeStatus string

const (
	WalletChargeStatusUnspecified       WalletChargeStatus = "CHARGE_STATUS_UNSPECIFIED"
	WalletChargeStatusSuccess           WalletChargeStatus = "CHARGE_STATUS_SUCCESS"
	WalletChargeStatusTicketTaken       WalletChargeStatus = "CHARGE_STATUS_TICKET_TAKEN"
	WalletChargeStatusInsufficientFunds WalletChargeStatus = "CHARGE_STATUS_INSUFFICIENT_FUNDS"
)

type WalletCharge struct {
	BookingID string
	UserID    int64
	EventID   int64
	TicketIDs []int64
	Amount    int64
	Currency  string
}

type WalletChargeResult struct {
	BookingID string
	Status    WalletChargeStatus
}
