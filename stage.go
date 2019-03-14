package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/xanderflood/pipeline/generic"
)

//DefaultStageTimeout is the default timeout for a stage
var DefaultStageTimeout = time.Second * 10

//FakeStageSpec specifies the behavior of a fake stage
type FakeStageSpec struct {
	Outputs []OutputChannelSpec
	Inputs  []InputChannelSpec
	Timeout time.Duration
}

//NewStageSpec creates a new stage spec with no inputs or output
func NewStageSpec() *FakeStageSpec {
	return &FakeStageSpec{Timeout: DefaultStageTimeout}
}

//WithTimeout adds output channels to a stage spec
func (spec *FakeStageSpec) WithTimeout(timeout time.Duration) *FakeStageSpec {
	spec.Timeout = timeout
	return spec
}

//WithOutputChannels adds output channels to a stage spec
func (spec *FakeStageSpec) WithOutputChannels(outputs ...OutputChannelSpec) *FakeStageSpec {
	spec.Outputs = append(spec.Outputs, outputs...)
	return spec
}

//WithInputChannels adds input channels to a stage spec
func (spec *FakeStageSpec) WithInputChannels(inputs ...InputChannelSpec) *FakeStageSpec {
	spec.Inputs = append(spec.Inputs, inputs...)
	return spec
}

//StarterFunc is the type used for starter stubs. The function will typically need to
//be wrapped with type assertions on the channels before it can be attached to a fake
//as a stub function.
type StarterFunc func(ctx context.Context, inputs ...interface{}) (<-chan error, []interface{})

//Starter returns a starter function for the specified stage.
//Also returns an error channel that will receive at most once,
//and that will be closed when all goroutines started here have
//exited.
func (spec *FakeStageSpec) Starter() StarterFunc {
	return func(ctx context.Context, inputs ...interface{}) (<-chan error, []interface{}) {
		/// validate inputs ///
		if len(inputs) != len(spec.Inputs) {
			panic(fmt.Sprintf("expected %v inputs, got %v", len(spec.Inputs), len(inputs)))
		}

		genericInputs := make([]*generic.Channel, len(inputs))
		for i := range spec.Inputs {
			genericInputs[i] = spec.Inputs[i].Validate(inputs[i])
		}

		/// fire up all the sub-goroutines, under a common context ///
		stageCtx, timeoutCancel := context.WithTimeout(ctx, spec.Timeout)
		defer timeoutCancel()
		stageCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		inputClosedSignal, allInputsClosed := StartCounter(len(spec.Inputs))
		for i := range spec.Inputs {
			Handle(
				stageCtx,
				spec.Inputs[i],
				genericInputs[i],
				func() { inputClosedSignal <- struct{}{} },
			)
		}

		genericOutputs := make([]*generic.Channel, len(spec.Outputs))
		outputChannels := make([]interface{}, len(spec.Outputs))
		outputClosedSignal, allOutputsClosed := StartCounter(len(spec.Outputs))
		for i := range spec.Outputs {
			genericOutputs[i] = Feed(
				stageCtx,
				spec.Outputs[i],
				// genericOutputs[i],
				func() { outputClosedSignal <- struct{}{} },
			)
			outputChannels[i] = genericOutputs[i].Channel
		}

		errs := make(chan error, 1)
		go func() {
			defer close(errs)
			defer func() {
				<-allInputsClosed
				<-allOutputsClosed
			}()

			//once all the inputs have been drained, cancel the context
			cancel()

			select {
			case <-allInputsClosed:
			case <-stageCtx.Done():
				errs <- context.Canceled
			}

			select {
			case <-allOutputsClosed:
			case <-stageCtx.Done():
				errs <- context.Canceled
			}
		}()

		return errs, outputChannels
	}
}
