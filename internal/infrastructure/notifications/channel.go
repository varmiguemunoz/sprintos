package notifications

type Event struct {
	Type    string
	OrgID   uint
	Title   string
	Details string
	URL     string
}

type Channel interface {
	Name() string
	Send(event Event) error
}
