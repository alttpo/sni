package ob

type Observable interface {
	Type() ObservableType
	Subscribe(observer Observer)
	Unsubscribe(observer Observer)
}

type Observer interface {
	Observe(object interface{})

	Equals(other Observer) bool
}
