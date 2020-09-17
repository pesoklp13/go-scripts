package closeable

type closeable interface {
	Close() error
}

type onError = func(err error)

func CloseStream(stream closeable, callback onError) {
	err := stream.Close()

	if err != nil && callback != nil {
		callback(err)
	}
}
