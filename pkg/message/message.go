package message

// Message - message info (matching the Nanit API)
type Message struct {
	// TODO: Determin format for ReadAt, SeenAt, and DismissedAt
	// TODO: unmarshall UpdatedAt and CreatedAt ISO8601 timestamp into time.Time
	// TODO: enumerate possible Data interface structures
	Id          int         `json:"id"`
	BabyUid     string      `json:"baby_uid"`
	UserId      int         `json:"user_id"`
	Type        string      `json:"type"`
	Time        UnixTime    `json:"time"`
	ReadAt      string      `json:"read_at"`
	SeenAt      string      `json:"seen_at"`
	DismissedAt string      `json:"dismissed_at"`
	UpdatedAt   string      `json:"updated_at"`
	CreatedAt   string      `json:"created_at"`
	Data        interface{} `json:"data"`
}

// FilterMessages allows a slice (?) of Messages to be filtered by an aribitrary function that returns true or false for each element, indicating whether it should be included in the filtered set or not
func FilterMessages(messages []Message, cond func(message Message) bool) []Message {
	var result []Message
	for i := range messages {
		if cond(messages[i]) {
			result = append(result, messages[i])
		}
	}

	return result
}
