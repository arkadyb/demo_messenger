package messenger

import (
	"encoding/json"
	"fmt"
	"github.com/arkadyb/demo_messenger/internal/pkg/buffer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// SendSMSViaMessageBird sends sms notifications with Message Bird API
func SendSMSViaTwilio(sid, token string, httpclient *http.Client) SendNotificationFunc {
	return func(msg *buffer.Message, recipients []*buffer.Recipient) error {
		var (
			urlStr = fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", sid)
		)

		for _, recipient := range recipients {
			req, err := http.NewRequest("POST", urlStr, newSms(msg.Originator, recipient.PhoneNumber, msg.Text))
			if err != nil {
				return errors.Wrapf(err, "failed to send sms to %s", recipient.PhoneNumber)
			}

			req.SetBasicAuth(sid, token)

			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			resp, err := httpclient.Do(req)
			if err != nil || (resp.StatusCode < 200 || resp.StatusCode >= 300) {
				strResp, _ := ioutil.ReadAll(resp.Body)
				err = fmt.Errorf("response %s", strResp)

				if err != nil {
					err = errors.Wrapf(err, "failed to send sms to %s", recipient.PhoneNumber)
				}

				return err
			}

			var data map[string]interface{}
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&data)
			if err == nil {
				log.Infof("message sent to %s with sid", recipient.PhoneNumber)
			}
		}

		return nil
	}
}

func newSms(from, to, body string) *strings.Reader {
	msgData := url.Values{}
	msgData.Set("To", to)
	msgData.Set("From", from)
	msgData.Set("Body", body)
	return strings.NewReader(msgData.Encode())
}
