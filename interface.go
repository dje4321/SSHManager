package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
)

/*
Args

-c Config Directory
*/

type Menu struct {
	ConfigDir string
	Username  string
	IP        string
	Port      uint16
}

func Start(argv []string) {
	menu := new(Menu)
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Failed to get config directory")
		panic(err)
	}

	// BUG Verify config directory exsists
	menu.ConfigDir = configDirectory + "/sshmanager"

	for arg := range argv {

		if argv[arg] == "-h" || argv[arg] == "--help" {
			PrintHelp(argv)
		}

		// Get the config directory from argv
		if (argv[arg] == "-c" || argv[arg] == "--config") && !IsValidOpt(argv[arg+1]) {
			if arg+1 < len(argv) {
				// BUG Verify config directory exsists
				menu.ConfigDir = argv[arg+1]
			} else {
				fmt.Println("Missing Arg after -c!")
			}
		}

		//Generate a config from argv
		//Relies on -c
		if argv[arg] == "-m" || argv[arg] == "--make" {
			// Name        Required!
			// Description Optional; Blank
			// Username    Optional; Current
			// Hostname    Required!
			// Port        Optional; Earth default is 22

			// UseKey
			// KeyPath Optional; Sets UseKey if used

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
						newConfig.SSHArgs = argv[iter+1]
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

			newConfig.Write(menu.ConfigDir)
			os.Exit(0)
		}

		// List the avalible configs
		// Relies on -c
		if argv[arg] == "-l" || argv[arg] == "--list" {
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

			return
		}

	}

	config := Load(argv[len(argv)-1], menu.ConfigDir)
	StartSSH(config, argv)

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
	if config.SSHArgs != "" {
		sshProcess.Args = append(sshProcess.Args, config.SSHArgs)
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

func PrintHelp(argv []string) {
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
	os.Exit(0)
}
