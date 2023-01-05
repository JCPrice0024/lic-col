# GO REPORT
[![Go Report](https://goreportcard.com/badge/github.com/JCPrice0024/lic-col)](https://goreportcard.com/badge/github.com/JCPrice0024/lic-col)

# lic-col
lic-col or License Collector is an opensource program that finds and collects all licenses in a program's sub-dependencies. The motivation behind this project was to quickly and efficiently copy all License files in your sub-dependencies to a single folder that you can place in your project. It is important to get all of the licenses of your sub-dependencies because when you import code that is not your's you MUST also import the license that is used to not be subject to legal problems (This is a condition in most License files). Another motivation of this program was also to help make sure your program isn't importing any licenses that could harm your code, (like a copy-left license in a closed source project).


# HOW TO INSTALL IT
mkdir -p $GOPATH/src/github.com/JCPrice0024

cd $GOPATH/src/github.com/JCPrice0024

git clone git@github.com:JCPrice0024/lic-col.git or git clone https://github.com/JCPrice0024/lic-col.git

cd lic-col/licenseCol/

go install


# HOW IT WORKS

lic-col works by performing a series of filepath.Walks on a repos go.sum file. In its simplest you give the program the repo name you want scanned and it performs a git clone on that repo, it then walks through the repo looking for go.sum files. When it finds one it runs a go mod download and then reads the go.sum file. Next it passes the info read to a function called ScanPath which takes the go sum information and does a strings.split on all new lines. Ranging over that information it formats all of the sub-dependencies in the same way as the go mod download to allow them to be scanned. It then performs a filepath.walk on the path that was just made and searches for License files. Finally it copies those files into the dst described below in addition there are some special configuration options and flags which will be described below. It also (if available) adds a link to the current github repo


# EXAMPLE
To generate html output and to get git's License guess use the following format when running the program:
licenseCol -repo="https://github.com/JCPrice0024/lic-testRepo5.git" -dst="c:/AllLicenses" -tohtml -git-check

To simply create directory with all licenses use the following format:
licenseCol -repo="https://github.com/JCPrice0024/lic-testRepo5.git" -dst="c:/AllLicenses"

# HOW TO USE

So far lic-col has 7 command line args, They are shown below:

![image](https://user-images.githubusercontent.com/111247018/209986038-e82555a2-ddc8-490c-aad2-532133aa87c6.png)

When you run the program you NEED to use at least the repo and dst flags.

-repo
The repo flag is a valid git clone link it can be of https or ssh and examples are shown in the flag description.

-dst
The dst flag is simply a path to the desired location of the License folder, it DOES NOT need to be a premade path as the program will make the necessary directories for you. Once the programmakes the path you entered it will add a few more folders in it for eassier organization. The top layer folder will be reponame_Licenses then inside that it will have a folder called Licenses that holds all of the copied licenses (in html format if specified in he Command Line Args). There will always be json file in that folder that holds the name, path, github repo link (if able), and the github license (if able) of all scanned licenses, it also holds what type of license they were and is formatted in map[string]struct

-version
The version flag allows you to specify which version of the repo you want to scan. To use this flag you need to get the hash of the commit you want and the program will perform a git checkout on that hash. If you don't specify a version lic-col scans the current version.

-tohtml
The tohtml flag is a boolean that declares whether you want all copied files to be in html, if you specify this the reponame_Licenses(described in the dst path) will have another file inside called index.html. This file organizes all of the copied html files and allows for a friendlier/easier to read output. 

-clean-mod
When launched the program preforms a go mod download on all mod files. This can eat up space so the clean-mod flag will erase all downloaded folders, it will ONLY erase NEW folders so if you run the program twice on the same repo and DON'T use the clean-mod flag the first time, it will clean nothing the second time.

-clean-clone
As well as performing a go-mod download the program will also if necessary perform a git clone, if you want to remove the clone once the program exits the clean-clone flag will perform an os.RemoveAll on it. This will erase the ENTIRE repo so use it only if that is the desired result.

-git-check
The git-check flag adds another layer of information to your scanned licenses. It is a boolean and if you mark it as true the program will ask you for your github Username and it will ask you for a github Personal Access Token (Here is a link showing how to get one: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token). It will then make requests to the github api to get the CURRENT Sub-dependency's repo's license. This license from github may be different from what the results of the scan say, this could be for a number of reasons but mainly it has to do with version differences. It is there for you to validate and check if you desire more information. IMPORTANT NOTE: Your personal access token is allowed about 5000 requests per hour the program is designed to stop sending requests if the number of remaining requests goes below 400 this is to prevent locking your token. If you clone the repo DO NOT REMOVE THE SAFETY MEASURE. 

In addition to those flags there are a few configuration files to help customize your results here is a list of the current config files and how to use them:

definedlicenses.json
This config file lists all defined licenses, this will allow the results to be labeled based on what type of license the file is. This is a map[string]string json file so to configure it you need the license name (Apache 2.0) and license definition (Apache License Version 2.0). lic-col comes with some pre-configured config files including this one simply clone the repo and you will have access to the file and you'll be able to edit it as well. However if you'd like to make your own definedlicenses.json file you can declare it with the environment variable DES_LIC .

excludedfiles.json
This config file lists all excluded files, this will allow you to enter in exact files or file names that you don't want to be scanned. This is a map[string]emptystruct{} the string is the file name or path. If you want to remove only one file use the entire path if you want to remove all files that have that name just enter the name. lic-col comes with some pre-configured config files including this one, however this file is empty if you want to change it simply clone the repo and you will have access to the file. If you'd like to make your own excludedfiles.json file you can declare it with the environment variable DES_EXCL.

excludedextensions.json
This config file lists all excluded file extenstions, this allows you to remove all file types of a certain extension. It is also a map[string]emptystruct{} the string is the .fileext you want to remove. This is another one of the pre-configured config files it blocks most programming language file extensions like before if you want to use it just clone the program and if you want to configure your own use the environment variable DES_EXT.

includedfiles.json
This config file lists all included files, this is the opposite of excludedfiles.json as it allows you to included files that would normally be skipped. The map and rules are the same just with the opposite effects. This is another one of the pre-configured config files it includes a different type of license file. Like before if you want to use it just clone the program and if you want to configure your own use the environment variable DES_INCL.

overridelicense.json
This config file allows you to override a path, this is in the case that a specific file in a sub-dependency is a license file but it is NOT a standard license file. You need to get the sub-dependency path, the file name and the license type you want to declare it as, this is a map[string]struct{Filename: License:} the string is the sub-dependency path. If the filename is not in the top level of the sub-dependency you need to specify the path to it for example dir/filename. Here is an example of how to set up the override: 

"github.com/owner/reponame@v0.0.0-20201107003712-816f3ae12d81": {
      "License": "Apache",
      "Filename": "dir/filename"
   }

There is also an example in the base config files. To create your own override like before you can edit the file after you clone the program or you can use the environment variable DES_OVER

cache.json
This config file is special. This one is not pre-configured but is made after the program is lauched. It is only made if you use the git-check command line arg. It holds all requested license names for a project. This is so that if you run the program multiple times you won't have to spam the githubapi as the info will be stored here. 

# SONAR RESULTS 
![image](https://user-images.githubusercontent.com/111247018/210660570-069e6dc3-bbab-4681-a162-31f3a8e18547.png)


# IMPORTANT NOTE
This program is not full proof nor does it claim to be. It has been tested and it works but there is always room for error and thus should NOT be considered legal advice.
