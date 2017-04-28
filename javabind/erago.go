package eragoj

import (
	"local/erago"
)

func Init(ui UI, baseDir string) (*InputPort, error) {
	err := erago.Init(uiAdapter{ui}, baseDir)
	if err != nil {
		return nil, err
	}
	return GetInputPort(), nil
}

func InitSingle(p Printer, baseDir string) (*InputPort, error) {
	return Init(singleUI{p}, baseDir)
}

func Main() error {
	return erago.Main()
}

func Quit() {
	erago.Quit()
}

func GetInputPort() *InputPort {
	return &InputPort{erago.GetInputPort()}
}
