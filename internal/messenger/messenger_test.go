package messenger_test

import (
	"context"
	"github.com/arkadyb/demo_messenger/internal/messenger"
	"github.com/arkadyb/demo_messenger/internal/pkg/buffer"
	"github.com/arkadyb/demo_messenger/internal/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

type MockedBuffer struct {
	mock.Mock
}

func (mb *MockedBuffer) PopNextMessage(ctx context.Context) (msg *buffer.Message, err error) {
	args := mb.Called(ctx)

	if args.Get(1) != nil {
		err = args.Error(1)
	}

	if args.Get(0) != nil {
		msg = args.Get(0).(*buffer.Message)
	}

	return
}

func (mb *MockedBuffer) GetRecipientsForMessageID(ctx context.Context, id int64) (rcpts []*buffer.Recipient, err error) {
	args := mb.Called(ctx, id)

	if args.Get(1) != nil {
		err = args.Error(1)
	}

	if args.Get(0) != nil {
		rcpts = args.Get(0).([]*buffer.Recipient)
	}

	return
}

func (mb *MockedBuffer) SaveMessageForRecipient(ctx context.Context, phoneNumber, originator, text string) (err error) {
	args := mb.Called(ctx, phoneNumber, originator, text)

	if args.Get(0) != nil {
		err = args.Error(0)
	}

	return
}

func TestMessenger_EnqueueSMS(t *testing.T) {
	type fields struct {
		errors chan error
		buffer func() buffer.Buffer
	}
	type args struct {
		ctx context.Context
		sms *types.SMS
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Success",
			fields{
				nil,
				func() buffer.Buffer {
					mock := &MockedBuffer{}
					mock.On("SaveMessageForRecipient", context.Background(), "12345", "originator", "some text").Return(nil)
					return mock
				},
			},
			args{
				context.Background(),
				&types.SMS{
					Recipient:  "12345",
					Originator: "originator",
					Message:    "some text",
				},
			},
			false,
		},
		{
			"Bad args",
			fields{
				nil,
				func() buffer.Buffer {
					return nil
				},
			},
			args{
				context.Background(),
				nil,
			},
			true,
		},
		{
			"Error in SaveMessageForRecipient",
			fields{
				nil,
				func() buffer.Buffer {
					mock := &MockedBuffer{}
					mock.On("SaveMessageForRecipient", context.Background(), "12345", "originator", "some text").Return(errors.New("error"))
					return mock
				},
			},
			args{
				context.Background(),
				&types.SMS{
					Recipient:  "12345",
					Originator: "originator",
					Message:    "some text",
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := messenger.NewMessenger(nil, tt.fields.buffer())
			if err := a.EnqueueSMS(tt.args.ctx, tt.args.sms); (err != nil) != tt.wantErr {
				t.Errorf("Messenger.EnqueueSMS() error = %v, wantErr %v", err, tt.wantErr)
			}
			a.Shutdown()
		})
	}
}

func TestMessenger_Processing(t *testing.T) {
	type params struct {
		expectedCount int
		timeout       time.Duration
	}

	tests := []struct {
		name     string
		sendFunc messenger.SendNotificationFunc
		buff     func() buffer.Buffer
		params   params
	}{
		{
			"Success",
			func(msg *buffer.Message, recipients []*buffer.Recipient) error {
				logrus.Infof("%v %v\n", msg, recipients)
				return nil
			},
			func() buffer.Buffer {
				buff := &MockedBuffer{}
				buff.On("PopNextMessage", context.Background()).Return(&buffer.Message{
					1,
					"originator",
					"text",
					true,
				}, nil)

				buff.On("GetRecipientsForMessageID", context.Background(), int64(1)).Return([]*buffer.Recipient{
					{
						MessageID:   int64(1),
						PhoneNumber: "12345",
					},
					{
						MessageID:   int64(1),
						PhoneNumber: "67899",
					},
				}, nil)

				return buff
			},
			params{
				5,
				6 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopChan := make(chan bool)
			counter := 0
			f := func(msg *buffer.Message, recipients []*buffer.Recipient) error {
				counter++
				if counter == tt.params.expectedCount {
					close(stopChan)
				}
				return nil
			}
			a := messenger.NewMessenger(f, tt.buff())

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				select {
				case <-time.NewTimer(tt.params.timeout).C:
					wg.Done()
					return
				case _, open := <-stopChan:
					if !open {
						wg.Done()
						return
					}
				}
			}()
			wg.Wait()

			assert.Equal(t, tt.params.expectedCount, counter)
			a.Shutdown()
		})
	}
}
