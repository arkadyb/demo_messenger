package messenger

import (
	"github.com/arkadyb/demo_messenger/internal/pkg/buffer"
	messagebird "github.com/messagebird/go-rest-api"
	"github.com/messagebird/go-rest-api/sms"
	log "github.com/sirupsen/logrus"
)

// SendSMSViaMessageBird sends sms notifications with Message Bird API
func SendSMSViaMessageBird(client *messagebird.Client) SendNotificationFunc {
	return func(msg *buffer.Message, recipients []*buffer.Recipient) error {
		var (
			mbrecipients []string
		)

		for _, recipient := range recipients {
			mbrecipients = append(mbrecipients, recipient.PhoneNumber)
		}

		log.Infof("sending messages from originator %s to recipients %v", msg.Originator, mbrecipients)
		_, err := sms.Create(client, msg.Originator, mbrecipients, msg.Text, nil)
		if err != nil {
			return err
		}

		return nil
	}
}
