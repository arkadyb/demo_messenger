package messenger

import (
	"context"
	"database/sql"
	"github.com/arkadyb/demo_messenger/internal/pkg/buffer"
	"github.com/arkadyb/demo_messenger/internal/types"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"time"
)

const BATCH_TIMEOUT = 1 * time.Second

// Application interface describes behaviour of the application
type Application interface {
	// EnqueueSMS used to enqueue sms notifications, places them into waiting queue
	EnqueueSMS(context.Context, *types.SMS) error
}

// SendNotificationFunc defines function used to deliver Message to Recipients
type SendNotificationFunc func(msg *buffer.Message, recipients []*buffer.Recipient) error

// NewMessenger creates new Messenger instance
func NewMessenger(sendNotification SendNotificationFunc, buf buffer.Buffer) *Messenger {
	a := &Messenger{
		buffer:            buf,
		timer:             time.NewTimer(BATCH_TIMEOUT),
		stopSignalChannel: make(chan bool),
	}

	go func() {
		for {
			select {
			case <-a.stopSignalChannel:
				a.timer.Stop()
				close(a.stopSignalChannel)
				return
			case <-a.timer.C:
				var (
					ctx = context.Background()
				)
				a.timer.Stop()

				nextMessage, err := a.buffer.PopNextMessage(ctx)
				if err == sql.ErrNoRows {
					// no rows were found, skip
					a.timer.Reset(BATCH_TIMEOUT)
					break
				}
				if err != nil {
					if a.Errors != nil {
						a.Errors <- errors.Wrap(err, "failed to pop next message")
					}
					a.timer.Reset(BATCH_TIMEOUT)
					break
				}

				if nextMessage != nil {
					recipients, err := a.buffer.GetRecipientsForMessageID(ctx, nextMessage.MessageID)
					if err != nil {
						a.Errors <- errors.Wrapf(err, "failed to get recipients for message %d", nextMessage.MessageID)
						a.timer.Reset(BATCH_TIMEOUT)
						break
					}
					if len(recipients) > 0 {
						if err := sendNotification(nextMessage, recipients); err != nil {
							if a.Errors != nil {
								a.Errors <- errors.Wrap(err, "failed to send notification")
								a.timer.Reset(BATCH_TIMEOUT)
								break
							}
						}
					}
				}

			}
		}
	}()

	return a
}

// Messenger implements Application interface
type Messenger struct {
	buffer buffer.Buffer

	timer             *time.Timer
	stopSignalChannel chan bool

	// errors channel delivers information of notifications notifications failed to be sent
	Errors chan error
}

// EnqueueSMS places sms into buffered queue
func (a *Messenger) EnqueueSMS(ctx context.Context, sms *types.SMS) error {
	if sms == nil {
		return errors.New("sms cant be nil")
	}

	if err := a.buffer.SaveMessageForRecipient(ctx, sms.Recipient, sms.Originator, sms.Message); err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to send sms")
	}

	return nil
}

// Shutdown gracefully stops application
func (a *Messenger) Shutdown() {
	shutdownTimer := time.NewTimer(5 * time.Second)
	a.stopSignalChannel <- true
	log.Infoln("gracefully shutting down application...")
	for {
		select {
		case _, open := <-a.stopSignalChannel:
			if !open {
				return
			}
		case <-shutdownTimer.C:
			log.Error("failed to gracefully shutdown application")
		}
	}
}
