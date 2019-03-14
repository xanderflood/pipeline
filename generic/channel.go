package generic

import (
	"context"
	"fmt"
	"reflect"
)

//Channel a wrapper around a generic channel
type Channel struct {
	Underlying reflect.Type
	Channel    interface{}
}

//NewChannel makes a new generic channel of the given
//type and buffering
func NewChannel(underlying reflect.Type, buffer int) *Channel {
	return &Channel{
		Underlying: underlying,
		Channel:    reflect.MakeChan(reflect.ChanOf(reflect.BothDir, underlying), buffer).Interface(),
	}
}

//FromChannel makes a new generic channel out of an
//existing native channel.
func FromChannel(ch interface{}) *Channel {
	underlying := channelType(ch)
	return &Channel{
		Underlying: underlying,
		Channel:    ch,
	}
}

//Send blocks until i can be sent on c
func (c *Channel) Send(i interface{}) {
	if !c.sendable(i) {
		panic(fmt.Sprintf("expected `%s`, got `%s`", c.Underlying, reflect.TypeOf(i)))
	}

	reflect.Select(
		[]reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(c.Channel),
				Send: reflect.ValueOf(i),
			},
		},
	)
}

//SendNonBlocking sends and returns true if possible right
//away, otherwise does nothing and returns false
func (c *Channel) SendNonBlocking(i interface{}) bool {
	if !c.sendable(i) {
		panic(fmt.Sprintf("expected `%s`, got `%s`", c.Underlying, reflect.TypeOf(i)))
	}

	chosen, _, _ := reflect.Select(
		[]reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(c.Channel),
				Send: reflect.ValueOf(i),
			},
			reflect.SelectCase{
				Dir: reflect.SelectDefault,
			},
		},
	)
	return chosen == 0
}

//SendContext sends a message if possible before the context is
//canceled. If the context is canceled, it returns false.
func (c *Channel) SendContext(ctx context.Context, i interface{}) bool {
	if !c.sendable(i) {
		panic(fmt.Sprintf("expected `%s`, got `%s`", c.Underlying, reflect.TypeOf(i)))
	}

	chosen, _, _ := reflect.Select(
		[]reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(c.Channel),
				Send: reflect.ValueOf(i),
			},
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done()),
			},
		},
	)
	return chosen == 0
}

//Receive blocks until a value can be received and returns it.
//Returns true if the receive was a send, and false if it was
//a zero value from a closed channel.
func (c *Channel) Receive() (interface{}, bool) {
	_, value, ok := reflect.Select(
		[]reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(c.Channel),
			},
		},
	)
	return value.Interface(), ok
}

//ReceiveNonBlocking if a sent value is ready to be received, receives it
//and returns true. If no value is ready or if the channel is
//closed and empty, returns nil and false.
func (c *Channel) ReceiveNonBlocking() (interface{}, bool) {
	chosen, value, ok := reflect.Select(
		[]reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(c.Channel),
			},
			reflect.SelectCase{
				Dir: reflect.SelectDefault,
			},
		},
	)
	return value.Interface(), ok && chosen == 0
}

//ReceiveContext receives a message if possible before the context is
//canceled. If the context is canceled, or the channel is closed and
//drained, it returns false.
func (c *Channel) ReceiveContext(ctx context.Context) (interface{}, bool) {
	chosen, value, ok := reflect.Select(
		[]reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(c.Channel),
			},
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done()),
			},
		},
	)
	return value.Interface(), chosen == 0 && ok
}

//Close closes the underlying channel.
func (c *Channel) Close() {
	reflect.ValueOf(c.Channel).Close()
}

func (c *Channel) sendable(i interface{}) bool {
	return reflect.TypeOf(i).AssignableTo(c.Underlying)
}

func channelType(i interface{}) reflect.Type {
	typ := reflect.TypeOf(i)
	if typ.Kind() != reflect.Chan {
		panic(fmt.Sprintf("expected channel type, got `%s`", typ))
	}

	return typ.Elem()
}
