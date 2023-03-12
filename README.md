<p align="center"><img src="https://raw.githubusercontent.com/lankryf/gallery/master/peach.png" width="400" alt="Laravel Logo"></p>

## Introducing
Hi there, I'm Peach, a console program created by LankryF to simplify the process of managing PHP versions for XAMPP.

## Installation
1. To use Peach, simply [download](https://sourceforge.net/projects/lankryf-peach/files/latest/download) it onto your local machine.
2. Move peach.exe file to any folder of your choice, but **it must be on the common disk with XAMPP**.
3. Add peach folder path to environment settings.
4. And finaly you must to setup peach and set XAMPP folder path by using:
```
peach setup
peach xampp path/to/your/xampp
```
_hint: Replace 'path/to/your/xampp' with the path of your XAMPP folder._

You can check info to make sure that instalation has been done corectly.
```
peach info
```
After corect instalation output must be:
```
[info] Setuped: YES
[info] Php versions folder: DEFAULT
[info] XAMPP folder path: path/to/your/xampp
```

## Downloading versions
To download PHP version simply use:
```
peach download 8.0.25
```
_hint: as 8.0.25 you can write version you want._
Wait for downloading and you can use this version.

## What versions do you have?
You can check PHP versions by using:
```
peach list
```
For example output can be:
```
Current version:
        8.2.0
Has been loaded:
        8.0.25
Popular versions for download:
        8.2.0
        8.1.12
        8.0.25
        7.4.33
```

## Loading version to XAMPP
To load version from downloads you can use:
```
peach load 8.0.25
```
_hint: as 8.0.25 you can write version you have._
Current version will be saved to PHP versions folder.

## Customisation
Peach has PHP versions folder, used as storage of saved versions. By default it apears in the same folder with Peach. You can chenge versions folder path by using:
```
peach phps path/to/php/versions/folder
```
_hint:  path/to/php/versions/folder must be replaced to your custom PHP versions folder path._

Also you can chage XAMPP folder by using:
```
peach setup
peach xampp path/to/your/xampp
```
_hint: Replace 'path/to/your/xampp' with the path of your XAMPP folder._

**Notice that XAAMP path and PHP versions folder path must be on the common disk with Peach.**

## Manual
Peach has a small manual, which you can access by using:
```
peach help
```
or just:
```
peach
```
## Mention
Thanks, Black Sirius, for this sweet logo :)