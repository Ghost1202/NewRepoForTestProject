package service

import (
	eventv1 "github.com/turtlepavlo/proto-contract/gen/go/analytics/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConvToProto struct{}

func NewConvToProto() *ConvToProto { return &ConvToProto{} }

func (c *ConvToProto) LoginEvent(userID int64, typ string) *eventv1.LoginEvent {
	return &eventv1.LoginEvent{
		UserId:    userID,
		Type:      typ,
		CreatedAt: timestamppb.Now(),
	}
}

func (c *ConvToProto) RegisterEvent(userID int64, email string) *eventv1.RegisterEvent {
	return &eventv1.RegisterEvent{
		UserId:    userID,
		Email:     email,
		CreatedAt: timestamppb.Now(),
	}
}
