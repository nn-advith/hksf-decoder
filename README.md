### Hollow Knight/Silksong save file decoder

All logic from : [bloodorca/hollow](https://github.com/bloodorca/hollow)

A golang implementation of the widely used save file editor. For no reason.

Single Save file decryption / Batch save file decryption.

Encryption procedure:

1. Stringify JSON
2. Encode using ECB-AES with 128 bit key (PKCS7 padding with 0x0B)
3. Encode using base64
4. Add length prefix: max 5 bytes with separator.
5. Add header and last byte (for some reason) 


Decryption procedure:

(Opposite of encryption)
