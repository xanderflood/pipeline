package pipeline

//FatalErrorChannelSpec provides a spec for an error channel
//that publishes zero or one times and then closes
func FatalErrorChannelSpec(err error) OutputChannelSpec {
	if err != nil {
		return Output(&err, err)
	} else {
		return Output(&err)
	}
}
