package plugin

import (
	"log"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
)

var pluginList = make([]*api.SeatsurfingPlugin, 0)
var pluginListOnce sync.Once

func GetPlugins() []*api.SeatsurfingPlugin {
	pluginListOnce.Do(func() {
		files, err := os.ReadDir(filepath.Join(GetConfig().FilesystemBasePath, GetConfig().PluginsSubPath))
		if err != nil {
			return
		}
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".so") {
				loadPlugin(f)
			}
		}
	})
	return pluginList
}

func AddPlugin(plg *api.SeatsurfingPlugin) {
	pluginList = append(pluginList, plg)
}

func loadPlugin(f os.DirEntry) {
	plg, err := plugin.Open(filepath.Join(GetConfig().FilesystemBasePath, GetConfig().PluginsSubPath, f.Name()))
	if err != nil {
		log.Println("Failed to load plugin", f.Name(), err)
		return
	}
	v, err := plg.Lookup("Plugin")
	if err != nil {
		log.Println("Failed to lookup Plugin in", f.Name(), err)
		return
	}
	castV, ok := v.(api.SeatsurfingPlugin)
	if !ok {
		log.Println("Failed to cast Plugin in", f.Name(), err)
		return
	}
	log.Println("Loaded plugin", f.Name())
	AddPlugin(&castV)
	//pluginList = append(pluginList, &castV)
}
