package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func isCommandAvailable(command string) bool {
	cmd := exec.Command(command, "-v")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func getAppNameFromArgs(args []string) string {
	var appName string

	if args[0] != "" {
		appName = args[0]
	} else {
		appName = "poa-app"
	}

	return appName
}

func writeStructToFile(content interface{}, appName string, filename string) {
	packageJSON, _ := json.Marshal(content)
	packageJSONPath := appName + "/" + filename

	err := ioutil.WriteFile(packageJSONPath, packageJSON, 0644)

	if err != nil {
		fmt.Println("Error creating package.json")
		os.Exit(1)
	}
}

func executeShell(command *exec.Cmd) bool {
	shellCommand := command
	var shellCommandOutBuffer, shellCommandErrorBuffer bytes.Buffer

	shellCommand.Stdout = &shellCommandOutBuffer
	shellCommand.Stderr = &shellCommandErrorBuffer

	if err := shellCommand.Run(); err != nil {
		fmt.Println("Command failure")
		fmt.Println(err)
		fmt.Println("out:", shellCommandOutBuffer.String(), "err:", shellCommandErrorBuffer.String())

		return false
	}

	return true
}

type PackageJSONScripts struct {
	Start  string `json:"start"`
	Build  string `json:"build"`
	Test   string `json:"test"`
	Eject  string `json:"eject"`
	Format string `json:"format"`
}

type PackageJSON struct {
	Name    string             `json:"name"`
	Version string             `json:"version"`
	Private bool               `json:"private"`
	Scripts PackageJSONScripts `json:"scripts"`
}

type PrettierJSON struct {
	PrintWidth  int  `json:"printWidth"`
	SingleQuote bool `json:"singleQuote"`
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create new application",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		var useYarn bool = isCommandAvailable("yarn")
		var appName string = getAppNameFromArgs(args)

		fmt.Println("isYarn=", useYarn)
		fmt.Println("appName=", appName)

		// create app directory and source folder
		os.MkdirAll(appName+"/src", os.ModePerm)
		os.MkdirAll(appName+"/public", os.ModePerm)

		// create package json file
		scripts := PackageJSONScripts{"react-scripts start", "react-scripts start build", "react-scripts test --env=jsdom", "react-scripts eject", `prettier --write *.{js,css,json,md} '**/*.{js,css,json,md}'`}
		packageContent := PackageJSON{appName, "0.0.0", true, scripts}

		writeStructToFile(packageContent, appName, "package.json")

		// install dependencies
		devDependencies := []string{"react-scripts", "prettier"}
		dependencies := []string{"poa", "react", "react-dom"}
		yarnCommonArgs := []string{"--no-progress", "--dev", "--non-interactive", "--cwd", appName}
		yarnInstallArgs := append([]string{"add"}, yarnCommonArgs...)

		fmt.Println("Dependencies installation")

		var isInstallationWasOk bool = true

		if useYarn {
			// dev deps
			devArgs := append(yarnInstallArgs, "--dev")
			if !executeShell(exec.Command("yarn", append(devArgs, devDependencies...)...)) {
				isInstallationWasOk = false
			}

			// deps
			args := append(yarnInstallArgs)
			if !executeShell(exec.Command("yarn", append(args, dependencies...)...)) {
				isInstallationWasOk = false
			}

		} else {
			// dev deps
			devArgs := append(devDependencies, "--dev", "--save")
			installDevCommand := exec.Command("npm install", devArgs...)
			installDevCommand.Dir = appName

			if !executeShell(installDevCommand) {
				isInstallationWasOk = false
			}

			// deps
			args := append(devDependencies, "--save")
			installCommand := exec.Command("npm install", args...)
			installCommand.Dir = appName

			if !executeShell(installCommand) {
				isInstallationWasOk = false
			}
		}

		if isInstallationWasOk {
			fmt.Println("Dependencies installation OK")
		} else {
			os.Exit(1)
		}

		// create prettier json file
		prettierContent := PrettierJSON{100, true}

		writeStructToFile(prettierContent, appName, ".prettierrc.json")

		fmt.Println("Fomatting codebase")

		var formatCommand *exec.Cmd

		if useYarn {
			formatCommand = exec.Command("yarn", "format")
			formatCommand.Dir = appName

		} else {
			formatCommand = exec.Command("npm", "run", "format")
			formatCommand.Dir = appName
		}

		if executeShell(formatCommand) {
			fmt.Println("Fomatting codebase OK")
		} else {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
