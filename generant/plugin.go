package generant
import (
	"sync"
	"errors"
	"path"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"fmt"
)

type LoadConf func(raw []byte) ([]Generant, error)

type Generant interface {
	Catch()

	//Stop: maybe could continue this round catch, and insert into mysql
	Stop()

	//ForceStop: stop immediately, drop this round, only do some clean etc
	ForceStop()
}

var (
	pluginsMu    sync.Mutex
	regedPlugins = make(map[string]LoadConf)

	generants []Generant
)

func Register(name string, fLoadConf LoadConf) {
	pluginsMu.Lock()
	defer pluginsMu.Unlock()

	regedPlugins[name] = fLoadConf
}

func loadPluginConfig() error {
	//
	if generants != nil && len(generants) > 0 {
		return errors.New("can't load config twice")
	}

	if len(config.ConfFileNames) != len(config.ConfPluginNames) {
		return errors.New("the length of ConfFileNames and ConfPluginNames not same")
	}

	generants = []Generant{}

	config.ConfDir = path.Clean(config.ConfDir)

	log.Infof("%d plugin configs to load", len(config.ConfFileNames))

	for i, fn := range config.ConfFileNames {
		log.Infof("[%d/%d]...", i+1, len(config.ConfFileNames))
		bs, err := ioutil.ReadFile(config.ConfDir + "/" + fn)
		if err != nil {
			return err
		}

		if plugin, hv := regedPlugins[config.ConfPluginNames[i]]; !hv {
			return errors.New("Doesn't registed Plugin : " + config.ConfPluginNames[i])
		} else {
			gs, err := plugin(bs)
			if err != nil {
				return fmt.Errorf("Error when load %d th plugin : %s", i+1, err.Error())
			}
			generants = append(generants, gs...)
		}
	}

	return nil
}

