package playready

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/colde/psshutil/fileHandling"
	"log"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

type WRMHeader struct {
	XMLName xml.Name `xml:"WRMHEADER"`
	Version string   `xml:"version,attr"`
	Data    []Data   `xml:"DATA"`
}
type Data struct {
	XMLName     xml.Name      `xml:"DATA"`
	ProtectInfo []ProtectInfo `xml:"PROTECTINFO"`
	KeyID       string        `xml:"KID"`
	Checksum    string        `xml:"CHECKSUM"`
	LicenseUrl  string        `xml:"LA_URL"`
}
type ProtectInfo struct {
	XMLName     xml.Name `xml:"PROTECTINFO"`
	KeyLength   string   `xml:"KEYLEN"`
	AlgorithmID string   `xml:"ALGID"`
}

func Parse(f *os.File, size int64) {
	dataSize, err := fileHandling.ReadFromFile(f, 4)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	sizeInt := int64(binary.BigEndian.Uint32(dataSize))

	// Read PlayReady Header Length (identical to previous length, but little endian)
	_, err = fileHandling.ReadFromFile(f, 4)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	// Read record count
	_, err = fileHandling.ReadFromFile(f, 2)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	// Read rest of data
	buf, err := fileHandling.ReadFromFile(f, sizeInt-6)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	// Assume just 1 record and slice of the record type and record length
	header, err := DecodeUTF16(buf[4:])
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	xmlheader := WRMHeader{}
	err = xml.Unmarshal([]byte(header), &xmlheader)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	for _, data := range xmlheader.Data {
		key_bytes, _ := base64.StdEncoding.DecodeString(data.KeyID)

		tenc_bytes := make([]byte, 16)
		tenc_bytes[0] = key_bytes[3]
		tenc_bytes[1] = key_bytes[2]
		tenc_bytes[2] = key_bytes[1]
		tenc_bytes[3] = key_bytes[4]
		tenc_bytes[4] = key_bytes[3]
		tenc_bytes[5] = key_bytes[6]
		tenc_bytes[6] = key_bytes[5]
		copy(tenc_bytes[7:], key_bytes[7:])

		tenc_keyid := base64.StdEncoding.EncodeToString(tenc_bytes)

		fmt.Println("PlayReady Header/License KID:", data.KeyID)
		fmt.Println("PlayReady tenc KID:", tenc_keyid)
		fmt.Println("PlayReady LA_URL:", data.LicenseUrl)
	}
}

func DecodeUTF16(b []byte) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}
