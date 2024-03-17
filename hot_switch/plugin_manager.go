package hot_switch

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/pierrec/xxHash/xxHash32"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

const (
	SoSuffix       = ".so"
	DateTimeFormat = "20060102150405"
)

type PluginManager struct {
	Log       *log.Logger
	dirName   string
	pluginMap map[string]*Plugin
	when      time.Time
}

func NewPluginManger(logger *log.Logger) *PluginManager {
	now := time.Now().Format(DateTimeFormat)
	dirName := fmt.Sprintf("%s-%d", now, os.Getpid())

	return &PluginManager{
		Log:       logger,
		dirName:   dirName,
		pluginMap: make(map[string]*Plugin),
		when:      time.Now(),
	}
}

func (pm *PluginManager) loadPlugins(files []string, oldManager *PluginManager) error {
	if len(pm.pluginMap) != 0 {
		return errors.New("never call loadPlugins twice")
	}

	counters := make(map[string]int)
	for _, file := range files {
		counters[name2Key(pluginName(file))]++
	}
	for k, v := range counters {
		if v > 1 {
			return fmt.Errorf("duplicate name detected: %s. files: %s", k, strings.Join(files, ", "))
		}
	}

	infoMap, err := buildFileInfoMap(files)
	if err != nil {
		return err
	}

	if oldManager != nil {
		notChanged := make(map[string]*fileInfo)
		for k, info := range infoMap.m {
			if oldP := oldManager.pluginMap[k]; oldP != nil {
				if oldP.FileSha1 == info.fileSha1 {
					notChanged[k] = info
					delete(infoMap.m, k)
				}
			}
		}
		for k := range notChanged {
			newP := oldManager.pluginMap[k]
			newP.unchanged = true
			newP.Remark = "unchanged"
			pm.pluginMap[k] = newP
		}
	}

	if infoMap.Len()+len(pm.pluginMap) != len(files) {
		return fmt.Errorf("infoMap.len() + len(pm.pluginMap) != len(files)")
	}

	for _, info := range infoMap.m {
		if err = pm.loadPlugin(info); err != nil {
			return fmt.Errorf("failed to load the dy: %s, error: %w", info.name, err)
		}
	}

	return nil
}

func (pm *PluginManager) loadPlugin(info *fileInfo) error {
	pm.Log.Printf("plugin: %s\n", info.name)
	actual, err := pm.copyPlugin(info)
	if err != nil {
		return err
	}

	pg := NewPlugin()
	pg.Name = info.name
	pg.File = info.file
	pg.FileSha1 = info.fileSha1
	pm.Log.Printf("plugin: %s\n", actual)
	pg.p, err = plugin.Open(actual)
	if err != nil {
		return err
	}

	var missing []string
	if e := pg.Lookup("InvokeFunc", &pg.InvokeFunc); e != nil {
		if errors.Is(e, NotExistErr) {
			missing = append(missing, "InvokeFunc")
		} else {
			return err
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing functions: %s", strings.Join(missing, ", "))
	}

	pm.pluginMap[name2Key(pg.Name)] = pg
	return nil
}

func (pm *PluginManager) copyPlugin(info *fileInfo) (string, error) {
	tmpDir := filepath.Join(filepath.Dir(info.file), "tmp", pm.dirName)
	if err := os.MkdirAll(tmpDir, 0744); err != nil {
		return "", err
	}

	sum := xxHash32.Checksum(info.fileSha1[:], 0)
	dst := filepath.Join(tmpDir, fmt.Sprintf("%s-%#8x%s", info.name, sum, SoSuffix))
	return dst, os.WriteFile(dst, info.fileData, 0644)
}

func (pm *PluginManager) FindPlugin(name string) *Plugin {
	return pm.pluginMap[name2Key(name)]
}

func (pm *PluginManager) Invoke(name string, params ...any) ([]any, error) {
	invokeImpl := func(p *Plugin, funcName string) ([]any, error) {
		defer func() {
			if r := recover(); r != nil {
				pm.Log.Printf("<hotSwitch: %s> panic: %v\n%s", p.Name, r, debug.Stack())
			}
		}()

		if rt, err := p.InvokeFunc(funcName, params...); err != nil {
			pm.Log.Println(err)
			return nil, err
		} else {
			return rt, nil
		}
	}

	stirs := strings.Split(name, ".")
	p := pm.FindPlugin(stirs[0])
	return invokeImpl(p, stirs[1])
}

type fileInfo struct {
	name     string
	file     string
	fileData []byte
	fileSha1 [sha1.Size]byte
}

type fileInfoMap struct {
	m map[string]*fileInfo
}

func (infoMap *fileInfoMap) names() []string {
	var names []string
	for _, info := range infoMap.m {
		names = append(names, info.name)
	}
	sort.Strings(names)
	return names
}

func (infoMap *fileInfoMap) Len() int {
	return len(infoMap.m)
}

func (infoMap *fileInfoMap) add(file string, fileData []byte) {
	name := pluginName(file)
	fileSha1 := sha1.Sum(fileData)
	info := &fileInfo{
		name:     name,
		file:     file,
		fileData: fileData,
		fileSha1: fileSha1,
	}
	k := name2Key(name)
	infoMap.m[k] = info
}

func pluginName(file string) string {
	return strings.TrimSuffix(filepath.Base(file), SoSuffix)
}

func name2Key(name string) string {
	return strings.ToLower(name)
}

func buildFileInfoMap(files []string) (fileInfoMap, error) {
	infoMap := fileInfoMap{
		m: make(map[string]*fileInfo),
	}
	for _, file := range files {
		if data, err := os.ReadFile(file); err != nil {
			return infoMap, err
		} else {
			infoMap.add(file, data)
		}
	}
	return infoMap, nil
}
