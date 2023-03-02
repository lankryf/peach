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

func checkerr(err error) {
	if err != nil {
		println("[error]", err.Error())
		os.Exit(1)
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func selfPath() (path string) {
	ex, err := os.Executable()
	checkerr(err)
	return filepath.Dir(ex)
}

func downloadFile(filepath, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Download status: %s", resp.Status)
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func extractZip(zipPath, destPath string) error {
	commandString := fmt.Sprintf("7z x %s -o%s", zipPath, destPath)
	commandSlice := strings.Fields(commandString)
	c := exec.Command(commandSlice[0], commandSlice[1:]...)
	return c.Run()
}

type Config struct {
	XamppPath             string
	PhpVersionsFolderPath string
}

func getPhpVersion(path string) (version string, err error) {
	out, err := ioutil.ReadFile(path + "\\php\\snapshot.txt")
	if err != nil {
		return "", err
	}
	return strings.Split(strings.Split(string(out), "Version: ")[1], "\n")[0], nil
}

func downloadPhpVersion(version, filepath string) error {
	url := "https://sourceforge.net/projects/xampp/files/XAMPP%%20Windows/%s/xampp-portable-windows-x64-%s-0-%s.7z/download"
	var err error

	for _, typ := range [...]string{"VS16", "VC15", "VC11"} {
		err = downloadFile(filepath, fmt.Sprintf(url, version, version, typ))
		if err == nil {
			break
		}
	}
	return err
}

func loadPhpVersion(conf Config, version string) {
	exists, err := pathExists(conf.PhpVersionsFolderPath + "\\" + version)

	checkerr(err)
	if !exists {
		fmt.Println("Version " + version + " isn't found.")
		os.Exit(1)
	}
	exists, err = pathExists(conf.XamppPath)
	checkerr(err)
	if !exists {
		fmt.Println("[error] Invalid xampp path.")
		os.Exit(1)
	}
	savePhpVersion(conf, conf.XamppPath, false)
	err = os.Rename(conf.PhpVersionsFolderPath+"\\"+version+"\\php", conf.XamppPath+"\\php")
	checkerr(err)
	info("Php loaded. PHP " + version)
	err = os.Rename(conf.PhpVersionsFolderPath+"\\"+version+"\\apache", conf.XamppPath+"\\apache")
	checkerr(err)
	info("Apache loaded. PHP " + version)
	os.Remove(conf.PhpVersionsFolderPath + "\\" + version)
}

func savePhpVersion(conf Config, path string, phpinichange bool) {
	version, err := getPhpVersion(path)
	if err != nil {
		fmt.Println("[error] No php in xaamp, can't save.")
		os.Exit(1)
	}
	if phpinichange {
		err = formatPhpini(conf, path)
		if err != nil {
			fmt.Println("[error] Format php.ini error:")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	exists, err := pathExists(conf.PhpVersionsFolderPath + "\\" + version)
	checkerr(err)
	if exists {
		fmt.Println("[info] Version " + version + " is already in php_versions folder, it'll be overwritten.")
		err = os.RemoveAll(conf.PhpVersionsFolderPath + "\\" + version)
		checkerr(err)
	}

	err = os.Mkdir(conf.PhpVersionsFolderPath+"\\"+version, os.ModePerm)
	checkerr(err)
	err = os.Rename(path+"\\php", conf.PhpVersionsFolderPath+"\\"+version+"\\php")
	checkerr(err)
	info("PHP saved. PHP " + version)
	err = os.Rename(path+"\\apache", conf.PhpVersionsFolderPath+"\\"+version+"\\apache")
	checkerr(err)
	info("Apache saved. PHP " + version)
}

func formatPhpini(conf Config, path string) (err error) {
	out, err := ioutil.ReadFile(path + "\\php\\php.ini")
	if err != nil {
		return err
	}
	text := string(out)
	text = strings.Replace(text, " \\xampp", " "+conf.XamppPath, -1)
	text = strings.Replace(text, "\"\\xampp", "\""+conf.XamppPath, -1)
	err = ioutil.WriteFile(path+"\\php\\php.ini", []byte(text), 0644)
	if err != nil {
		return err
	}
	info("php.ini formated.")
	return nil
}

func printHelp() {
	fmt.Println(
		"Hi there, I'm Peach.\nMade by LankryF\n\npeach setup              <- create workplace !needed.\npeach xampp <path>       <- set xaamp folder path !needed.\npeach phps <path>        <- set php_versions folder (optional).\npeach list               <- list of your php versions.\npeach load <version>     <- load version (see peach list). Also saves current version.\npeach download <version> <- download version from the internet.")
}

func (conf *Config) read() {
	file, err := os.Open(selfPath() + "\\config.json")
	defer file.Close()
	if err != nil {
		fmt.Println("[error] config.json configuration file is not found. Use \"setup\" command!")
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	checkerr(err)
}

func (conf *Config) write() {
	content, err := json.MarshalIndent(conf, "", "  ")
	checkerr(err)
	err = ioutil.WriteFile(selfPath()+"\\config.json", content, 0644)
	checkerr(err)
}

func clearFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	err = os.Mkdir(path, os.ModePerm)
	return err
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
		configExists, err := pathExists(selfp + "\\config.json")
		checkerr(err)
		phpsExists, err := pathExists(selfp + "\\php_versions")
		checkerr(err)
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
		version, err := getPhpVersion(conf.XamppPath)
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
		fmt.Println("[error] There must be one argument: " + args[0] + " <argument>")
		os.Exit(1)
	}

	switch args[0] {
	case "download":
		checkerr(clearFolder(selfp + "\\temps"))
		info("Downloading...")
		err := downloadPhpVersion(args[1], selfp+"\\temps\\downloaded_version.7z")
		checkerr(err)

		info("Extracting...")
		err = extractZip(selfp+"\\temps\\downloaded_version.7z", selfp+"\\temps")
		if err != nil {
			fmt.Println("[error] Extracting error, maybe the archive is corrupted. Also check that you have 7zip on this machine :(")
			os.Exit(1)
		}

		info("Saving version...")
		savePhpVersion(conf, selfp+"\\temps\\xampp", true)

		info("Clearing temps...")
		checkerr(clearFolder(selfp + "\\temps"))

		info("Done!")

	case "xampp":
		exists, err := pathExists(args[1])
		checkerr(err)
		if !exists {
			fmt.Println("[error] Invalid path.")
			os.Exit(1)
		}
		conf.XamppPath = args[1]
		conf.write()

	case "phps":
		exists, err := pathExists(args[1])
		checkerr(err)
		if !exists {
			fmt.Println("[error] Invalid path.")
			os.Exit(1)
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
