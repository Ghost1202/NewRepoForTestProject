package service

import (
	"github.com/google/uuid"
	"github.com/turtlepavlo/stripe_integration/internal/domain"
)

type ToStorageConvert struct{}

func NewToStorageConvert() *ToStorageConvert { return &ToStorageConvert{} }

func (s *ToStorageConvert) ToReservationParams(payment domain.CreatePaymentInput) domain.Payment {
	return domain.Payment{
		OrderID:    payment.OrderID,
		UserID:     payment.UserID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		ExternalID: "",
		Status:     StatusProcessing,
	}
}

func (s *ToStorageConvert) ToTopupReservationParams(payment domain.TopupWalletInput) domain.Payment {
	return domain.Payment{
		OrderID:    payment.OrderID,
		UserID:     payment.UserID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		ExternalID: "",
		Status:     StatusProcessing,
	}
}

func (s *ToStorageConvert) ToUpdateStatusParams(orderID uuid.UUID, status, sessionID string) domain.Payment {
	return domain.Payment{
		OrderID:    orderID,
		Status:     status,
		ExternalID: sessionID,
	}
}

func (s *ToStorageConvert) ToCompleteParams(orderID uuid.UUID, sessionID string) domain.Payment {
	return s.ToUpdateStatusParams(orderID, StatusPending, sessionID)
}

func (s *ToStorageConvert) ToFailedParams(orderID uuid.UUID) domain.Payment {
	return s.ToUpdateStatusParams(orderID, StatusFailed, "")
}

func (s *ToStorageConvert) ToPaidParams(orderID uuid.UUID, sessionID string) domain.Payment {
	return s.ToUpdateStatusParams(orderID, StatusPaid, sessionID)
}

func (s *ToStorageConvert) ToSetRefundedStatus(externalID string) (updatedExternalID, status string) {
	return externalID, StatusRefunded
}
