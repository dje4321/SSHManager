package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// 3 States can exsist for a arg
//   1. Not supplied aka NIL
//   2. Supplied but Invalid
//   3. Supplied and Valid

type Arg struct {
	Key   string
	Value string
	Arg   string
	Pos   int
	Valid bool
	Error string
}

type Menu struct {
	Args   []Arg
	Config *ConfigObject
}

func ErrorAndExit(err string, exit int) {
	stderr := os.Stderr
	stderr.Write([]byte(err + "\n"))
	os.Exit(exit)
}

func (menu *Menu) GetArgSlice(ValidKeys []string) (output []Arg) {
	output = []Arg{}
	for _, arg := range menu.Args {
		for _, key := range ValidKeys {
			if arg.Key == key {
				output = append(output, arg)
			}
		}
	}
	return output
}

func (menu *Menu) GetSingleArg(ValidKey string) Arg {
	for _, arg := range menu.Args {
		if arg.Key == ValidKey {
			return arg
		}
	}
	return Arg{}
}

func IsArg(arg string, allowList []string) bool {
	for _, val := range allowList {
		if val == arg {
			return true
		}
	}
	return false
}

func (menu *Menu) GetConfig() string {
	for _, arg := range menu.Args {
		if arg.Key == "config" && arg.Valid {
			return arg.Value
		}
	}

	output, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("No config specified and unable to locate user config direcotry.")
		panic(err)
	}
	output += "/sshmanager"

	return output
}

func (menu *Menu) Parse(argv []string) {
	New := func(Argument string, Key string, Pos int, Value string, Error string, Valid bool) Arg {
		output := Arg{
			Arg:   Argument,
			Key:   Key,
			Pos:   Pos,
			Value: Value,
			Error: Error,
			Valid: Valid,
		}
		return output
	}
	IsArgValid := func(argv []string, pos int, arg string) (valid bool, value string, err string) {
		valid = false
		value = ""
		err = ""

		if pos+1 < len(argv) && IsValidValue(argv[pos+1]) {
			value = argv[pos+1]
			valid = true
		} else {
			valid = false
			err = fmt.Sprintf("Unable to locate value for %s", arg)
		}

		return valid, value, err
	}

	for k, arg := range argv {
		switch {
		case IsArg(arg, []string{"-m", "--make"}):
			config := New(arg, "make", k, "", "", true)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-c", "--config"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "config", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-l", "--list"}):
			config := New(arg, "list", k, "", "", true)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-h", "--host"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "host", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-u", "--user"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "user", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-p", "--port"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "port", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-n", "--name"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "name", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-d", "--desc"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "desc", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-k", "--key"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "key", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-o", "--option"}):
			valid, value, error := IsArgValid(argv, k, arg)
			config := New(arg, "option", k, value, error, valid)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"--help"}):
			config := New(arg, "help", k, "", "", true)
			menu.Args = append(menu.Args, config)

		case IsArg(arg, []string{"-debug"}):
			config := New(arg, "debug", k, "", "", true)
			menu.Args = append(menu.Args, config)

		}
	}

	config := New(argv[len(argv)-1], "profile", len(argv)-1, argv[len(argv)-1], "", true)
	menu.Args = append(menu.Args, config)
}

func (menu *Menu) Start(argv []string) {
	//Extract the arguments
	menu.Parse(argv)

	//Verify that all arguments pass are valid
	for _, arg := range menu.Args {
		if !arg.Valid {
			ErrorAndExit(arg.Error, 1)
		}

		if arg.Key == "help" {
			menu.PrintOptions(argv)
		}
	}

	//Determine what mode to run.
	for _, arg := range menu.Args {
		if arg.Key == "make" {
			menu.MMake()
		}
	}

	menu.MRun()

}

func (menu *Menu) MMake() {
	var SetName bool
	var SetHost bool
	//Entered mode make from Start()
	//All args known to be valid

	/*
		Name        string
		Description string
		Username    string
		Hostname    string
		Port        uint16

		UseKey  bool
		KeyPath string

		SSHArgs []string
	*/
	NewConfig := NewConfigObject()

	for _, arg := range menu.Args {
		if arg.Key == "name" {
			SetName = true
			NewConfig.Name = arg.Value
		}
		if arg.Key == "desc" {
			NewConfig.Description = arg.Value
		}
		if arg.Key == "user" {
			NewConfig.Username = arg.Value
		}
		if arg.Key == "host" {
			SetHost = true
			NewConfig.Hostname = arg.Value
		}
		if arg.Key == "port" {
			port, err := strconv.Atoi(arg.Value)
			if err != nil {
				fmt.Println("Error when converting number for port")
				panic(err)
			}
			NewConfig.Port = uint16(port)
		}

		if arg.Key == "key" {
			NewConfig.UseKey = true
			NewConfig.KeyPath = arg.Value
		}

		if arg.Key == "option" {
			NewConfig.SSHArgs = append(NewConfig.SSHArgs, strings.Split(arg.Value, " ")...)
		}
	}

	for _, debug := range menu.Args {
		if debug.Key == "debug" {
			err := os.Stderr
			err.Write([]byte("I [DEBUG] SSHConfig {\n"))
			NewConfig._Debug_Print()
			err.Write([]byte("}\n"))
		}
	}

	if !SetHost {
		ErrorAndExit("No hostname is defined!", 1)
	}
	if !SetName {
		ErrorAndExit("Profile name is defined", 1)
	}

	NewConfig.Write(menu.GetConfig())
	os.Exit(0)

}

func (menu *Menu) MRun() {
	for _, arg := range menu.Args {
		if arg.Key == "profile" {
			menu.Config = Load(arg.Arg, menu.GetConfig())
			for _, debug := range menu.Args {
				if debug.Key == "debug" {
					err := os.Stderr
					err.Write([]byte("I [DEBUG] SSHConfig {\n"))
					menu.Config._Debug_Print()
					err.Write([]byte("}\n"))
				}
			}
			menu.StartSSH()
		}
	}
}

func (menu *Menu) StartSSH() {
	if menu.Config == nil {
		panic("PANIC: Empty config object")
	}

	if menu.Config.Username == "NULL" || menu.Config.Username == "" {
		user, err := user.Current()
		if err != nil {
			fmt.Println("Failed to get username")
			panic(err)
		}
		menu.Config.Username = user.Username
	}

	sshProcess := exec.Cmd{
		Path:   "/usr/bin/ssh",
		Args:   []string{},
		Env:    os.Environ(), // Hand over current shell connection
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	// /usr/bin/ssh
	sshProcess.Args = append(sshProcess.Args, sshProcess.Path)

	// -i KeyFile
	if menu.Config.UseKey {
		sshProcess.Args = append(sshProcess.Args, "-i")
		sshProcess.Args = append(sshProcess.Args, menu.Config.KeyPath)
	}

	// -o Args for ssh
	if menu.Config.SSHArgs != nil {
		sshProcess.Args = append(sshProcess.Args, menu.Config.SSHArgs...)
	}

	// -p Port
	sshProcess.Args = append(sshProcess.Args, "-p")
	sshProcess.Args = append(sshProcess.Args, strconv.Itoa(int(menu.Config.Port)))

	// User@Hostname:Port
	sshProcess.Args = append(sshProcess.Args, menu.Config.Username+"@"+menu.Config.Hostname)

	for _, debug := range menu.Args {
		if debug.Key == "debug" {
			err := os.Stderr
			err.Write([]byte("I [DEBUG] Exec Command:" + sshProcess.String() + "\n"))
		}
	}
	fmt.Println("Starting SSH...")
	err := sshProcess.Run()
	exit := sshProcess.ProcessState.ExitCode()
	if exit > 0 {
		if exit == 127 || exit == 130 { // Handle unclean SSH connection closure
			return
		}
		if err != nil {
			fmt.Println("Unknown error after attempting to run SSH")
			fmt.Println(sshProcess.String())
			panic(err)
		}
	}
}

func IsValidValue(arg string) bool {
	if arg[0] != '-' {
		return true
	} else {
		return false
	}
}

func (menu *Menu) PrintOptions(argv []string) {
	fmt.Println(argv[0] + ": [options] Profile")
	fmt.Println("    --help")
	fmt.Println("    	Displays this help text")
	fmt.Println("    -l, --list")
	fmt.Println("    	Lists avalible profiles to run in the current config directoy")
	fmt.Println("    -c, --config")
	fmt.Println("    	Sets the config directory. Default: $XDG_CONFIG_HOME else $HOME/.config")
	fmt.Println("    -m, --make")
	fmt.Println("    	Create a new profile to use")
	fmt.Println("    	-n, --name")
	fmt.Println("    		Profile Name, Required")
	fmt.Println("    	-h, --host")
	fmt.Println("    		Hostname for connection, Required")
	fmt.Println("    	-p, --port")
	fmt.Println("    		Port for connection")
	fmt.Println("    	-d, --desc")
	fmt.Println("    		Set the description field for the profile")
	fmt.Println("    	-u, --user")
	fmt.Println("    		Sets the username")
	fmt.Println("    	-k, --key")
	fmt.Println("    		Keyfile for connection")
	fmt.Println("    	-o, --option")
	fmt.Println("    		Addtional arguments to pass to ssh")

	for _, val := range menu.Args {
		if val.Key == "debug" {
			fmt.Println("    -debug")
			fmt.Println("    	Debugging flag to enable extra output. Usage of this flag is unstable and unsupported")
		}
	}
	os.Exit(0)
}
