package db

type StoredMessage struct {
	PeerID    string `json:"peer_id"`
	SenderID  string `json:"sender_id"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}

func (s *Store) SaveMessage(peerID, senderID, text string, ts int64) error {
	_, err := s.db.Exec("INSERT INTO messages (peer_id, sender_id, text, timestamp) VALUES (?, ?, ?, ?)", peerID, senderID, text, ts)
	return err
}

func (s *Store) LoadMessageByPeerID(peerID string) ([]StoredMessage, error) {
	rows, err := s.db.Query("SELECT peer_id, sender_id, text, timestamp FROM messages WHERE peer_id = ? ORDER BY timestamp ASC", peerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []StoredMessage
	for rows.Next() {
		var m StoredMessage
		if err := rows.Scan(&m.PeerID, &m.SenderID, &m.Text, &m.Timestamp); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}
