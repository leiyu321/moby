package libnetwork

import (
	"github.com/docker/libnetwork/drvregistry"
	"github.com/docker/libnetwork/tcmanagerapi"
	"github.com/docker/libnetwork/tcmanagers/null"
	"github.com/docker/libnetwork/tcmanagers/sample"
)

func initTrafficControlDrivers(r *drvregistry.DrvRegistry, config map[string]string) error {
	for _, fn := range [](func(tcmanagerapi.Callback, map[string]string) error){
		null.Init,
		sample.Init,
	} {
		if err := fn(r, config); err != nil {
			return err
		}
	}

	return nil
}
