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
	plugins := make([]SeatsurfingPlugin, len(registeredPlugins))
	copy(plugins, registeredPlugins)
	return plugins
}

// ResetPluginsForTest clears every registered plugin. Test-only: production
// code has no supported way to unregister a plugin once connected, but tests
// that RegisterPlugin a fake need to undo it so it doesn't leak into other
// tests sharing the same process.
func ResetPluginsForTest() {
	registeredPluginsMu.Lock()
	defer registeredPluginsMu.Unlock()
	registeredPlugins = nil
}
