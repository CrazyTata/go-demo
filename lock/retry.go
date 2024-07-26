package lock

import "time"

type mustSetOperation func() (string, error)

type ErrNotify func(error)

func mustSetRetryNotify(operation mustSetOperation, b Backoff, notify ErrNotify) (string, error) {

	var err error
	var randVal string
	var wait time.Duration
	var retry bool
	var n int

	for {

		if randVal, err = operation(); err == nil {
			return randVal, nil
		}

		if b == nil {
			return "", err
		}

		n++
		wait, retry = b.Next(n)
		if !retry {
			return "", err
		}

		if notify != nil {
			notify(err)
		}

		time.Sleep(wait)
	}

}
