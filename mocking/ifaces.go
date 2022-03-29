package mocking

type PersonInterface interface {
	PrintName() string
	PrintLastName() string
}

type Person struct {
	Name     string
	LastName string
}

func (m *Person) PrintName() string {
	if m == nil {
		return ""
	}
	return m.Name
}
func (m *Person) PrintLastName() string {
	if m == nil {
		return ""
	}
	return m.LastName
}
