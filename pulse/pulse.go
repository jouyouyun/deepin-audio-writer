package pulse

/*
#include "dde-pulse.h"
#cgo pkg-config: libpulse glib-2.0
*/
import "C"
import "fmt"
import "unsafe"
import "runtime"

type Callback func(eventType int, idx uint32)

const (
	EventTypeNew    = C.PA_SUBSCRIPTION_EVENT_NEW
	EventTypeChange = C.PA_SUBSCRIPTION_EVENT_CHANGE
	EventTypeRemove = C.PA_SUBSCRIPTION_EVENT_REMOVE
)
const (
	FacilityServer       = C.PA_SUBSCRIPTION_EVENT_SERVER
	FacilitySink         = C.PA_SUBSCRIPTION_EVENT_SINK
	FacilitySource       = C.PA_SUBSCRIPTION_EVENT_SOURCE
	FacilitySinkInput    = C.PA_SUBSCRIPTION_EVENT_SINK_INPUT
	FacilitySourceOutput = C.PA_SUBSCRIPTION_EVENT_SOURCE_OUTPUT
	FacilityCard         = C.PA_SUBSCRIPTION_EVENT_CARD
	FacilityClient       = C.PA_SUBSCRIPTION_EVENT_CLIENT
	FacilityModule       = C.PA_SUBSCRIPTION_EVENT_MODULE
	FacilitySampleCache  = C.PA_SUBSCRIPTION_EVENT_SAMPLE_CACHE
)

type Context struct {
	cbs map[int][]Callback

	ctx  *C.pa_context
	loop *C.pa_threaded_mainloop
}

func (c *Context) GetCardList() (r []*Card) {
	ck := newCookie()

	C.get_card_info_list(c.ctx, C.int64_t(ck.id))
	for _, info := range ck.ReplyList() {
		card := info.ToCard()
		if card == nil {
			continue
		}
		r = append(r, card)
	}
	return
}

func (c *Context) GetCard(index uint32) (*Card, error) {
	ck := newCookie()
	C.get_card_info(c.ctx, C.int64_t(ck.id), C.uint32_t(index))
	info := ck.Reply()
	if info == nil {
		return nil, fmt.Errorf("Can't obtain this instance for: %v", index)
	}

	card := info.ToCard()
	if card == nil {
		return nil, fmt.Errorf("'%d' not a valid card index", index)
	}
	return card, nil
}

func (c *Context) GetSinkList() (r []*Sink) {
	ck := newCookie()

	C.get_sink_info_list(c.ctx, C.int64_t(ck.id))
	for _, info := range ck.ReplyList() {
		sink := info.ToSink()
		if sink == nil {
			continue
		}
		r = append(r, sink)
	}
	return
}

func (c *Context) GetSink(index uint32) (*Sink, error) {
	ck := newCookie()
	C.get_sink_info(c.ctx, C.int64_t(ck.id), C.uint32_t(index))
	info := ck.Reply()
	if info == nil {
		return nil, fmt.Errorf("Can't obtain this instance for: %v", index)
	}

	sink := info.ToSink()
	if sink == nil {
		return nil, fmt.Errorf("'%d' not a valid sink index", index)
	}
	return sink, nil
}

func (c *Context) GetSinkInputList() (r []*SinkInput) {
	ck := newCookie()

	C.get_sink_input_info_list(c.ctx, C.int64_t(ck.id))
	for _, info := range ck.ReplyList() {
		si := info.ToSinkInput()
		if si == nil {
			continue
		}
		r = append(r, si)
	}
	return
}

func (c *Context) GetSinkInput(index uint32) (*SinkInput, error) {
	ck := newCookie()
	C.get_sink_input_info(c.ctx, C.int64_t(ck.id), C.uint32_t(index))

	info := ck.Reply()
	if info == nil {
		return nil, fmt.Errorf("Can't obtain this instance for: %v", index)
	}

	si := info.ToSinkInput()
	if si == nil {
		return nil, fmt.Errorf("'%d' not a valid sinkinput index", index)
	}
	return si, nil
}

func (c *Context) GetSourceList() (r []*Source) {
	ck := newCookie()

	C.get_source_info_list(c.ctx, C.int64_t(ck.id))
	for _, info := range ck.ReplyList() {
		source := info.ToSource()
		if source == nil {
			continue
		}
		r = append(r, source)
	}
	return
}

func (c *Context) GetSource(index uint32) (*Source, error) {
	ck := newCookie()
	C.get_source_info(c.ctx, C.int64_t(ck.id), C.uint32_t(index))

	info := ck.Reply()
	if info == nil {
		return nil, fmt.Errorf("Can't obtain this instance for: %v", index)
	}

	source := info.ToSource()
	if source == nil {
		return nil, fmt.Errorf("'%d' not a valid source index", index)
	}
	return source, nil
}

func (c *Context) GetServer() (*Server, error) {
	ck := newCookie()
	C.get_server_info(c.ctx, C.int64_t(ck.id))

	info := ck.Reply()
	if info == nil {
		return nil, fmt.Errorf("Can't obtain the server instance.")
	}

	s := info.ToServer()
	if s == nil {
		return nil, fmt.Errorf("Not found valid server")
	}
	return s, nil
}

func (c *Context) GetSourceOutputList() (r []*SourceOutput) {
	ck := newCookie()

	C.get_source_output_info_list(c.ctx, C.int64_t(ck.id))
	for _, info := range ck.ReplyList() {
		so := info.ToSourceOutput()
		if so == nil {
			continue
		}
		r = append(r, so)
	}
	return
}

func (c *Context) GetSourceOutput(index uint32) (*SourceOutput, error) {
	ck := newCookie()
	C.get_source_output_info(c.ctx, C.int64_t(ck.id), C.uint32_t(index))
	info := ck.Reply()
	if info == nil {
		return nil, fmt.Errorf("Can't obtain the this instance for: %v", index)
	}

	so := info.ToSourceOutput()
	if so == nil {
		return nil, fmt.Errorf("'%d' not a valid sourceoutput index", index)
	}
	return so, nil
}

func (c *Context) GetDefaultSource() string {
	return ""
}
func (c *Context) GetDefaultSink() string {
	return ""
}
func (c *Context) SetDefaultSink(name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	c.SafeDo(func() {
		C.pa_context_set_default_sink(c.ctx, cname, C.get_success_cb(), nil)
	})
}
func (c *Context) SetDefaultSource(name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	c.SafeDo(func() {
		C.pa_context_set_default_source(c.ctx, cname, C.get_success_cb(), nil)
	})
}

// SafeDo invoke an function with lock
func (c *Context) SafeDo(fn func()) {
	runtime.LockOSThread()
	C.pa_threaded_mainloop_lock(c.loop)
	fn()
	C.pa_threaded_mainloop_unlock(c.loop)
	runtime.UnlockOSThread()
}

var __context *Context

func GetContext() *Context {
	if __context == nil {
		loop := C.pa_threaded_mainloop_new()
		C.pa_threaded_mainloop_start(loop)
		ctx := C.pa_init(loop)

		__context = &Context{
			cbs:  make(map[int][]Callback),
			ctx:  ctx,
			loop: loop,
		}
	}
	return __context
}

//export receive_some_info
func receive_some_info(cookie int64, infoType int, info unsafe.Pointer, status int) {
	c := fetchCookie(cookie)
	if c == nil {
		fmt.Println("Warning: recieve_some_info with nil cookie", cookie, infoType, info, status)
		return
	}

	switch {
	case status == 1:
		c.EndOfList()
	case status == 0:
		c.Feed(infoType, info)
	case status < 0:
		c.Failed()
	}
}

func (c *Context) ConnectPeekDetect(cb func(idx int, v float64)) {
}

func (c *Context) Connect(facility int, cb func(eventType int, idx uint32)) {
	// sink sinkinput source sourceoutput
	c.cbs[facility] = append(c.cbs[facility], cb)
}

func (c *Context) handlePAEvent(facility, eventType int, idx uint32) {
	if cb, ok := c.cbs[facility]; ok {
		for _, c := range cb {
			go c(eventType, idx)
		}
	} else {
		fmt.Println("unknow event", facility, eventType, idx)
	}
}

//export go_handle_changed
func go_handle_changed(facility int, event_type int, idx uint32) {
	GetContext().handlePAEvent(facility, event_type, idx)
}
