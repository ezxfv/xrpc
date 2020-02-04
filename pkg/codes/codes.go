package codes

type Code uint

const (
	Unimplemented Code = 12
)

func CodeFromError(err error) string {
	if err == nil {
		return "ok"
	}
	return "unknown"
}
