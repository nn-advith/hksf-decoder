package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nn-advith/hksfe/functions"
)

func createDirs() error {
	// create DECODED and ENCODED for output
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(wd, "DECODED"), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(wd, "ENCODED"), 0755)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	operation := os.Args[1]
	sfpath := os.Args[2]
	createDirs()
	if info, err := os.Stat(sfpath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("path not found")
		} else {
			fmt.Println("error while getting argument info ", err)
		}
		os.Exit(0)
	} else {
		fullname, err := filepath.Abs(sfpath)
		if err != nil {
			fmt.Println("error: getting abs path")
		}
		if strings.ToLower(operation) == "d" || strings.ToLower(operation) == "decode" {

			if info.IsDir() {
				fmt.Println("directory mode")
				err := functions.DecryptDirectory(fullname)
				if err != nil {
					fmt.Println("error decrypting directory")
				}
				fmt.Fprintf(os.Stdout, "\033[032mSUCCESS\033[0m: decoded files; Go get your GEO!\n")
			} else {
				//get file name and verify if it matches pattern
				// pattern := `^user[1-4]\.dat$`

				// re := regexp.MustCompile(pattern)
				// if re.MatchString(fname) {
				// 	fmt.Println("OK")
				// } else {
				// 	fmt.Println("not a valid file name; must match user[1-4].dat")
				// }

				err := functions.DecryptSingleFile(fullname)
				if err != nil {
					fmt.Println("error decrypting file", err)
				}

				fmt.Fprintf(os.Stdout, "\033[032mSUCCESS\033[0m: decoded into file; Go get your GEO!\n")

			}
		} else if strings.ToLower(operation) == "e" || strings.ToLower(operation) == "encode" {
			if info.IsDir() {
				fmt.Println("directory mode")
				// err := functions.E(fullname)
				// if err != nil {
				// 	fmt.Println("error decrypting directory")
				// }
				err := functions.EncryptDirectory(fullname)
				if err != nil {
					fmt.Println("error encryptinh directory: ", err)
				}
				fmt.Fprintf(os.Stdout, "\033[032mSUCCESS\033[0m: encoded files; Rename files to match user[1-4].dat to use them in game\n")
			} else {
				err := functions.EncryptSingleFile(fullname)
				if err != nil {
					fmt.Println("error decrypting file")
				}
				// equal, _ := functions.CheckFileEquality("./DAT/user4.dat", "./ENCODED/user4-encoded.dat")
				// fmt.Println(equal)
				fmt.Fprintf(os.Stdout, "\033[032mSUCCESS\033[0m: encoded into file; Rename file to match user[1-4].dat to use them in game\n")
			}
		}
	}

}
