package pcmsynth

import (
	"encoding/binary"
	"io"
)

/*func FindSmplToOffset(s io.ReadSeeker) (offset int64, size int64, err error) {
	_, err = s.Seek(12, io.SeekStart)
	if err != nil {
		return 0, 0, err
	}

	for {
		var chunkID [4]byte
		var chunkSize uint32

		// read chunk header
		if err = binary.Read(s, binary.LittleEndian, &chunkID); err != nil {
			return
		}
		if err = binary.Read(s, binary.LittleEndian, &chunkSize); err != nil {
			return
		}

		cur, _ := s.Seek(0, io.SeekCurrent)

		if string(chunkID[:]) == "smpl" {
			return int64(cur), int64(chunkSize), nil
		}

		// skip chunk (+ pad)
		skip := int64(chunkSize)
		if skip%2 == 1 {
			skip++
		}

		_, err = s.Seek(skip, io.SeekCurrent)
		if err != nil {
			return
		}
	}
}*/

func FindSmplToOffset(f io.ReadSeeker) (int64, int64, error) {
	// skip RIFF header
	_, err := f.Seek(12, io.SeekStart)
	if err != nil {
		return 0, 0, err
	}

	for {
		var id [4]byte
		var size uint32

		if err := binary.Read(f, binary.LittleEndian, &id); err != nil {
			return 0, 0, err
		}
		if err := binary.Read(f, binary.LittleEndian, &size); err != nil {
			return 0, 0, err
		}

		dataStart, _ := f.Seek(0, io.SeekCurrent)

		switch string(id[:]) {
		case "LIST":
			var listType [4]byte
			binary.Read(f, binary.LittleEndian, &listType)

			if string(listType[:]) == "sdta" {
				// dive into sdta
				end := dataStart + int64(size)
				for {
					pos, _ := f.Seek(0, io.SeekCurrent)
					if pos >= end {
						break
					}

					var cid [4]byte
					var csize uint32

					binary.Read(f, binary.LittleEndian, &cid)
					binary.Read(f, binary.LittleEndian, &csize)

					cstart, _ := f.Seek(0, io.SeekCurrent)

					if string(cid[:]) == "smpl" {
						return cstart, int64(csize), nil
					}

					skip := int64(csize)
					if skip%2 == 1 {
						skip++
					}

					f.Seek(skip, io.SeekCurrent)
				}
			} else {
				// skip LIST (size includes type, so subtract 4 already read)
				f.Seek(int64(size-4), io.SeekCurrent)
			}

		default:
			skip := int64(size)
			if skip%2 == 1 {
				skip++
			}
			f.Seek(skip, io.SeekCurrent)
		}
	}
}
