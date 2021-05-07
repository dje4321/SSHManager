package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type ConfigObject struct {
	Version     string
	Name        string
	Description string
	Username    string
	Hostname    string
	Port        uint16

	UseKey  bool
	KeyPath string

	SSHArgs []string
}

func NewConfigObject() (output *ConfigObject) {
	output = new(ConfigObject)
	output.Version = "v0.0.2"
	output.Name = "Error"
	output.Description = ""
	output.Username = "NULL"
	output.Hostname = "127.0.0.1"
	output.Port = 22

	output.UseKey = false
	output.KeyPath = "/dev/null"

	output.SSHArgs = []string{}

	return output
}

func Load(name string, configPath string) (output *ConfigObject) {
	output = new(ConfigObject)
	file, err := os.ReadFile(configPath + "/" + name + ".json")
	if err != nil {
		fmt.Println("Unable to locate ssh profile " + name)
		os.Exit(1)
	}

	json.Unmarshal(file, &output)
	return output
}

func (c *ConfigObject) Write(dir string) {
	file, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		fmt.Println("Failed to marshall data!")
		c._Debug_Print()
		panic(err)
	}

	err = os.WriteFile(dir+"/"+c.Name+".json", file, 0640)
	if err != nil {
		fmt.Println("Failed to write file!")
		c._Debug_Print()
		panic(err)
	}
}

func (c *ConfigObject) _Debug_Print() {
	output := os.Stderr
	output.Write([]byte("    " + "Version:" + c.Version + "\n"))
	output.Write([]byte("    " + "Name:" + c.Name + "\n"))
	output.Write([]byte("    " + "Description:" + c.Description + "\n"))
	output.Write([]byte("    " + "Username:" + c.Username + "\n"))
	output.Write([]byte("    " + "Hostname:" + c.Hostname + "\n"))
	output.Write([]byte("    " + "Port:" + strconv.Itoa(int(c.Port)) + "\n"))
	output.Write([]byte("    " + "UseKey:" + strconv.FormatBool(c.UseKey) + "\n"))
	output.Write([]byte("    " + "KeyPath:" + c.KeyPath + "\n"))

	output.Write([]byte("    " + "SSHArgs [" + "\n"))
	for _, val := range c.SSHArgs {
		output.Write([]byte("    " + "    " + val + ",\n"))
	}
	output.Write([]byte("    " + "]" + "\n"))
}
