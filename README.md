### Hollow Knight/Silksong save file decoder

All logic from : [bloodorca/hollow](https://github.com/bloodorca/hollow)

A golang implementation of the widely used save file editor. For no reason.

Single Save file decryption / Batch save file decryption.

##### Encryption procedure:

1. Read JSON file and compact stringify
2. Encode using ECB-AES with 128 bit key (PKCS7 padding)
3. Encode using base64
4. Add length prefix: calculate
5. Add header and last byte (for some reason) 
6. Write to DAT file


##### Decryption procedure:

1. Read DAT file
2. Remove header and last byte
3. Remove length prefix: iterate till you reach byte with low MSB i.e last byte of length prefix
4. Decode using base64
5. Decode using ECB-AES with 128 bit key
6. Indent JSON to make it pretty.
7. Write to JSON file.

---

##### Example:

- Decode single save file
```
go run main.go d ./DAT/user4.dat
ls ./DECODED/
```
- Batch decode save files
```
go run main.go d ./DAT/
ls ./DECODED/
```
- Encode single save file
```
go run main.go d ./DECODED/user4-decoded.json
ls ./ENCODED/
```
- Batch encode save files
```
go run main.go d ./DECODED/
ls ./ENCODED/
```

After decode; the json files will be saved under ./DECODED. Modify the parameters here. Then encode the modified JSON files. The encoded files will be saved under ./ENCODED. Rename the encoded DAT files to match format ```user[1-4].dat``` and copy to the savefile directory of HK/HK:SS.

(**WARNING**: Too bored to test in depth so make a backup of save files.)

Shaw!
