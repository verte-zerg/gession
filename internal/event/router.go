package event

import (
	"github.com/verte-zerg/gession/pkg/logging"
)

const (
	MaxQueue = 100
)

var (
	logger = logging.GetInstance().WithGroup("event")
)

type Type string

type Consumer interface {
	GetInputCh() chan Event
}

type Producer interface {
	SetOutputCh(outputCh chan Event)
}

type ConsumerProducer interface {
	Consumer
	Producer
}

type Router struct {
	outputChs map[Type][]chan Event
	inputCh   chan Event
}

func New() *Router {
	return &Router{
		outputChs: make(map[Type][]chan Event),
		inputCh:   make(chan Event, MaxQueue),
	}
}

func (e *Router) Start() {
	logger.Info("starting event router")

	go e.proxyEvents()
}

func (e *Router) GetInputCh() chan Event {
	return e.inputCh
}

func (e *Router) RegisterConsumer(eventTypes []Type, consumer Consumer) {
	for _, eventType := range eventTypes {
		inputCh := consumer.GetInputCh()
		outputChs, ok := e.outputChs[eventType]

		if !ok {
			outputChs = make([]chan Event, 0)
		}

		outputChs = append(outputChs, inputCh)
		e.outputChs[eventType] = outputChs
	}
}

func (e *Router) RegisterProducer(producer Producer) {
	producer.SetOutputCh(e.inputCh)
}

func (e *Router) proxyEvents() {
	for {
		event := <-e.inputCh
		outputChs, ok := e.outputChs[event.Type]

		if !ok {
			return
		}

		for _, outputCh := range outputChs {
			outputCh <- event
		}
	}
}

func (e *Router) EmitEvent(event Event) {
	e.inputCh <- event
}
