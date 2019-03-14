package pipeline

import (
	"fmt"
	"reflect"

	"github.com/xanderflood/pipeline/generic"
)

//OutputChannelSpec specifies the behavior of an output channel,
//including its underlying type and what data should be sent through.
type OutputChannelSpec struct {
	Underlying reflect.Type
	Sends      []interface{}
	WaitUntil  func()
}

//Output checks that all seeds are assignable to the type pointed to by
//kind. It panics if this check fails, and otherwise it returns a newly
//constructed OutputChannelSpec.
func Output(
	kind interface{},
	sends ...interface{},
) OutputChannelSpec {
	baseType := TargetOf(kind)

	for _, s := range sends {
		if !reflect.TypeOf(s).AssignableTo(baseType) {
			panic(fmt.Sprintf("expected `%s`, got `%s`", baseType, s))
		}
	}

	return OutputChannelSpec{
		Underlying: baseType,
		Sends:      sends,
		WaitUntil:  func() {}, //no delay
	}
}

//WithDelay build a new output channel for implementing the spec
func (spec *OutputChannelSpec) WithDelay(f func()) *OutputChannelSpec {
	spec.WaitUntil = f
	return spec
}

//Build build a new output channel for implementing the spec
func (spec *OutputChannelSpec) Build() *generic.Channel {
	return generic.NewChannel(spec.Underlying, 0)
}

//InputChannelSpec specifies the behavior of a fake output channel,
//including its underlying type and what should be done with the data
//it receives.
type InputChannelSpec struct {
	Underlying reflect.Type
	OnReceive  func(interface{})
}

//Input checks that `handle` is a function that accepts a the type pointed
//to by `kind`. If not, it panics, but if all is well, it returns a new .
func Input(
	kind interface{},
	handle func(interface{}),
) InputChannelSpec {
	baseType := TargetOf(kind)

	return InputChannelSpec{
		Underlying: baseType,
		OnReceive:  handle,
	}
}

//Validate verifies that the given interface is a channel
//matching this spec, and returns a generic wrapper if so.
//Otherwise it panics.
func (spec *InputChannelSpec) Validate(ch interface{}) *generic.Channel {
	gCh := generic.FromChannel(ch)
	if !spec.Underlying.AssignableTo(gCh.Underlying) {
		panic(fmt.Sprintf("expected channel accepting `%s`, got channel of `%s`", spec.Underlying, gCh.Underlying))
	}

	return gCh
}
