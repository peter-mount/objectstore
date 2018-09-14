package objectstore

import (
  "bytes"
  "fmt"
  "io"
  "strconv"
)

// dechunk takes the raw post and parses each chunk to form the original object.
// TODO add support to verify the chunks? Right now we just read them unverified.
func (s *ObjectStore) dechunk( payload []byte, dl string ) ([]byte, error) {
  el, err := strconv.Atoi( dl )
  if err != nil {
    return nil, err
  }

  // src & dest buffers
  src := bytes.NewBuffer( payload )
  dest := bytes.NewBuffer( nil )

  // Decode chunks until we get one of 0 length
  l := 1
  for l > 0 {
    l, err = s.decodechunk( src, dest )
    if err != nil {
      return nil, err
    }
  }

  if dest.Len() != el {
    return nil, fmt.Errorf( "Expected %d but got %d after parsing chunks", el, dest.Len() )
  }

  return dest.Bytes(), nil
}

func (s *ObjectStore) decodechunk( src, dest *bytes.Buffer ) (int, error) {

  // Remove cr/lf if any before the start
  c, err := src.ReadByte()
  if err != nil {
    return 0, err
  }
  for c=='\n' || c=='\r' {
    c, err = src.ReadByte()
    if err != nil {
      return 0, err
    }
  }
  src.UnreadByte()

  // Chunk header: length;checksum\n

  // Read the chunk size
  lenstr, err := src.ReadString( ';' )
  if err != nil {
    return 0, err
  }
  len, err := strconv.ParseInt( lenstr[:len(lenstr)-1], 16, 64 )
  if err != nil {
    return 0, err
  }

  // Read the checksum
  _, err = src.ReadString( '\n' )
  if err != nil {
    return 0, err
  }

  // Copy the chunk's content to the destination
  if len > 0 {
    wl, err := io.CopyN( dest, src, int64(len) )
    if err != nil {
      return 0, err
    }
    if wl != len {
      return 0, fmt.Errorf( "Wrote %d expected %d", wl, len )
    }
  }

  return int(len), nil
}
