/*
    Copyright (c) 2014, Percona LLC and/or its affiliates. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package agent

import (
	"encoding/json"
	"errors"
	"github.com/percona/cloud-protocol/proto"
	"io/ioutil"
	"log"
	"os"
)

// Defaults
const (
	API_HOSTNAME = "cloud-api.percona.com"
	CONFIG_FILE  = "/etc/percona/agent.conf"
	DATA_DIR     = "/var/spool/percona"
	LOG_DIR      = "/var/log/percona"
	LOG_FILE     = "agent.log"
	LOG_LEVEL    = "info"
)

type Config struct {
	// Required, read-only:
	ApiKey    string
	AgentUuid string
	// API-controlled:
	ApiHostname string
	DataDir     string
	LogDir      string
	LogFile     string
	LogLevel    string
	PidFile     string
	// Local-only, hacker:
	Links   map[string]string
	Enable  []string
	Disable []string
	// Internal:
	ConfigDir string
}

// Load config from JSON file.
func LoadConfig(file string) *Config {
	config := &Config{}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalln(err)
		}
	} else {
		if err = json.Unmarshal(data, config); err != nil {
			log.Fatalln(err)
		}
	}
	return config
}

// Write config into  JSON file.
func WriteConfig(file string, cur *Config) error {

	b, err := json.MarshalIndent(cur, "", "    ")
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(file, b, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

// Apply current config, i.e. overwrite this config with current config.
func (c *Config) Apply(cur *Config) error {
	if cur.ApiHostname != "" {
		c.ApiHostname = cur.ApiHostname
	}
	if cur.ApiKey != "" {
		c.ApiKey = cur.ApiKey
	}
	if cur.AgentUuid != "" {
		c.AgentUuid = cur.AgentUuid
	}
	if cur.LogDir != "" {
		c.LogDir = cur.LogDir
	}
	if cur.LogFile != "" {
		c.LogFile = cur.LogFile
	}
	if cur.LogLevel != "" {
		_, ok := proto.LogLevelNumber[cur.LogLevel]
		if !ok {
			return errors.New("Invalid log level: " + cur.LogLevel)
		}
		c.LogLevel = cur.LogLevel
	}
	if cur.DataDir != "" {
		c.DataDir = cur.DataDir
	}
	c.PidFile = cur.PidFile
	c.Links = cur.Links
	c.Enable = cur.Enable
	c.Disable = cur.Disable
	return nil
}

func (c *Config) Enabled(option string) bool {
	for _, o := range c.Enable {
		if o == option {
			return true
		}
	}
	return false
}

func (c *Config) Disabled(option string) bool {
	for _, o := range c.Disable {
		if o == option {
			return true
		}
	}
	return false
}
