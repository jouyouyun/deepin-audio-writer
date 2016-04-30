package main

import (
	"bytes"
	"dbus/com/deepin/daemon/audio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"pkg.linuxdeepin.com/lib/dbus"
	"pkg.linuxdeepin.com/lib/glib-2.0"
	"pkg.linuxdeepin.com/lib/log"
	"pkg.linuxdeepin.com/lib/pulse"
	"sort"
	"sync"
)

const (
	audioHelperFile = ".config/deepin_audio_helper.conf"

	audioDest = "com.deepin.daemon.Audio"
	audioPath = "/com/deepin/daemon/Audio"

	dbusDest = "com.deepin.helper.AudioSaver"
	dbusPath = "/com/deepin/helper/AudioSaver"
	dbusIFC  = "com.deepin.helper.AudioSaver"
)

type Manager struct {
	Info *AudioInfo
}

func (*Manager) GetDBusInfo() dbus.DBusInfo {
	return dbus.DBusInfo{
		Dest:       dbusDest,
		ObjectPath: dbusPath,
		Interface:  dbusIFC,
	}
}

type AudioInfo struct {
	ActiveProfile    string
	ActiveSink       string
	ActiveSinkPort   string
	ActiveSource     string
	ActiveSourcePort string

	SinkVolume   float64
	SourceVolume float64
}

var (
	locker   sync.Mutex
	upLocker sync.Mutex
	audioObj *audio.Audio
	ctx      = pulse.GetContext()
	logger   = log.NewLogger("audio/helper")
)

func (info *AudioInfo) Update() *AudioInfo {
	upLocker.Lock()
	defer upLocker.Unlock()
	v := getCurrentAudioInfo()
	if info.Equal(v) {
		logger.Info("Audio info equal")
		return info
	}

	logger.Info("Will update config(src, old):", info, v)
	err := saveConfig(v)
	if err != nil {
		logger.Info("Save config failed:", err)
		return info
	}
	return v
}

func (info *AudioInfo) Apply() {
	cards := ctx.GetCardList()
	if len(cards) != 0 {
		cards[0].SetProfile(info.ActiveProfile)
	}

	audioObj.SetDefaultSink(info.ActiveSink)
	sink := getDefaultSink()
	if sink != nil {
		sink.SetPort(info.ActiveSinkPort)
		sink.SetVolume(info.SinkVolume, false)
		audio.DestroyAudioSink(sink)
	}

	audioObj.SetDefaultSource(info.ActiveSource)
	source := getDefaultSource()
	if source != nil {
		source.SetPort(info.ActiveSourcePort)
		source.SetVolume(info.SourceVolume, false)
		audio.DestroyAudioSource(source)
	}
}

func (info *AudioInfo) Equal(v *AudioInfo) bool {
	if info.ActiveProfile != v.ActiveProfile ||
		info.ActiveSink != v.ActiveSink ||
		info.ActiveSinkPort != v.ActiveSinkPort ||
		info.ActiveSource != v.ActiveSource ||
		info.ActiveSourcePort != v.ActiveSourcePort ||
		info.SinkVolume != v.SinkVolume ||
		info.SourceVolume != v.SourceVolume {
		return false
	}
	return true
}

func readConfig() (*AudioInfo, error) {
	locker.Lock()
	defer locker.Unlock()

	var file = path.Join(os.Getenv("HOME"), audioHelperFile)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var reader = bytes.NewBuffer(content)
	dec := gob.NewDecoder(reader)
	var info AudioInfo
	err = dec.Decode(&info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func saveConfig(info *AudioInfo) error {
	if info == nil {
		return nil
	}

	locker.Lock()
	defer locker.Unlock()

	var writer bytes.Buffer
	enc := gob.NewEncoder(&writer)
	err := enc.Encode(info)
	if err != nil {
		return err
	}

	var file = path.Join(os.Getenv("HOME"), audioHelperFile)
	err = os.MkdirAll(path.Dir(file), 0755)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, writer.Bytes(), 0644)
}

func getCurrentAudioInfo() *AudioInfo {
	var info AudioInfo
	cards := ctx.GetCardList()
	if len(cards) != 0 {
		info.ActiveProfile = cards[0].ActiveProfile.Name
	}

	sink := getDefaultSink()
	if sink != nil {
		info.ActiveSink = sink.Name.Get()
		port := sink.ActivePort.Get()
		if len(port) > 0 {
			info.ActiveSinkPort = port[0].(string)
		}
		info.SinkVolume = sink.Volume.Get()
		audio.DestroyAudioSink(sink)
	}

	source := getDefaultSource()
	if source != nil {
		info.ActiveSource = source.Name.Get()
		port := source.ActivePort.Get()
		if len(port) > 0 {
			info.ActiveSourcePort = port[0].(string)
		}
		info.SourceVolume = source.Volume.Get()
		audio.DestroyAudioSource(source)
	}

	return &info
}

func getDefaultSink() *audio.AudioSink {
	p, err := audioObj.GetDefaultSink()
	if err != nil {
		logger.Error("Get default sink failed:", err)
		return nil
	}

	sink, _ := audio.NewAudioSink(audioDest, p)
	return sink
}

func getDefaultSource() *audio.AudioSource {
	p, err := audioObj.GetDefaultSource()
	if err != nil {
		logger.Error("Get default source failed:", err)
		return nil
	}

	source, _ := audio.NewAudioSource(audioDest, p)
	return source
}

func (info *AudioInfo) PrintAudioInfo() {
	fmt.Println("Current audio info:", info)
}

func (info *AudioInfo) initProfile() bool {
	cards := ctx.GetCardList()
	if len(cards) == 0 {
		return false
	}

	profiles := cProfileInfos(cards[0].Profiles)
	if len(profiles) == 0 {
		return false
	}
	sort.Sort(profiles)
	if profiles[0].Name == info.ActiveProfile {
		return false
	}

	logger.Info("Init profile:", profiles[0].Name)
	cards[0].SetProfile(profiles[0].Name)
	return true
}

func main() {
	var err error
	audioObj, err = audio.NewAudio(audioDest, audioPath)
	if err != nil {
		logger.Error("New audio failed:", err)
		return
	}

	info, err := readConfig()
	if err != nil {
		logger.Warning("Read audio helper config failed:", err)
		info = getCurrentAudioInfo()
		info.initProfile()
		saveConfig(info)
	} else {
		info.Apply()
	}
	info.PrintAudioInfo()

	// Fixed the app not exit when logout
	var m = &Manager{Info: info}
	err = dbus.InstallOnSession(m)
	if err != nil {
		logger.Error("Install session bus failed:", err)
		return
	}
	dbus.DealWithUnhandledMessage()
	ctx.Connect(pulse.FacilityCard, func(ty int, idx uint32) {
		logger.Debug("[Event] card:", ty, idx)
		switch ty {
		case pulse.EventTypeNew:
			card, err := ctx.GetCard(idx)
			if err != nil {
				logger.Warning("Get card failed:", idx, err)
				return
			}
			reselectProfile(card)
		case pulse.EventTypeChange:
			info = info.Update()
		}
	})
	ctx.Connect(pulse.FacilityServer, func(ty int, idx uint32) {
		logger.Debug("[Event] server:", ty, idx)
		info = info.Update()
	})
	ctx.Connect(pulse.FacilitySink, func(ty int, idx uint32) {
		logger.Debug("[Event] sink:", ty, idx)
		info = info.Update()
	})
	ctx.Connect(pulse.FacilitySource, func(ty int, idx uint32) {
		logger.Debug("[Event] source:", ty, idx)
		info = info.Update()
	})

	audioObj.Sinks.ConnectChanged(func() {
		logger.Debug("[Event] sinks changed")
		info = info.Update()
	})

	go glib.StartLoop()
	err = dbus.Wait()
	if err != nil {
		logger.Error("Lose dbus connect:", err)
		os.Exit(-1)
	}
	os.Exit(0)
}
