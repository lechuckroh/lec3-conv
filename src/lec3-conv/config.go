package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/olebedev/config"
)

// Config defines configuration
type Config struct {
	srcDir             string
	destDir            string
	destFilename       string
	width              int
	height             int
	quality            int
	maxCPU             int
	emptyLineThreshold float64
}

// LoadYaml loads *.yaml file
func (c *Config) LoadYaml(filename string) {
	cfg, err := config.ParseYamlFile(filename)
	if err != nil {
		log.Printf("Error : Failed to parse %v : %v\n", filename, err)
		return
	}

	fmt.Printf("Loading %v\n", filename)

	c.srcDir = cfg.UString("srcDir", "./")
	c.destDir = cfg.UString("destDir", "./output")
	c.destFilename = cfg.UString("destFilename", "${filename}")
	c.width = cfg.UInt("width", -1)
	c.height = cfg.UInt("height", -1)
	c.quality = cfg.UInt("quality", 100)
	c.maxCPU = cfg.UInt("maxCPU", runtime.NumCPU())
	if c.maxCPU <= 0 {
		c.maxCPU = runtime.NumCPU()
	}
	c.emptyLineThreshold, _ = cfg.Float64("emptyLineThreshold")
}

// FormatDestFilename formats destFilename pattern
func (c *Config) FormatDestFilename(filename string) string {
	result := strings.Replace(c.destFilename, "${filename}", filename, -1)
	base := strings.ToLower(GetBase(filename))
	result = strings.Replace(result, "${base}", base, -1)
	return result
}

// Print displays configurations
func (c *Config) Print() {
	fmt.Printf("srcDir : %v\n", c.srcDir)
	fmt.Printf("destDir : %v\n", c.destDir)
	fmt.Printf("destFilename : %v\n", c.destFilename)
	fmt.Printf("size : (%v, %v)\n", c.width, c.height)
	fmt.Printf("quality : %v%%\n", c.quality)
	fmt.Printf("maxCPU : %v\n", c.maxCPU)
	fmt.Printf("emptyLineThreshold : %v\n", c.emptyLineThreshold)
}

// NewConfig creates an instance of Config
func NewConfig(cfgFilename string, srcDir string, destDir string) *Config {
	config := Config{}

	if cfgFilename != "" {
		config.LoadYaml(cfgFilename)
	}
	if srcDir != "" {
		config.srcDir = srcDir
	}
	if destDir != "" {
		config.destDir = destDir
	}

	return &config
}
