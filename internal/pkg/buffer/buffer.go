package buffer

import (
	"context"
)

// Buffer describes behaviour of buffered store
type Buffer interface {
	// PopNextMessage takes next messages ready for delivery from the queue and marks it as processed
	PopNextMessage(context.Context) (*Message, error)
	// GetRecipientsForMessageID returns list of recipients for given message ID
	GetRecipientsForMessageID(context.Context, int64) ([]*Recipient, error)
	// SaveMessageForRecipient stores next message into waiting queue
	SaveMessageForRecipient(ctx context.Context, phoneNumber, originator, text string) error
}
