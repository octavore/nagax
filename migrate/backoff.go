package migrate

import backoff "gopkg.in/cenkalti/backoff.v2"

func (m *Module) SetBackOff(b backoff.BackOff) {
	m.backoff = b
}
