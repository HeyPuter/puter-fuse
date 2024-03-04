package streamutil

type NullCloser struct{}

func (n *NullCloser) Close() error {
	return nil
}
