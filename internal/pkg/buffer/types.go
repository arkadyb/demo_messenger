package buffer

type Message struct {
	MessageID  int64  `db:"message_id"`
	Originator string `db:"originator"`
	Text       string `db:"text"`
	Processed  bool   `db:"processed"`
}

type Recipient struct {
	MessageID   int64  `db:"message_id"`
	PhoneNumber string `db:"phone_number"`
}
