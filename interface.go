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
	for k, arg := range argv {
		switch {
		case IsArg(arg, []string{"-m", "--make"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "make"
			config.Pos = k
			config.Valid = true
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-c", "--config"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "config"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-l", "--list"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "list"
			config.Pos = k
			config.Valid = true
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-h", "--host"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "host"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-u", "--user"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "user"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-p", "--port"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "port"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-n", "--name"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "name"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-d", "--desc"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "desc"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-k", "--key"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "key"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-o", "--option"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "option"
			config.Pos = k
			if k+1 < len(argv) {
				config.Value = argv[k+1]
				config.Valid = true
			} else {
				config.Valid = false
				config.Error = fmt.Sprintf("Unable to locate value for %s\n", arg)
			}
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"--help"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "help"
			config.Pos = k
			config.Valid = true
			menu.Args = append(menu.Args, config)
		case IsArg(arg, []string{"-debug"}):
			config := Arg{}
			config.Arg = arg
			config.Key = "debug"
			config.Pos = k
			config.Valid = true
			menu.Args = append(menu.Args, config)
		}
	}

	config := Arg{}
	config.Pos = len(argv) - 1
	config.Arg = argv[config.Pos]
	config.Key = "profile"
	config.Valid = true
	menu.Args = append(menu.Args, config)
}

func (menu *Menu) Start(argv []string) {
	//Extract the arguments
	menu.Parse(argv)

	//Verify that all arguments pass are valid
	for _, arg := range menu.Args {
		if !arg.Valid {
			fmt.Printf("%#v", arg)
			os.Exit(1)
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
		fmt.Println("Hostname not set!")
		os.Exit(1)
	}
	if !SetName {
		fmt.Println("Profile name not set!")
		os.Exit(1)
	}

	NewConfig.Write(menu.GetConfig())
	os.Exit(0)

}

func (menu *Menu) MRun() {
	for _, arg := range menu.Args {
		if arg.Key == "profile" {
			menu.Config = Load(arg.Arg, menu.GetConfig())
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
	if arg[0] == '-' {
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
