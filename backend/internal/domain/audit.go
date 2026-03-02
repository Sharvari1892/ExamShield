package domain

type AuditEvent struct {
	PrevHash    string
	Payload     string
	Timestamp   string
	CurrentHash string
}