package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	TraceLog struct {
		StartSignal struct {
			Type      string `yaml:"type"`
			Offset    uint   `yaml:"offset"`
			Value     uint64 `yaml:"value"`
			Direction string `yaml:"direction"`
			DataWidth uint   `yaml:"datawidth"`
		} `yaml:"startsignal"`
		StopSignal struct {
			Type      string `yaml:"type"`
			Offset    uint   `yaml:"offset"`
			Value     uint64 `yaml:"value"`
			Direction string `yaml:"direction"`
			DataWidth uint   `yaml:"datawidth"`
		} `yaml:"stopsignal"`
		Serial struct {
			Type                 string `yaml:"type"`
			Port                 string `yaml:"port"`
			BaudRate             int    `yaml:"baudrate"`
			ReadWriteTimeout     uint   `yaml:"timeout"`
			DeviceHotplugTimeout uint   `yaml:"hotplugtimeout"`
		} `yaml:"serial"`
		DutControl struct {
			StartCmd   string `yaml:"startcmd"`
			StopCmd    string `yaml:"stopcmd"`
			RestartCmd string `yaml:"restartcmd"`
			InitCmd    string `yaml:"initcmd"`
		} `yaml:"dutcontrol"`
		VariableFirmareOptions []struct {
			Name       string `yaml:"name"`
			ByteOffset uint   `yaml:"byteoffset"`
			BitWidth   uint   `yaml:"bitwidth"`
			Min        uint64 `yaml:"min"`
			Max        uint64 `yaml:"max"`
		} `yaml:"variable_options"`
		OptionsDefaultTable string `yaml:"options_default_table"`
	}
	Database struct {
		HostName string `yaml:"hostname"` // Ignoring for now
		Port     uint   `yaml:"port"`     // Ignoring for now
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}

func GetConfig() (Config, error) {
	f, err := os.Open("config.yml")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(&cfg)
	return cfg, err
}

//return all FirmwareOptions and their possible values as map
func GetConfigFirmwareOptionsByName(cfg Config) map[string][]uint64 {
	optionsset := map[string][]uint64{}

	for _, opt := range cfg.TraceLog.VariableFirmareOptions {
		for j := opt.Min; j <= opt.Max; j++ {
			optionsset[opt.Name] = append(optionsset[opt.Name], j)
		}
	}

	return optionsset
}
