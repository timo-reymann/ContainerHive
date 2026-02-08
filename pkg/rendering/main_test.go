package rendering

import (
	"testing"

	"github.com/timo-reymann/ContainerHive/pkg/discovery"
)

func TestRenderProject(t *testing.T) {
	project, err := discovery.DiscoverProject(t.Context(), "../testdata/simple-project")
	if err != nil {
		t.Fatal(err)
	}
	temp := "../../example/dist"
	err = RenderProject(t.Context(), project, temp)
	if err != nil {
		t.Fatal(err)
	}
}
