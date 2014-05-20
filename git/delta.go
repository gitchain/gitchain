package git

const delta_size_min = 4

func deltaHeaderSize(datap *[]byte, sz *uint) uint {
	data := *datap
	var size, i, j uint
	var cmd byte
	for {
		data = data[j:]
		cmd = data[0]
		size |= (uint(cmd) & 0x7f) << i
		if !(uint(cmd)&0x80 != 0 && j < *sz) {
			break
		}
		i += 7
		j++
	}
	datap = &data
	newSz := *sz - j
	sz = &newSz
	return size
}

func PatchDelta(src, delta []byte) []byte {
	if len(delta) < delta_size_min {
		return nil
	}

	top := uint(len(delta))

	size := deltaHeaderSize(&delta, &top)
	if size != uint(len(src)) {
		return nil
	}

	size = deltaHeaderSize(&delta, &top)
	dest := make([]byte, size)

	var data []byte
	copy(data, src)

	var offset uint
	var cmd byte
	for {
		data = src[offset:]
		cmd = data[0]
		if (cmd & 0x80) != 0 {
			var cp_off, cp_size uint
			if (cmd & 0x01) != 0 {
				offset++
				data = data[1:]
				cp_off = uint(data[0])
			}
			if (cmd & 0x02) != 0 {
				offset++
				data = data[1:]
				cp_off |= uint(data[0]) << 8
			}
			if (cmd & 0x04) != 0 {
				offset++
				data = data[1:]
				cp_off |= uint(data[0]) << 16
			}
			if (cmd & 0x08) != 0 {
				offset++
				data = data[1:]
				cp_off |= uint(data[0]) << 24
			}
			if (cmd & 0x10) != 0 {
				offset++
				data = data[1:]
				cp_size = uint(data[0])
			}
			if (cmd & 0x20) != 0 {
				offset++
				data = data[1:]
				cp_size |= uint(data[0]) << 8
			}
			if (cmd & 0x40) != 0 {
				offset++
				data = data[1:]
				cp_size |= uint(data[0]) << 16
			}
			if cp_size == 0 {
				cp_size = 0x10000
			}
			if cp_off+cp_size < cp_off ||
				cp_off+cp_size > uint(len(data)) ||
				cp_size > size {
				break
			}
			dest = append(append(dest[0:offset], data[cp_off:cp_off+cp_size]...), dest[offset+cp_off:]...)
			size -= cp_size
			offset += cp_size
		} else if cmd != 0 {
			if uint(cmd) > size {
				break
			}
			dest = append(append(dest[0:offset], data[0:uint(cmd)]...), dest[offset+uint(cmd):]...)
			size -= uint(cmd)
			offset += uint(cmd)
		} else {
			return nil
		}
		if size == 0 {
			break
		}
	}
	return dest
}
