package notifier

import (
	"context"

	"github.com/jacobmichels/Course-Sense-Go/internal/types"
	"github.com/twilio/twilio-go"
)

type TwilioNotifier struct {
	accountSid  string
	authToken   string
	phoneNumber string
	client      *twilio.RestClient
}

func NewTwilioNotifier(accountSid, authToken, phoneNumber string) *TwilioNotifier {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	return &TwilioNotifier{accountSid, authToken, phoneNumber, client}
}

func (tn *TwilioNotifier) Name() string {
	return "TwilioNotifier"
}

func (tn *TwilioNotifier) Notify(ctx context.Context, section *types.CourseSection) error {
	return nil
}
