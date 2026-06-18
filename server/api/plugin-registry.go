package api

import "sync"

var registeredPlugins []SeatsurfingPlugin
var registeredPluginsMu sync.RWMutex

func RegisterPlugin(plg SeatsurfingPlugin) {
	registeredPluginsMu.Lock()
	defer registeredPluginsMu.Unlock()
	registeredPlugins = append(registeredPlugins, plg)
}

func GetPlugins() []SeatsurfingPlugin {
	registeredPluginsMu.RLock()
	defer registeredPluginsMu.RUnlock()
	return registeredPlugins
}
