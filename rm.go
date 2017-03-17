package broadlink

type RmDevice struct {
	*BaseDevice
}

func newRM(dev *BaseDevice) *RmDevice {
	return &RmDevice{
		BaseDevice: dev,
	}
}

func (rm *RmDevice) Check() {

}

func (rm *RmDevice) Send(data []byte) {

}

func (rm *RmDevice) EnterLearning() {

}
