package ob

import (
	"fmt"
	"sync"
)

type ObservableType int

func (o ObservableType) String() string {
	switch o {
	case ObservableUnknown:
		return "Unknown"
	case ObservableObject:
		return "Object"
	case ObservableList:
		return "List"
	}
	return fmt.Sprintf("(unexpected ObservableType value %d)", o)
}

const (
	ObservableUnknown ObservableType = iota
	ObservableObject
	ObservableList
)

type ListOperation string

const (
	// ListInit initializes the contents of the list (or replaces an existing list with new contents)
	ListInit ListOperation = "init"
	// ListConcat appends elements to the end of the list
	ListConcat ListOperation = "concat"
)

type ListEvent struct {
	// Operation denotes what happened to the list
	Operation ListOperation `json:"op"`
	// Elements is the data supporting the operation
	Elements []interface{} `json:"e"`
}

// ObservableImpl represents either an observable Object or observable List.
// The first call to ObjectPublish() establishes the type as Object.
// The first call to ListAppendOne(), ListConcat(), or ListInit() establishes the type as List.
// The observable type cannot be changed once initialized; the call will panic if attempted.
// Calling Object() or List() will return nil until the type is established.
// List types publish ListEvent instances, reflecting the type of change made to the List.
// On first Subscribe(), an Object type will publish its last published state.
// On first Subscribe(), a List type will publish the entire list.
type ObservableImpl struct {
	lock      sync.Mutex
	observers []Observer

	observableType ObservableType

	object interface{}
	list   []interface{}
}

func NewObservable() *ObservableImpl {
	return &ObservableImpl{}
}

func (o *ObservableImpl) Type() ObservableType {
	return o.observableType
}

func (o *ObservableImpl) Subscribe(observer Observer) {
	if observer == nil {
		return
	}

	defer o.lock.Unlock()
	o.lock.Lock()

	// make sure only one instance is subscribed:
	o.unsubscribe(observer)
	o.observers = append(o.observers, observer)

	// send last published state:
	switch o.observableType {
	case ObservableUnknown:
		// intentionally do nothing here since the observable is not initialized yet
		break
	case ObservableObject:
		// objects publish the last published state on first subscribe:
		observer.Observe(o.object)
		break
	case ObservableList:
		// lists publish the entire list contents on first subscribe:
		observer.Observe(ListEvent{
			Operation: ListInit,
			Elements:  o.list,
		})
		break
	}
}

func (o *ObservableImpl) Unsubscribe(observer Observer) {
	if observer == nil {
		return
	}

	defer o.lock.Unlock()
	o.lock.Lock()

	if o.observers == nil {
		return
	}

	o.unsubscribe(observer)
}

func (o *ObservableImpl) unsubscribe(observer Observer) {
	for i := len(o.observers) - 1; i >= 0; i-- {
		if observer.Equals(o.observers[i]) {
			o.observers = append(o.observers[0:i], o.observers[i+1:]...)
		}
	}
}

func (o *ObservableImpl) enforceType(mustType ObservableType) {
	if o.observableType == ObservableUnknown {
		// set new type:
		o.observableType = mustType
	} else if o.observableType != mustType {
		// panic otherwise:
		panic(fmt.Errorf("observable attempted to change type from %s to %s", o.observableType, mustType))
	}
}

func (o *ObservableImpl) Object() interface{} {
	return o.object
}

func (o *ObservableImpl) ObjectPublish(object interface{}) {
	defer o.lock.Unlock()
	o.lock.Lock()

	o.enforceType(ObservableObject)
	o.object = object
	for _, observer := range o.observers {
		observer.Observe(object)
	}
}

func (o *ObservableImpl) List() []interface{} {
	return o.list
}

func (o *ObservableImpl) ListAppendOne(newElement interface{}) {
	defer o.lock.Unlock()
	o.lock.Lock()

	o.enforceType(ObservableList)
	o.list = append(o.list, newElement)
	newElements := []interface{}{newElement}
	for _, observer := range o.observers {
		observer.Observe(ListEvent{
			Operation: ListConcat,
			Elements:  newElements,
		})
	}
}

func (o *ObservableImpl) ListAppendMany(newElements []interface{}) {
	defer o.lock.Unlock()
	o.lock.Lock()

	o.enforceType(ObservableList)
	o.list = append(o.list, newElements...)
	for _, observer := range o.observers {
		observer.Observe(ListEvent{
			Operation: ListConcat,
			Elements:  newElements,
		})
	}
}

func (o *ObservableImpl) ListInit(newList []interface{}) {
	defer o.lock.Unlock()
	o.lock.Lock()

	o.enforceType(ObservableList)
	o.list = newList
	for _, observer := range o.observers {
		observer.Observe(ListEvent{
			Operation: ListInit,
			Elements:  newList,
		})
	}
}
