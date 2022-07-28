package ev

// Handler is a multiplexer over multiple subsystems waiting for events in
// parallel. When an event is delivered from one subsystem, handler bounces the
// message to all other listening handlers.
//
// Handler is thread safe, provided that handlers are only added in sequence
// and it is only ever passed by value after initialisation completes.
type Handler struct {
	msg  chan int
	hndl []chan int
}

func NewHandler() *Handler {
	return &Handler{
		make(chan int, 1),
		make([]chan int, 0, 3),
	}
}

func (h Handler) Run() {
	for msg := range h.msg {
		for _, elem := range h.hndl {
			elem <- msg
		}
	}
}
