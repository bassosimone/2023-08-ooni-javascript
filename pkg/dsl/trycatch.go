package dsl

// Try inspects the results of running a pipeline and returns an error if the
// returned error is an [ErrException], nil otherwise.
func Try[T any](results Maybe[T]) error {
	if IsErrException(results.Error) {
		return results.Error
	}
	return nil
}

// catch inspects a list of results and returns an error if there's an exception.
func catch[T any](results ...Maybe[T]) error {
	for _, entry := range results {
		if IsErrException(entry.Error) {
			return entry.Error
		}
	}
	return nil
}
