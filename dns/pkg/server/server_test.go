package server_test

type fakeUpstream struct {
	response []byte
	err      error
}

// fake upstream for later
func (f *fakeUpstream) SendQuery(_ []byte) ([]byte, error) {
	return f.response, f.err
}
