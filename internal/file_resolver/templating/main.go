package templating

import (
	"github.com/timo-reymann/ContainerHive/pkg/model"
)

type TemplateContext struct {
	Versions  model.Versions
	BuildArgs model.BuildArgs
	ImageName string
}

type Processor interface {
	Process(*TemplateContext, string, []byte) ([]byte, error)
}
