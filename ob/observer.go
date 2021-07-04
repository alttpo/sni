package ob

type ObserverFunc func(object interface{})

type observerImpl struct {
	key      string
	observer ObserverFunc
}

func NewObserver(key string, observer ObserverFunc) Observer {
	return &observerImpl{
		key:      key,
		observer: observer,
	}
}

func (o *observerImpl) Equals(other Observer) bool {
	if otherImpl, ok := other.(*observerImpl); ok {
		return o.key == otherImpl.key
	}
	return false
}

func (o *observerImpl) Observe(object interface{}) {
	o.observer(object)
}
