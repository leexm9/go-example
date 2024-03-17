package hot_switch

import (
	"bytes"
	"fmt"
	"go-example/hot_switch/utils"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type PluginSwitch struct {
	Logger        *log.Logger
	current       atomic.Value
	pluginDir     string
	reloadCounter int64
	mu            sync.Mutex
}

func NewPluginSwitch(logger *log.Logger, pluginDir string) *PluginSwitch {
	return &PluginSwitch{Logger: logger, pluginDir: pluginDir}
}

func (ps *PluginSwitch) Current() *PluginManager {
	val := ps.current.Load()
	pluginManager, _ := val.(*PluginManager)
	return pluginManager
}

func (ps *PluginSwitch) InitLoad() (Details, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return ps.loadPlugins()
}

func (ps *PluginSwitch) Reload() (Details, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	details, err := ps.loadPlugins()
	if err != nil {
		atomic.AddInt64(&ps.reloadCounter, 1)
	}
	return details, err
}

func (ps *PluginSwitch) loadPlugins() (Details, error) {
	var absDir string
	if err := utils.IsDirectory(ps.pluginDir, "pluginDir"); err != nil {
		return nil, err
	} else if absDir, err = filepath.Abs(ps.pluginDir); err != nil {
		return nil, err
	}

	dirs, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, item := range dirs {
		if item.IsDir() {
			continue
		}
		if strings.HasSuffix(item.Name(), SoSuffix) {
			files = append(files, filepath.Join(absDir, item.Name()))
		}
	}

	return ps.loadPluginFiles(files)
}

func (ps *PluginSwitch) loadPluginFiles(files []string) (Details, error) {
	if len(files) == 0 {
		return nil, nil
	}

	oldManager := ps.Current()
	newManager := NewPluginManger(ps.Logger)
	if err := newManager.loadPlugins(files, oldManager); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, file := range files {
		p := newManager.FindPlugin(pluginName(file))
		if p.Remark != "" {
			result[p.File] = p.Remark
		} else {
			result[p.File] = "OK"
		}
	}

	ps.current.Store(newManager)
	return result, nil
}

func (ps *PluginSwitch) ReloadCounter() int64 {
	return atomic.LoadInt64(&ps.reloadCounter)
}

type Details map[string]string

func (d Details) String() string {
	var tmp []string
	for k := range d {
		tmp = append(tmp, k)
	}
	sort.Strings(tmp)

	var out bytes.Buffer
	for i, k := range tmp {
		if i > 0 {
			_, _ = out.WriteString(", ")
		}
		x := strings.TrimSuffix(filepath.Base(k), SoSuffix)
		_, _ = fmt.Fprintf(&out, "%s: %s", x, d[k])
	}
	return out.String()
}
