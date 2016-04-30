package main

import (
	"pkg.linuxdeepin.com/lib/pulse"
	"sort"
)

const (
	CardBuildin   = 0
	CardBluethooh = 1
	CardUnknow    = 2
)

const (
	PropDeviceBus        = "device.bus"
	PropDeviceFromFactor = "device.form_factor"
)

type cProfileInfos []pulse.ProfileInfo2

func cardType(c *pulse.Card) int {
	if c.PropList[PropDeviceFromFactor] == "internal" {
		return CardBuildin
	}
	if c.PropList[PropDeviceBus] == "bluetooth" {
		return CardBluethooh
	}
	return CardUnknow
}

func profileBlacklist(c *pulse.Card) map[string]string {
	switch cardType(c) {
	case CardBluethooh:
		// TODO: bluez not full support headset_head_unit, please skip
		return map[string]string{
			"off":               "true",
			"headset_head_unit": "true",
		}
	case CardBuildin, CardUnknow:
		fallthrough
	default:
		return map[string]string{"off": "true"}
	}
}

func (infos cProfileInfos) Len() int {
	return len(infos)
}

func (infos cProfileInfos) Less(i, j int) bool {
	return infos[i].Priority > infos[j].Priority
}

func (infos cProfileInfos) Swap(i, j int) {
	infos[i], infos[j] = infos[j], infos[i]
}

func reselectProfile(card *pulse.Card) bool {
	blacklist := profileBlacklist(card)
	if blacklist[card.ActiveProfile.Name] != "true" {
		return false
	}

	var profiles cProfileInfos
	for _, p := range card.Profiles {
		if blacklist[p.Name] == "true" {
			continue
		}
		profiles = append(profiles, p)
	}

	sort.Sort(profiles)
	if len(profiles) < 1 {
		return false
	}

	logger.Info("Reselect profile:", profiles[0].Name)
	if profiles[0].Name == card.ActiveProfile.Name {
		return false
	}
	card.SetProfile(profiles[0].Name)
	return true
}
