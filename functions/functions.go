package functions

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var csheader []int = []int{0, 1, 0, 0, 0, 255, 255, 255, 255, 1, 0, 0, 0, 0, 0, 0, 0, 6, 1, 0, 0, 0}
var aesKey []byte = []byte("UKu52ePUBwetZ9wNX88o54dnfKRu0T1l") // well known; from game files ig

// DECRYPTION RELATED

// remove pkcs7 padding; used in ECB aes
func removepadding(data []byte) []byte { // pkcs7 pads N places with byte of value N eg: 2 plcaes = 02, 11 places = 0B
	padLen := int(data[len(data)-1]) // gets the last value and converts to inte =~ gets the number of spaces padded
	return data[:len(data)-padLen]   // removes the number of spaces that have been
}

// ECB decryption basic
func aesDecrypt(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length must be multiple of %d", aes.BlockSize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// fmt.Println("before decrypt: md5:", md5.Sum(ciphertext))
	plaintext := make([]byte, len(ciphertext))
	for start := 0; start < len(ciphertext); start += aes.BlockSize { // iterate over the ciphertext, per blocak and decrypt
		block.Decrypt(plaintext[start:start+aes.BlockSize], ciphertext[start:start+aes.BlockSize])
	}
	// fmt.Println("before removing padding: md5: ", md5.Sum(plaintext))
	return removepadding(plaintext), nil
}

func DecryptSingleFile(infile string) error {

	splitpath := strings.Split(infile, "/")
	fname := splitpath[len(splitpath)-1]
	opname := fmt.Sprintf("%s-decoded.json", strings.Split(fname, ".")[0])

	data, err := os.ReadFile(infile)
	if err != nil {
		return fmt.Errorf("error : fileopen %v", err)
	}
	// fmt.Println("before length prefix: md5: ", md5.Sum(data))
	data = data[len(csheader) : len(data)-1] // remove the csheader; part of every save file and the last byte which is valeue 11

	// prefix := data[0:3]
	// for _, v := range prefix {
	// 	fmt.Printf("0x%X ", v)
	// }
	// fmt.Println()
	lc := 0
	for i := range 5 {
		lc++
		if (data[i] & 0x80) == 0 {
			break
		}
	}
	data = data[lc:]
	// fmt.Println("before b64 decode: md5:", md5.Sum([]byte(data)))
	decodeddata, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return fmt.Errorf("error : base64 decode %v", err)
	}

	plaintext, err := aesDecrypt(decodeddata, aesKey)
	if err != nil {
		return fmt.Errorf("error : decrypt %v", err)
	}

	// fmt.Println("plaintext: md5:", md5.Sum(plaintext))
	var formatted bytes.Buffer
	err = json.Indent(&formatted, []byte(plaintext), "", " ")
	if err != nil {
		return fmt.Errorf("error: formatting to json %v", err)
	}
	// fmt.Println("before write: md5:", md5.Sum(formatted.Bytes()))

	err = os.WriteFile("./DECODED/"+opname, formatted.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func DecryptDirectory(indir string) error {

	err := filepath.WalkDir(indir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("skipping: some error while traversing dir for %s: %v", path, err)
			return nil
		}
		if d.Type().IsRegular() {
			err := DecryptSingleFile(path)
			if err != nil {
				fmt.Printf("skipping: error decrtypting single file %s: %v", path, err)
				return nil
			}
		} else if d.Type().IsDir() && path != indir {
			// dont want subdirs to be checked
			return fs.SkipDir
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error while descrypting directory %v", err)
	}
	return nil
}

// ENCRYPTION RELATED

func addPadding(data []byte) []byte {
	// can cheat by directly assuming 0x0B padding based on knowledge but for fun
	// dynamic padding
	// fmt.Println("modulus: ", len(data)%aes.BlockSize)
	padLen := aes.BlockSize - (len(data) % aes.BlockSize)
	padSlice := bytes.Repeat([]byte{byte(padLen)}, padLen)
	data = append(data, padSlice...)
	return data
}

func aesEncrypt(plaintext, key []byte) ([]byte, error) {
	// add padding here
	plaintext = addPadding(plaintext)
	// fmt.Println("after padding: md5: ", md5.Sum(plaintext))

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	for start := 0; start < len(plaintext); start = start + aes.BlockSize {
		block.Encrypt(ciphertext[start:start+aes.BlockSize], plaintext[start:start+aes.BlockSize])
	}

	return []byte(ciphertext), nil
}

func appendLengthPrefix(data []byte) ([]byte, error) {
	// from length calculate the length prefix with separator
	// pseudo:
	// convert length: decimal to binary
	// if len(length)%7 != 0; then pad with difference at MSB
	// blocksize = 7
	// strating from end;
	// for each block:
	// 	if ! reach beginning; then MSB=1 followed by blockdata
	// 	else MSB = 0 followed by blockdata
	// endfor
	// convert to byte array
	length := len(data)
	blocksize := 7
	bin := strconv.FormatInt(int64(length), 2)
	if len(bin)%blocksize != 0 {
		diff := blocksize - len(bin)%blocksize
		pref := strings.Repeat("0", diff)
		bin = pref + bin
	}
	lprefix := make([]byte, 0)
	// fmt.Println(lprefix)
	for start := len(bin); start >= blocksize; start -= blocksize {
		temp := bin[start-blocksize : start]
		if start > blocksize {
			//not finished yet; add mSB

			temp = "1" + temp
		} else {
			//done add separator
			temp = "0" + temp

		}
		conv, err := strconv.ParseUint(temp, 2, 8)
		if err != nil {
			return []byte{}, fmt.Errorf("conversion error %v", err)
		}
		lprefix = append(lprefix, byte(conv))
	}

	// for _, v := range lprefix {
	// 	fmt.Printf("0x%X ", v)
	// }
	// fmt.Println()
	data = append(lprefix, data...)
	return data, nil
}

func inttobyte(insclice []int) []byte {
	outslice := make([]byte, len(insclice))
	for i, v := range insclice {
		outslice[i] = byte(v)
	}
	return outslice
}

func EncryptSingleFile(infile string) error {
	splitpath := strings.Split(infile, "/")
	fname := splitpath[len(splitpath)-1]
	opname := fmt.Sprintf("%s-encoded.dat", strings.Split(fname, "-decoded")[0])

	fdata, err := os.ReadFile(infile)
	if err != nil {
		return fmt.Errorf("error : fileopen %v", err)
	}

	// fmt.Println("after read: md5:", md5.Sum(fdata))
	var scrap bytes.Buffer
	err = json.Compact(&scrap, fdata)
	if err != nil {
		return err
	}
	// fmt.Println("plaintext: md5:", md5.Sum(scrap.Bytes()))
	ciphertext, err := aesEncrypt(scrap.Bytes(), aesKey)
	if err != nil {
		return fmt.Errorf("error : encrypt %v", err)
	}
	// fmt.Println("after encrypt: md5:", md5.Sum(ciphertext))

	encodeddata := base64.StdEncoding.EncodeToString(ciphertext)
	// fmt.Println("after b64 encode: md5:", md5.Sum([]byte(encodeddata)))
	prefixed, err := appendLengthPrefix([]byte(encodeddata))
	if err != nil {
		return fmt.Errorf("lengthprefix error: %v", err)
	}
	prefixed = append(inttobyte(csheader), append(prefixed, []byte{byte(11)}...)...)
	// fmt.Println("after length prefix: md5: ", md5.Sum(prefixed))

	err = os.WriteFile("./ENCODED/"+opname, prefixed, 0644)
	if err != nil {
		return err
	}
	return nil
}

func CheckFileEquality(files ...string) (bool, error) {
	equal := true
	md5file := func(fp string) (string, error) {
		f, err := os.Open(fp)
		if err != nil {
			return "", err
		}
		defer f.Close()
		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	}

	md5comp, err := md5file(files[0])
	if err != nil {
		return false, fmt.Errorf("error getting md5 %v", err)
	}

	for i := 1; i < len(files); i++ {
		temphash, err := md5file(files[i])
		if err != nil {
			return false, fmt.Errorf("error getting md5 %v", err)
		}
		if temphash != md5comp {
			equal = false
			break
		}

	}
	return equal, nil
}
