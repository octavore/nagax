package migrate

import backoff "gopkg.in/cenkalti/backoff.v1"

func (m *Module) SetBackOff(b backoff.BackOff) {
	m.backoff = b
}
