package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

type Menu struct {
	ConfigDir   string
	MakeConfig  bool
	ListProfile bool
	PrintHelp   bool
}

func Start(argv []string) {
	menu := new(Menu)
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Failed to get config directory")
		panic(err)
	}

	// Set default config directory
	menu.ConfigDir = configDirectory + "/sshmanager"

	for arg := range argv {
		// Get the config directory from argv
		if argv[arg] == "-c" || argv[arg] == "--config" {
			if arg+1 < len(argv) {
				menu.ConfigDir = argv[arg+1]
			} else {
				fmt.Println("Missing Arg after -c!")
			}
		}
	}

	for arg := range argv {

		if argv[arg] == "-h" || argv[arg] == "--help" {
			menu.PrintHelp = true
		}

		//Generate a config from argv
		//Relies on -c
		if argv[arg] == "-m" || argv[arg] == "--make" {
			menu.MakeConfig = true
		}

		// List the avalible configs
		// Relies on -c
		if argv[arg] == "-l" || argv[arg] == "--list" {
			menu.ListProfile = true
		}

	}

	// -h, --help; conflicts with -m, --make as a sub-option is -h, --host
	if menu.PrintHelp && !menu.MakeConfig {
		PrintOptions(argv)
	}

	// -c, --config
	if !DoesPathExist(menu.ConfigDir) {
		fmt.Println("Config directory does not exsist!")
		fmt.Println("Config Dir:" + menu.ConfigDir)
	}

	// -l, --list
	if menu.ListProfile {
		ConfigDir, err := os.ReadDir(menu.ConfigDir)
		if err != nil {
			fmt.Println("Failed to get list of files")
			fmt.Println(menu.ConfigDir)
			panic(err)
		}

		for k, val := range ConfigDir {
			name := val.Name()
			name = name[:len(name)-5]
			fmt.Println("[" + strconv.Itoa(k) + "] " + name)
		}
	}

	// -m, --make
	if menu.MakeConfig {
		MakeConfig(argv, menu)
	}

	config := Load(argv[len(argv)-1], menu.ConfigDir)
	StartSSH(config, argv)

}

func MakeConfig(argv []string, menu *Menu) {
	// Name        Required!
	// Description Optional; Blank
	// Username    Optional; Current
	// Hostname    Required!
	// Port        Optional; Earth default is 22

	// UseKey
	// KeyPath Optional; Sets UseKey if used

	// SSHArgs Optional; Injects addtional arguments into ssh

	// sshmanager -m -n Localhost -d "Current System" -u root -p 22 -k "~/.ssh/HandsOff"
	setName := false
	setHostname := false

	newConfig := NewConfigObject()

	for iter := range argv {
		if argv[iter] == "-n" || argv[iter] == "--name" {
			if (iter+1 < len(argv)) && !IsValidOpt(argv[iter+1]) {
				setName = true
				newConfig.Name = argv[iter+1]
			} else {
				fmt.Println("Missing name after  " + argv[iter] + "!")
				os.Exit(1)
			}
		}
		if argv[iter] == "-d" || argv[iter] == "--desc" || argv[iter] == "--description" {
			if (iter+1 < len(argv)) && !IsValidOpt(argv[iter+1]) {
				newConfig.Description = argv[iter+1]
			} else {
				fmt.Println("Missing desciption after " + argv[iter] + "!")
				os.Exit(1)
			}
		}
		if argv[iter] == "-u" || argv[iter] == "--user" || argv[iter] == "--username" {
			if (iter+1 < len(argv)) && !IsValidOpt(argv[iter+1]) {
				newConfig.Username = argv[iter+1]
			} else {
				fmt.Println("Missing username after " + argv[iter] + "!")
				os.Exit(1)
			}
		}
		if argv[iter] == "-h" || argv[iter] == "--host" || argv[iter] == "--hostname" {
			if (iter+1 < len(argv)) && !IsValidOpt(argv[iter+1]) {
				setHostname = true
				newConfig.Hostname = argv[iter+1]
			} else {
				fmt.Println("Missing hostname after " + argv[iter] + "!")
				os.Exit(1)
			}
		}
		if argv[iter] == "-p" || argv[iter] == "--port" {
			if (iter+1 < len(argv)) && !IsValidOpt(argv[iter+1]) {
				port, err := strconv.Atoi(argv[iter+1])
				if err != nil {
					fmt.Println("Failed to convert String to int")
					panic(err)
				}
				newConfig.Port = uint16(port)
			} else {
				fmt.Println("Missing port after " + argv[iter] + "!")
				os.Exit(1)
			}
		}
		if argv[iter] == "-k" || argv[iter] == "--key" || argv[iter] == "--keyfile" {
			if (iter+1 < len(argv)) && !IsValidOpt(argv[iter+1]) {
				newConfig.UseKey = true
				newConfig.KeyPath = argv[iter+1]
			} else {
				fmt.Println("Missing key after " + argv[iter] + "!")
				os.Exit(1)
			}
		}

		if argv[iter] == "-o" || argv[iter] == "--option" {
			if iter+1 < len(argv) {
				newConfig.SSHArgs = strings.Split(argv[iter+1], " ")
			} else {
				fmt.Println("Missing Quoted SSH Arguments to pass after " + argv[iter] + "!")
				os.Exit(1)
			}
		}
	}

	if !setName {
		fmt.Println("Missing name! Please specify one using -n")
		os.Exit(1)
	}
	if !setHostname {
		fmt.Println("Missing hostname! Please specify one using -h")
		os.Exit(1)
	}

	if !DoesPathExist(menu.ConfigDir) {
		err := os.MkdirAll(menu.ConfigDir, 0755)
		if err != nil {
			fmt.Println("Config folder not found. Failed to generate the correct path for it")
			panic(err)
		}
	}
	newConfig.Write(menu.ConfigDir)
	os.Exit(0)
}

func StartSSH(config *ConfigObject, argv []string) {
	if config.Username == "NULL" || config.Username == "" {
		user, err := user.Current()
		if err != nil {
			fmt.Println("Failed to get username")
			panic(err)
		}
		config.Username = user.Username
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
	if config.UseKey {
		sshProcess.Args = append(sshProcess.Args, "-i")
		sshProcess.Args = append(sshProcess.Args, config.KeyPath)
	}

	// -o Args for ssh
	if config.SSHArgs != nil {
		sshProcess.Args = append(sshProcess.Args, config.SSHArgs...)
	}

	// -p Port
	sshProcess.Args = append(sshProcess.Args, "-p")
	sshProcess.Args = append(sshProcess.Args, strconv.Itoa(int(config.Port)))

	// User@Hostname:Port
	sshProcess.Args = append(sshProcess.Args, config.Username+"@"+config.Hostname)

	for _, debug := range argv {
		if debug == "-debug-sshcmd" {
			fmt.Println(sshProcess.String())
			os.Exit(0)
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

func IsValidOpt(arg string) bool {
	if arg[0] == '-' {
		return true
	} else {
		return false
	}
}

func PrintOptions(argv []string) {
	fmt.Println(argv[0] + ": [options] Profile")
	fmt.Println("    -h, --help")
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

	for _, val := range argv {
		if val == "-debug" {
			fmt.Println("    -d, --debug")
			fmt.Println("    	Print debug commands; NOT SUPPORTED!")
			fmt.Println("    -debug-sshcmd")
			fmt.Println("    	Does not execute SSH but instead prints out the command it would run")
		}
	}
	os.Exit(0)
}
