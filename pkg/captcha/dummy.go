package captcha

type dummyStore struct{}

func (d *dummyStore) Set(id string, value string) error {
	return nil
}

func (d *dummyStore) Get(id string, clear bool) string {
	return ""
}

func (d *dummyStore) Verify(id, answer string, clear bool) bool {
	return false
}
