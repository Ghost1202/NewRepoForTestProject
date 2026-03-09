package service

import (
	"context"
	"errors"
	"time"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (srv *BookingService) calculateFinalAmount(ctx context.Context, discountType string, booking domain.Booking, eventID int64, tickets []storage.TicketModel) (int64, error) {
	switch discountType {
	case DiscountTypePromo:
		promo, err := srv.tickets.GetPromoByCode(ctx, eventID, booking.PromoCode)
		switch {
		case err == nil:
		case errors.Is(err, gorm.ErrRecordNotFound):
			return 0, ErrInvalidBookingInput
		default:
			return 0, err
		}
		return srv.applyPromoDiscount(promo, tickets)

	case DiscountTypeEarly:
		early, err := srv.tickets.GetEarlyByID(ctx, booking.EarlyID)
		switch {
		case err == nil:
		case errors.Is(err, gorm.ErrRecordNotFound):
			return 0, ErrInvalidBookingInput
		default:
			return 0, err
		}
		return srv.applyEarlyDiscount(early, eventID, tickets)

	case DiscountTypeBundle:
		bundle, err := srv.tickets.GetBundleByID(ctx, booking.BundleID)
		switch {
		case err == nil:
		case errors.Is(err, gorm.ErrRecordNotFound):
			return 0, ErrInvalidBookingInput
		default:
			return 0, err
		}
		return srv.applyBundleDiscount(bundle, eventID, tickets)

	case DiscountTypeNone:
		return sumTicketPrices(tickets), nil

	default:
		return 0, ErrInvalidBookingInput
	}
}

func (srv *BookingService) applyPromoDiscount(promo storage.Promo, tickets []storage.TicketModel) (int64, error) {
	var finalAmount int64
	applied := 0
	for i := range tickets {
		if !srv.isValidPromo(promo, tickets[i]) {
			finalAmount += tickets[i].Price
			continue
		}
		finalAmount += srv.applyDiscountByType(tickets[i].Price, promo.Type, promo.Value)
		applied++
	}
	if applied == 0 {
		return 0, ErrInvalidBookingInput
	}
	return finalAmount, nil
}

func (srv *BookingService) applyEarlyDiscount(early storage.Early, eventID int64, tickets []storage.TicketModel) (int64, error) {
	if early.EventID != eventID || (!early.ValidUntil.IsZero() && time.Now().After(early.ValidUntil)) {
		return 0, ErrInvalidBookingInput
	}

	var finalAmount int64
	applied := 0
	for i := range tickets {
		if early.SectorName != "" && early.SectorName != tickets[i].SectorName {
			finalAmount += tickets[i].Price
			continue
		}
		finalAmount += srv.applyDiscountByType(tickets[i].Price, early.Type, early.Value)
		applied++
	}
	if applied == 0 {
		return 0, ErrInvalidBookingInput
	}
	return finalAmount, nil
}

func (srv *BookingService) applyBundleDiscount(bundle storage.Bundle, eventID int64, tickets []storage.TicketModel) (int64, error) {
	if bundle.EventID != eventID || bundle.BundleBuyCount <= 0 || bundle.BundleGetCount <= 0 {
		return 0, ErrInvalidBookingInput
	}

	total := sumTicketPrices(tickets)

	eligible := make([]int64, 0, len(tickets))
	for i := range tickets {
		if bundle.SectorName != "" && bundle.SectorName != tickets[i].SectorName {
			continue
		}
		eligible = append(eligible, tickets[i].Price)
	}

	bundleSize := bundle.BundleBuyCount + bundle.BundleGetCount
	eligibleCount := int64(len(eligible))
	freeCount := eligibleCount / bundleSize * bundle.BundleGetCount
	if eligibleCount < bundleSize || freeCount <= 0 {
		return 0, ErrInvalidBookingInput
	}

	var discount int64
	for i := 0; i < int(freeCount); i++ {
		discount += eligible[i]
	}

	finalAmount := total - discount
	if finalAmount < 0 {
		return 0, ErrInvalidBookingInput
	}
	return finalAmount, nil
}

func sumTicketPrices(tickets []storage.TicketModel) int64 {
	var total int64
	for i := range tickets {
		total += tickets[i].Price
	}
	return total
}

func (srv *BookingService) applyDiscountByType(price int64, discountType string, value float64) int64 {
	switch discountType {
	case PromoTypePercent:
		discount := float64(price) * (value / 100.0)
		return price - int64(discount)

	case PromoTypeFixed:
		res := price - int64(value)
		if res < 0 {
			return 0
		}
		return res

	default:
		srv.log.Warn("unknown promo type encountered", zap.String("type", discountType))
		return price
	}
}
