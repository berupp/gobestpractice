package options

import "fmt"

type connection struct {
	ip   string
	port int
}
type connectionOption func(*connection)

func WithIP(ip string) func(connection *connection) {
	return func(connection *connection) {
		connection.ip = ip
	}
}
func WithPort(port int) func(connection *connection) {
	return func(connection *connection) {
		connection.port = port
	}
}

func New(opts ...connectionOption) *connection {
	conn := &connection{
		ip:   "default",
		port: 0,
	}
	//Apply all options
	for idx := range opts {
		opts[idx](conn)
	}
	return conn
}

func (c connection) ToString() string {
	return fmt.Sprintf("ip: %s, port: %d", c.ip, c.port)
}
