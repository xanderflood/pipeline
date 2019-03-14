package pipeline

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/xanderflood/pipeline/generic"
)

//Feed starts a goroutine to feed data into the channel
func Feed(ctx context.Context, spec OutputChannelSpec, finishHook func()) *generic.Channel {
	ch := spec.Build()
	go func() {
		// defer ginkgo.GinkgoRecover()
		defer finishHook()
		defer ch.Close()

		spec.WaitUntil()

		spew.Dump("have", spec.Sends)
		for _, recv := range spec.Sends {
			fmt.Println("got", recv)
			if ok := ch.SendContext(ctx, recv); !ok {
				return
			}
			fmt.Println("sent", recv)
		}
	}()

	return ch
}

//Handle starts a goroutine to handle data from the channel
func Handle(ctx context.Context, spec InputChannelSpec, channel interface{}, finishHook func()) {
	ch := spec.Validate(channel)
	spew.Dump(ch)
	go func() {
		// defer ginkgo.GinkgoRecover()
		defer finishHook()

		for {
			recv, ok := ch.ReceiveContext(ctx)
			if !ok {
				break
			}

			spec.OnReceive(recv)
		}
	}()
}
