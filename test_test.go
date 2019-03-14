package pipeline_test

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/xanderflood/pipeline"
)

func init() {
	//Question: can variables assigned to in the hooks
	//be accessed for assertions without creating data
	//races? How can I convince `go test -race` that
	//these accesses are safely synchronized?

	var (
		str string
		err error
		i   int

		queryErr   error
		succLogErr error
		failLogErr error
	)
	queryStubSpec := NewStageSpec().
		WithTimeout(time.Second*20).
		WithOutputChannels(
			Output(&str, "1", "2", "3", "4"),
			FatalErrorChannelSpec(&err, queryErr),
			Output(&i, "4"),
		)

	cleanseStubSpec := NewStageSpec().
		WithTimeout(time.Second*20).
		WithInputChannels(Input(&s, func(str interface{}) {
			id, ok := str.(string)
			Expect(ok).To(BeTrue())
		})).
		WithOutputsChannels(
			Output(&str, "1", "2", "3"),            //successes
			Output(&str, "4\tfailed bc bad stuff"), //errors
		)

	successLogStubSpec := NewStageSpec().
		WithTimeout(time.Second*20).
		WithInputChannels(Input(&s, func(str interface{}) {
			id, ok := str.(string)
			Expect(ok).To(BeTrue())

			//TODO save id to an array for inspection?
		})).
		WithOutputChannels(
			FatalErrorChannelSpec(&err, succLogErr),
			Output(&i, 3),
		)

	failureLogStubSpec := NewStageSpec().
		WithTimeout(time.Second*20).
		WithInputChannels(Input(&s, func(str interface{}) {
			line, ok := str.(string)
			Expect(ok).To(BeTrue())

			//TODO validate format?
		})).
		WithOutputChannels(
			FatalErrorChannelSpec(&err, failLogErr),
			Output(&i, 1),
		)
}
