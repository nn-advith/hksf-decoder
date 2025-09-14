package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nn-advith/hksfe/functions"
)

func main() {

	sfpath := os.Args[1]
	fmt.Println(sfpath)
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
		if info.IsDir() {
			fmt.Println("directory")
			err := functions.DecryptDirectory(fullname)
			if err != nil {
				fmt.Println("error decrypting directory")
			}
		} else {
			//get file name and verify if it matches pattern
			// pattern := `^user[1-4]\.dat$`

			// re := regexp.MustCompile(pattern)
			// if re.MatchString(fname) {
			// 	fmt.Println("OK")
			// } else {
			// 	fmt.Println("not a valid file name; must match user[1-4].dat")
			// }

			err = functions.DecryptSingleFile(fullname)
			if err != nil {
				fmt.Println("error decrypting file")
			}

			fmt.Fprintf(os.Stdout, "\033[032mSUCCESS\033[0m: decrypyted into file\n")

		}
	}

}
