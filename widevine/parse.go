package widevine

import (
	"os"
  "log"
  "encoding/binary"
  "github.com/golang/protobuf/proto"
  "github.com/colde/psshutil/fileHandling"
)

func Parse(f *os.File, size int64) {
  dataSize, err := fileHandling.ReadFromFile(f, 4)
  if err != nil {
    log.Fatalln(err.Error())
    return
  }

  sizeInt := int64(binary.BigEndian.Uint32(dataSize))

  buf, err := fileHandling.ReadFromFile(f, sizeInt)
  if err != nil {
    log.Fatalln(err.Error())
    return
  }

  widevineHeader := &WidevinePsshData{}
  err = proto.Unmarshal(buf, widevineHeader)
  if err != nil {
      log.Fatal("unmarshaling error: ", err)
  }

  log.Println("Widevine Content Id", string(widevineHeader.GetContentId()))
  log.Println("Widevine provider Id", string(widevineHeader.GetProvider()))
}
