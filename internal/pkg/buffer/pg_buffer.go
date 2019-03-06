package buffer

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	// Postgres driver
	_ "github.com/lib/pq"
)

// PostgresBuffer implements Buffer interface for Postgres
type PostgresBuffer struct {
	*sqlx.DB
}

// NewPostgresBuffer creates new instance of PostgresBuffer
func NewPostgresBuffer(connString string, maxConnections int) (*PostgresBuffer, error) {
	if len(connString) == 0 {
		return nil, errors.New("connection string cant be empty")
	}
	if maxConnections == 0 {
		maxConnections = -1
	}

	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to Postgres at %s", connString)
	}
	db.SetMaxOpenConns(maxConnections)

	return &PostgresBuffer{
		DB: db,
	}, nil
}

// PopNextMessage takes next available message waiting for processing from messages queue and marks it as processed
func (pb *PostgresBuffer) PopNextMessage(ctx context.Context) (*Message, error) {
	var (
		message = &Message{}
	)
	tx, err := pb.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new transaction")
	}
	defer tx.Rollback()

	err = tx.GetContext(ctx, message, "SELECT message_id, originator, text  FROM messages WHERE message_id = (SELECT MIN(message_id) FROM messages WHERE processed=FALSE) FOR UPDATE")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}

		return nil, errors.Wrap(err, "failed to get maximum sequence number from projection")
	}

	_, err = tx.ExecContext(ctx, "UPDATE messages SET processed=true WHERE message_id=$1", message.MessageID)
	if err != nil {
		return nil, errors.Wrap(err, "failed mark message as processed")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return message, nil
}

// SaveMessageForRecipient saves message into the queue
func (pb *PostgresBuffer) SaveMessageForRecipient(ctx context.Context, phoneNumber, originator, text string) error {
	if len(phoneNumber) == 0 || len(originator) == 0 || len(text) == 0 {
		return errors.New("input arguments cant be empty")
	}

	tx, err := pb.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create new transaction")
	}
	defer tx.Rollback()

	var unprocessedMesages []*Message
	if err = tx.SelectContext(ctx, &unprocessedMesages, "SELECT message_id, originator, text, processed FROM messages WHERE originator=$1 AND text=$2 AND processed = FALSE FOR UPDATE", originator, text); err != nil {
		return errors.Wrap(err, "failed to select unprocessed messages")
	}

	var msgID int64
	if len(unprocessedMesages) > 0 {
		msgID = unprocessedMesages[0].MessageID
	} else {
		stmt, err := tx.PrepareContext(ctx, "INSERT INTO messages (originator, text) VALUES($1, $2) RETURNING message_id")
		if err != nil {
			return errors.Wrap(err, "failed to prepare insert statement")
		}
		defer stmt.Close()

		err = stmt.QueryRowContext(ctx, originator, text).Scan(&msgID)
		if err != nil {
			return errors.Wrap(err, "failed to save message")
		}

	}

	_, err = tx.NamedExecContext(ctx, "INSERT INTO recipients (message_id, phone_number) VALUES(:message_id, :phone_number) ON CONFLICT DO NOTHING", &Recipient{MessageID: msgID, PhoneNumber: phoneNumber})
	if err != nil {
		return errors.Wrap(err, "failed to save recipient")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// GetRecipientsForMessageID returns list of recipients for message ID
func (pb *PostgresBuffer) GetRecipientsForMessageID(ctx context.Context, messageID int64) ([]*Recipient, error) {
	var recipients []*Recipient

	err := pb.SelectContext(ctx, &recipients, "SELECT message_id, phone_number FROM recipients WHERE message_id = $1", messageID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get maximum sequence number from projection")
	}

	return recipients, nil
}
