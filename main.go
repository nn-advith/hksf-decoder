package main

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var csheader []int = []int{0, 1, 0, 0, 0, 255, 255, 255, 255, 1, 0, 0, 0, 0, 0, 0, 0, 6, 1, 0, 0, 0}
var aesKey []byte = []byte("UKu52ePUBwetZ9wNX88o54dnfKRu0T1l") // well known; from game files ig

// remove pkcs7 padding; used in ECB aes
func removepadding(data []byte) []byte { // pkcs7 pads N places with byte of value N eg: 2 plcaes = 02, 11 places = 0B
	padLen := int(data[len(data)-1]) // gets the last value and converts to inte =~ gets the number of spaces padded
	return data[:len(data)-padLen]   // removes the number of spaces that have been
}

// ECB decryption basic
func aesDecrypt(ciphertext, key []byte) ([]byte, error) {
	fmt.Println(aes.BlockSize)
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length must be multiple of %d", aes.BlockSize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	for start := 0; start < len(ciphertext); start += aes.BlockSize { // iterate over the ciphertext, per blocak and decrypt
		block.Decrypt(plaintext[start:start+aes.BlockSize], ciphertext[start:start+aes.BlockSize])
	}
	return removepadding(plaintext), nil
}

func decryptSingleFile(infile string, opname string) error {
	fmt.Println("testing")
	data, err := os.ReadFile(infile)
	if err != nil {
		return fmt.Errorf("error : fileopen %v", err)
	}
	data = data[len(csheader) : len(data)-1] // remove the csheader; part of every save file and the last byte which is valeue 11

	lc := 0
	for i := range 5 {
		lc++
		if (data[i] & 0x80) == 0 {
			break
		}
	}

	data = data[lc:]

	decodeddata, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return fmt.Errorf("error : base64 decode %v", err)
	}
	// fmt.Println(string(decodeddata))

	plaintext, err := aesDecrypt(decodeddata, aesKey)
	if err != nil {
		return fmt.Errorf("error : decrypt %v", err)
	}

	var formatted bytes.Buffer
	err = json.Indent(&formatted, []byte(plaintext), "", " ")
	if err != nil {
		return fmt.Errorf("error: formatting to json %v", err)
	}
	err = os.WriteFile("./OP/"+opname, formatted.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

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

		if info.IsDir() {
			fmt.Println("directory")
		} else {
			//get file name and verify if it matches pattern
			// pattern := `^user[1-4]\.dat$`
			splitpath := strings.Split(sfpath, "/")
			fname := splitpath[len(splitpath)-1]
			// re := regexp.MustCompile(pattern)
			// if re.MatchString(fname) {
			// 	fmt.Println("OK")
			// } else {
			// 	fmt.Println("not a valid file name; must match user[1-4].dat")
			// }

			fullname, err := filepath.Abs(sfpath)
			if err != nil {
				fmt.Println("error: getting abs path")
			}
			err = decryptSingleFile(fullname, fmt.Sprintf("%s-decoded.json", strings.Split(fname, ".")[0]))
			if err != nil {
				fmt.Println("error decrypting file")
			}

			fmt.Fprintf(os.Stdout, "\033[032mSUCCESS\033[0m: decrypyted into file\n")

		}
	}

}
