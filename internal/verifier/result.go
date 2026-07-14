package verifier

type Status int

const (
	StatusUnknown  Status = iota
	StatusValid           // SMTP server confirmed mailbox exists (250)
	StatusInvalid         // SMTP server rejected mailbox (550)
	StatusCatchAll        // domain accepts all email regardless of mailbox
)

func (s Status) String() string {
	switch s {
	case StatusValid:
		return "valid"
	case StatusInvalid:
		return "invalid"
	case StatusCatchAll:
		return "catch_all"
	default:
		return "unknown"
	}
}
