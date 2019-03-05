package buffer_test

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"

	"github.com/arkadyb/demo_messenger/internal/pkg/buffer"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func TestPostgresBuffer_PopNextMessage(t *testing.T) {
	type fields struct {
		DB func() (*sqlx.DB, sqlmock.Sqlmock)
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *buffer.Message
		wantErr bool
	}{
		{
			"Success",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectBegin()
					mock.ExpectQuery(`SELECT message_id, originator, text  FROM messages WHERE message_id = \(SELECT MIN\(message_id\) FROM messages WHERE processed=FALSE\)`).WillReturnRows(
						sqlmock.NewRows([]string{"message_id", "originator", "text"}).AddRow(1, "MockedOriginator", "MockedText"))
					mock.ExpectExec("UPDATE messages SET processed=true WHERE message_id=\\$1$").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
			},
			&buffer.Message{
				MessageID:  1,
				Originator: "MockedOriginator",
				Text:       "MockedText",
				Processed:  false,
			},
			false,
		},
		{
			"Failed to start transaction",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)
					mock.ExpectBegin().WillReturnError(errors.New("error"))

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
			},
			nil,
			true,
		},
		{
			"No messages to be send",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectBegin()
					mock.ExpectQuery(`SELECT message_id, originator, text  FROM messages WHERE message_id = \(SELECT MIN\(message_id\) FROM messages WHERE processed=FALSE\)`).WillReturnError(sql.ErrNoRows)

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.fields.DB()
			pb := &buffer.PostgresBuffer{
				DB: db,
			}
			defer pb.Close()

			got, err := pb.PopNextMessage(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresBuffer.PopNextMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgresBuffer.PopNextMessage() = %v, want %v", got, tt.want)
			}
			if mock.ExpectationsWereMet() != nil {
				t.Error("Not all expectations were met")
			}
		})
	}
}

func TestPostgresBuffer_SaveMessageForRecipient(t *testing.T) {
	type fields struct {
		DB func() (*sqlx.DB, sqlmock.Sqlmock)
	}
	type args struct {
		ctx         context.Context
		phoneNumber string
		originator  string
		text        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Add new batch batch",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectBegin()
					mock.ExpectQuery(`^SELECT message_id, originator, text, processed FROM messages.*`).
						WillReturnRows(
							sqlmock.NewRows([]string{"message_id", "originator", "text"}))
					mock.ExpectPrepare(`^INSERT INTO messages \(originator, text\).*`).ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"message_id"}).AddRow(1))

					mock.ExpectExec(`^INSERT INTO recipients \(message_id, phone_number\).*`).WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
				"1234567",
				"MockedOriginator",
				"MockedText",
			},
			false,
		},
		{
			"Add to existing batch",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectBegin()
					mock.ExpectQuery(`^SELECT message_id, originator, text, processed FROM messages.*`).
						WillReturnRows(
							sqlmock.NewRows([]string{"message_id", "originator", "text"}).AddRow(1, "MockedOriginator", "MockedText"))

					mock.ExpectExec(`^INSERT INTO recipients \(message_id, phone_number\).*`).WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
				"1234567",
				"MockedOriginator",
				"MockedText",
			},
			false,
		},
		{
			"DB Error",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectBegin().WillReturnError(errors.New("error"))

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
				"1234567",
				"MockedOriginator",
				"MockedText",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.fields.DB()
			pb := &buffer.PostgresBuffer{
				DB: db,
			}
			defer pb.Close()

			if err := pb.SaveMessageForRecipient(tt.args.ctx, tt.args.phoneNumber, tt.args.originator, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("PostgresBuffer.SaveMessageForRecipient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if mock.ExpectationsWereMet() != nil {
				t.Error("Not all expectations were met")
			}
		})
	}
}

func TestPostgresBuffer_GetRecipientsForMessageID(t *testing.T) {
	type fields struct {
		DB func() (*sqlx.DB, sqlmock.Sqlmock)
	}
	type args struct {
		ctx       context.Context
		messageID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*buffer.Recipient
		wantErr bool
	}{
		{
			"Success",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectQuery(`^SELECT message_id, phone_number FROM recipients.*`).
						WillReturnRows(
							sqlmock.NewRows([]string{"message_id", "phone_number"}).AddRow(1, "12345678").AddRow(2, "0987654"))

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
				1,
			},
			[]*buffer.Recipient{
				{
					1,
					"12345678",
				},
				{
					2,
					"0987654",
				},
			},
			false,
		},
		{
			"No rows",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectQuery(`^SELECT message_id, phone_number FROM recipients.*`).WillReturnError(sql.ErrNoRows)

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
				1,
			},
			nil,
			false,
		},
		{
			"DB Error",
			fields{
				func() (*sqlx.DB, sqlmock.Sqlmock) {
					db, mock, _ := sqlmock.New()
					mock.MatchExpectationsInOrder(true)

					mock.ExpectQuery(`^SELECT message_id, phone_number FROM recipients.*`).WillReturnError(errors.New("error"))

					return sqlx.NewDb(db, "sqlmock"), mock
				},
			},
			args{
				context.Background(),
				1,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.fields.DB()
			pb := &buffer.PostgresBuffer{
				DB: db,
			}
			defer pb.Close()
			got, err := pb.GetRecipientsForMessageID(tt.args.ctx, tt.args.messageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresBuffer.GetRecipientsForMessageID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgresBuffer.GetRecipientsForMessageID() = %v, want %v", got, tt.want)
			}
			if mock.ExpectationsWereMet() != nil {
				t.Error("Not all expectations were met")
			}
		})
	}
}
