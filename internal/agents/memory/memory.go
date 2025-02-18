package memory

const (
	MessageRoleAI   = "ai"
	MessageRoleUser = "user"
)

type Message struct {
	Role    string
	Content string
}
