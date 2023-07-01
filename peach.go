// Copyright 2023 LankryF

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func info(text string) {
	println("[info]", text)
}
func errorAndExit(text string) {
	println("[error]", text)
	os.Exit(1)
}

func checkerr(err error, message string) {
	if err != nil {
		errorAndExit(message)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	errorAndExit("Fail to check path: " + path)
	return false
}

func selfPath() (path string) {
	ex, err := os.Executable()
	checkerr(err, "Can't get self path.")
	return filepath.Dir(ex)
}

func downloadFile(filepath, url string) error {
	out, err := os.Create(filepath)
	checkerr(err, "Can't create a destination file.")
	defer out.Close()

	resp, err := http.Get(url)
	checkerr(err, "Can't make an request. Check your internet connection.")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Download status: %s", resp.Status)
	}
	_, err = io.Copy(out, resp.Body)
	checkerr(err, "Can't write data to the file.")
	return err
}

func extractZip(zipPath, destPath string) {
	commandString := fmt.Sprintf("7z x %s -o%s", zipPath, destPath)
	commandSlice := strings.Fields(commandString)
	c := exec.Command(commandSlice[0], commandSlice[1:]...)
	checkerr(c.Run(), "Extracting error, maybe the archive is corrupted. Also check that you have 7zip on this machine.")
}

type Config struct {
	XamppPath             string
	PhpVersionsFolderPath string
}

func getPhpVersion(path string) string {
	out, err := ioutil.ReadFile(path + "\\php\\snapshot.txt")
	checkerr(err, "Failed to check a PHP version from snapshot.txt")
	return strings.Split(strings.Split(string(out), "Version: ")[1], "\n")[0]
}

func downloadPhpVersion(version, filepath string) {
	url := "https://sourceforge.net/projects/xampp/files/XAMPP%%20Windows/%s/xampp-portable-windows-x64-%s-0-%s.7z/download"
	for _, typ := range [...]string{"VS16", "VC15", "VC11"} {
		err := downloadFile(filepath, fmt.Sprintf(url, version, version, typ))
		if err == nil {
			break
		}
	}
}

func loadPhpVersion(conf Config, version string) {
	exists := pathExists(conf.PhpVersionsFolderPath + "\\" + version)
	if !exists {
		errorAndExit("Version " + version + " isn't found.")
	}
	exists = pathExists(conf.XamppPath)
	if !exists {
		errorAndExit("Invalid xampp path.")
	}
	savePhpVersion(conf, conf.XamppPath, false)
	err := os.Rename(conf.PhpVersionsFolderPath+"\\"+version+"\\php", conf.XamppPath+"\\php")
	checkerr(err, "Failed to load a PHP version to xampp.")
	info("Php loaded. PHP " + version)
	err = os.Rename(conf.PhpVersionsFolderPath+"\\"+version+"\\apache", conf.XamppPath+"\\apache")
	checkerr(err, "Failed to load an Apache version to xampp.")
	info("Apache loaded. PHP " + version)
	os.Remove(conf.PhpVersionsFolderPath + "\\" + version)
}

func savePhpVersion(conf Config, path string, phpinichange bool) {
	version := getPhpVersion(path)
	if phpinichange {
		formatPhpini(conf, path)
	}

	exists := pathExists(conf.PhpVersionsFolderPath + "\\" + version)
	if exists {
		info("Version " + version + " is already in the php_versions folder, it'll be overwritten.")
		err := os.RemoveAll(conf.PhpVersionsFolderPath + "\\" + version)
		checkerr(err, "Can't delete the excisting version "+version+" before download.")
	}

	err := os.Mkdir(conf.PhpVersionsFolderPath+"\\"+version, os.ModePerm)
	checkerr(err, "Can't make and folder for the version.")
	err = os.Rename(path+"\\php", conf.PhpVersionsFolderPath+"\\"+version+"\\php")
	checkerr(err, "Can't move PHP to the php_versions folder.")
	info("PHP saved. PHP " + version)
	err = os.Rename(path+"\\apache", conf.PhpVersionsFolderPath+"\\"+version+"\\apache")
	checkerr(err, "Can't move Apache to the php_versions folder.")
	info("Apache saved. PHP " + version)
}

func formatPhpini(conf Config, path string) {
	out, err := ioutil.ReadFile(path + "\\php\\php.ini")
	checkerr(err, "Failed to open the php.ini file.")
	text := string(out)
	text = strings.Replace(text, " \\xampp", " "+conf.XamppPath, -1)
	text = strings.Replace(text, "\"\\xampp", "\""+conf.XamppPath, -1)
	err = ioutil.WriteFile(path+"\\php\\php.ini", []byte(text), 0644)
	checkerr(err, "Failed to write the php.ini file.")
	info("php.ini formated.")
}

func printHelp() {
	fmt.Println(
		"Hi there, I'm Peach.\nMade by LankryF\n\npeach setup              <- create workplace !needed.\npeach xampp <path>       <- set xaamp folder path !needed.\npeach phps <path>        <- set php_versions folder (optional).\npeach list               <- list of your php versions.\npeach load <version>     <- load version (see peach list). Also saves current version.\npeach download <version> <- download version from the internet.\npeach info               <- get info.")
}

func (conf *Config) read() {
	file, err := os.Open(selfPath() + "\\config.json")
	defer file.Close()
	checkerr(err, "peach config.json configuration file is not found. Use \"setup\" command!")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	checkerr(err, "Failed to decode peach config file.")
}

func (conf *Config) write() {
	content, err := json.MarshalIndent(conf, "", "  ")
	checkerr(err, "Failed to generate a json file for config.")
	err = ioutil.WriteFile(selfPath()+"\\config.json", content, 0644)
	checkerr(err, "Failed to write config to the file config.json.")
}

func clearFolder(path string) {
	err := os.RemoveAll(path)
	checkerr(err, "Failed to delete a folder: "+path)
	err = os.Mkdir(path, os.ModePerm)
	checkerr(err, "Failed to make a folder: "+path)
}

func main() {

	selfp := selfPath()
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}
	args := os.Args[1:]
	conf := Config{}

	if args[0] != "setup" {
		conf.read()
	}

	switch args[0] {
	case "help":
		printHelp()
		os.Exit(0)

	case "info":
		configExists := pathExists(selfp + "\\config.json")
		phpsExists := pathExists(selfp + "\\php_versions")
		if configExists && phpsExists {
			info("Setuped: YES")
		} else {
			info("Setuped: NO")
		}
		if conf.PhpVersionsFolderPath == selfp+"\\php_versions" {
			info("Php versions folder: DEFAULT")
		} else {
			info("Php versions folder: " + conf.PhpVersionsFolderPath)
		}
		info("XAMPP folder path: " + conf.XamppPath)
		os.Exit(0)

	case "list":
		files, err := ioutil.ReadDir(conf.PhpVersionsFolderPath)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Current version:")
		version := getPhpVersion(conf.XamppPath)
		if err != nil {
			fmt.Println("	Invalid xampp path in config.")
		} else {
			fmt.Println("	" + version)
		}

		fmt.Println("Has been loaded:")
		for _, f := range files {
			if f.IsDir() {
				fmt.Println("	" + f.Name())
			}
		}

		fmt.Println("Popular versions for download:\n	8.2.0\n	8.1.12\n	8.0.25\n	7.4.33")
		os.Exit(0)

	case "setup":
		os.Mkdir(selfp+"\\php_versions", os.ModePerm)
		os.Mkdir(selfp+"\\temps", os.ModePerm)
		conf.PhpVersionsFolderPath = selfp + "\\php_versions"
		conf.XamppPath = "IS NOT SET"
		conf.write()
		os.Exit(0)
	}

	if len(args) != 2 {
		errorAndExit("There must be one argument: " + args[0] + " <argument>")
	}

	switch args[0] {
	case "download":
		clearFolder(selfp + "\\temps")
		info("Downloading...")
		downloadPhpVersion(args[1], selfp+"\\temps\\downloaded_version.7z")

		info("Extracting...")
		extractZip(selfp+"\\temps\\downloaded_version.7z", selfp+"\\temps")

		info("Saving version...")
		savePhpVersion(conf, selfp+"\\temps\\xampp", true)

		info("Clearing temps...")
		clearFolder(selfp + "\\temps")

		info("Done!")

	case "xampp":
		exists := pathExists(args[1])
		if !exists {
			errorAndExit("Invalid path.")
		}
		conf.XamppPath = args[1]
		conf.write()

	case "phps":
		exists := pathExists(args[1])
		if !exists {
			errorAndExit("Invalid path.")
		}
		conf.PhpVersionsFolderPath = args[1]
		conf.write()

	case "load":
		loadPhpVersion(conf, args[1])

	default:
		fmt.Println("Command " + args[0] + " is not found! see peach help.")
	}
	os.Exit(0)
}
