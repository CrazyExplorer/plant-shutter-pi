package camera

import (
	"strconv"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
	"go.uber.org/zap"

	"plant-shutter-pi/pkg/ov"
	"plant-shutter-pi/pkg/utils"
)

type Settings map[v4l2.CtrlID]v4l2.CtrlValue

var (
	logger       *zap.SugaredLogger
	initSettings = Settings{
		//10094849: 1,  // Auto Exposure: Auto Mode
		//10094868: 0,  // White Balance, Auto & Preset: Manual
		//10094872: 0,  // ISO Sensitivity, Auto: Manual

		//10291459: 90, // Compression Quality: 90
	}
	knownCtrlID = []v4l2.CtrlID{
		10094849, // Auto Exposure: Auto Mode
		10094868, // White Balance, Auto & Preset: Manual
		10094872, // ISO Sensitivity, Auto: Manual
		9963807,  // Color Effects: Set Cb/Cr

		9963776,  // Brightness
		9963777,  // Contrast
		9963778,  // Saturation
		9963790,  // Red Balance
		9963791,  // Blue Balance
		9963803,  // Sharpness
		9963818,  // Color Effects, CbCr
		10094850, // Exposure Time, Absolute
		10094871, // ISO Sensitivity
		10291459, // Compression Quality
	}
)

func init() {
	logger = utils.GetLogger()
}

func InitControls(dev *device.Device) {
	ApplySettings(dev, initSettings)
}

func ApplySettings(dev *device.Device, settings Settings) {
	for k, v := range settings {
		if err := dev.SetControlValue(k, v); err != nil {
			logger.Warnf("set ctrl(%d) to %d, err: %s", k, v, err)
		}
	}
}

func GetKnownCtrlConfigs(dev *device.Device) ([]ov.Config, error) {
	var res []ov.Config
	for _, id := range knownCtrlID {
		ctrl, err := v4l2.GetControl(dev.Fd(), id)
		if err != nil {
			logger.Warnf("The device does not support control(%d)", id)
			continue
		}
		cfg, err := ctrlToConfig(ctrl)
		if err != nil {
			return nil, err
		}
		res = append(res, cfg)
	}

	return res, nil
}

func GetKnownCtrlSettings(dev *device.Device) (Settings, error) {
	res := make(Settings)
	for _, id := range knownCtrlID {
		ctrl, err := v4l2.GetControl(dev.Fd(), id)
		if err != nil {
			continue
		}
		res[ctrl.ID] = ctrl.Value
	}

	return res, nil
}

func ctrlToConfig(ctrl v4l2.Control) (ov.Config, error) {
	res := ov.Config{
		ID:      ctrl.ID,
		Value:   ctrl.Value,
		Name:    ctrl.Name,
		Minimum: ctrl.Minimum,
		Maximum: ctrl.Maximum,
		Step:    ctrl.Step,
	}
	if !ctrl.IsMenu() {
		return res, nil
	}

	res.IsMenu = true
	items, err := ctrl.GetMenuItems()
	if err != nil {
		return ov.Config{}, err
	}
	menu := make(map[uint32]string)
	for _, i := range items {
		if ctrl.Type == v4l2.CtrlTypeIntegerMenu {
			menu[i.Index] = strconv.FormatInt(utils.Str2int64(i.Name), 10)
		} else {
			menu[i.Index] = i.Name
		}
	}
	res.MenuItems = menu

	return res, nil
}
