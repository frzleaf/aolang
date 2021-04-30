package proxy

type Monitor struct {
}

func (m *Monitor) onError(err error) {
	LOG.Error(err)
}
