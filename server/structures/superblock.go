package structures

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"server/utilities"
	"strings"
	"time"
)

type SuperBlock struct {
	FilesystemType  int32
	InodesCount     int32
	BlocksCount     int32
	FreeInodesCount int32
	FreeBlocksCount int32
	Mtime           int64
	Utime           int64
	MntCount        int32
	Magic           int32
	InodeSize       int32
	BlockSize       int32
	FirstIno        int32
	FirstBlo        int32
	BmInodeStart    int32
	BmBlockStart    int32
	InodeStart      int32
	BlockStart      int32
}

func NewSuperBlock(partition *Partition) *SuperBlock {
	superBlockSize := binary.Size(SuperBlock{})
	inodeSize := binary.Size(Inode{})
	blockSize := binary.Size(FileBlock{})
	n := CalculateStructureCount(partition.Size, superBlockSize, inodeSize, blockSize)
	bmInodeStart, bmBlockStart, inodeStart, blockStart := calculateSuperBlockOffsets(partition.Start, n, superBlockSize, inodeSize)

	return &SuperBlock{
		FilesystemType:  2,
		InodesCount:     n,
		BlocksCount:     3 * n,
		FreeInodesCount: n,
		FreeBlocksCount: 3 * n,
		Mtime:           int64(time.Now().Unix()),
		Utime:           0,
		MntCount:        1,
		Magic:           0xEF53,
		InodeSize:       int32(inodeSize),
		BlockSize:       int32(blockSize),
		FirstIno:        0,
		FirstBlo:        0,
		BmInodeStart:    bmInodeStart,
		BmBlockStart:    bmBlockStart,
		InodeStart:      inodeStart,
		BlockStart:      blockStart,
	}
}

func (s *SuperBlock) InitializeBitMaps(file *os.File) error {
	bmInodeBuffer := make([]byte, s.InodesCount)
	for i := range bmInodeBuffer {
		bmInodeBuffer[i] = '0'
	}

	if err := utilities.WriteBytes(file, bmInodeBuffer, int64(s.BmInodeStart)); err != nil {
		return fmt.Errorf("error al crear bitmap de inodos: %v", err)
	}

	bmBlockBuffer := make([]byte, s.BlocksCount)
	for i := range bmBlockBuffer {
		bmBlockBuffer[i] = '0'
	}

	if err := utilities.WriteBytes(file, bmBlockBuffer, int64(s.BmBlockStart)); err != nil {
		return fmt.Errorf("error al crear bitmap de bloques: %v", err)
	}

	return nil
}

func CalculateStructureCount(partitionSize int32, superBlockSize, inodeSize, blockSize int) int32 {
	numerator := int(partitionSize) - superBlockSize
	denominator := 4 + inodeSize + 3*blockSize
	return int32(math.Floor(float64(numerator) / float64(denominator)))
}

func calculateSuperBlockOffsets(partitionStart, n int32, superBlockSize, inodeSize int) (bmInodeStart, bmBlockStart, inodeStart, blockStart int32) {
	bmInodeStart = partitionStart + int32(superBlockSize)
	bmBlockStart = bmInodeStart + n
	inodeStart = bmBlockStart + (3 * n)
	blockStart = inodeStart + (int32(inodeSize) * n)
	return
}

func (s *SuperBlock) UpdateInodeBitmap(index int32, state [1]byte, file *os.File) error {
	offset := int64(s.BmInodeStart + index)
	if err := utilities.WriteBytes(file, state[:], offset); err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos: %v", err)
	}

	switch state[0] {
	case '1':
		s.FreeInodesCount--
	case '0':
		s.FreeInodesCount++
	}
	return nil
}

func (s *SuperBlock) UpdateBlockBitmap(index int32, state [1]byte, file *os.File) error {
	offset := int64(s.BmBlockStart + index)
	if err := utilities.WriteBytes(file, state[:], offset); err != nil {
		return fmt.Errorf("error al actualizar bitmap de bloques: %v", err)
	}

	switch state[0] {
	case '1':
		s.FreeBlocksCount--
	case '0':
		s.FreeBlocksCount++
	}
	return nil
}

func (s *SuperBlock) GetFreeInodeIndex(file *os.File) (int32, error) {
	bitmap, err := utilities.ReadBytes(file, int(s.InodesCount), int64(s.BmInodeStart))
	if err != nil {
		return -1, fmt.Errorf("error al leer bitmap de inodos: %v", err)
	}

	startIndex := s.FirstIno
	if startIndex < 0 || startIndex >= s.InodesCount {
		startIndex = 0
	}

	for i := range bitmap {
		checkIndex := (int(startIndex) + i) % len(bitmap)

		if bitmap[checkIndex] == '0' {
			s.FirstIno = int32(checkIndex) + 1
			return int32(checkIndex), nil
		}
	}

	return -1, fmt.Errorf("no se encontró un inodo libre")
}

func (s *SuperBlock) GetFreeBlockIndex(file *os.File) (int32, error) {
	bitmap, err := utilities.ReadBytes(file, int(s.BlocksCount), int64(s.BmBlockStart))
	if err != nil {
		return -1, fmt.Errorf("error al leer bitmap de bloques: %v", err)
	}

	startIndex := s.FirstBlo
	if startIndex < 0 || startIndex >= s.BlocksCount {
		startIndex = 0
	}

	for i := range bitmap {
		checkIndex := (int(startIndex) + i) % len(bitmap)

		if bitmap[checkIndex] == '0' {
			s.FirstBlo = int32(checkIndex) + 1
			return int32(checkIndex), nil
		}
	}

	return -1, fmt.Errorf("no se encontró un bloque libre")
}

func (s *SuperBlock) String() string {
	mTime := time.Unix(s.Mtime, 0).Format("2006-01-02 15:04:05")
	uTime := time.Unix(s.Utime, 0).Format("2006-01-02 15:04:05")

	return fmt.Sprintf(
		"------ SuperBlock ------\n- Filesystem Type: %d\n- Inodes Count: %d\n- Blocks Count: %d\n- Free Inodes Count: %d\n- Free Blocks Count: %d\n- Mtime: %s\n- Utime: %s\n- Mnt Count: %d\n- Magic: 0x%X\n- Inode Size: %d\n- Block Size: %d\n- First Inode: %d\n- First Block: %d\n- Bitmap Inode Start: %d\n- Bitmap Block Start: %d\n- Inode Start: %d\n- Block Start: %d",
		s.FilesystemType,
		s.InodesCount,
		s.BlocksCount,
		s.FreeInodesCount,
		s.FreeBlocksCount,
		mTime,
		uTime,
		s.MntCount,
		s.Magic,
		s.InodeSize,
		s.BlockSize,
		s.FirstIno,
		s.FirstBlo,
		s.BmInodeStart,
		s.BmBlockStart,
		s.InodeStart,
		s.BlockStart,
	)
}

func (s *SuperBlock) GenerateTable() string {
	var sb strings.Builder
	sb.WriteString(`digraph G {node [shape=plaintext]; table [label=<
	<table border="0" cellborder="1" cellspacing="0">
	<tr><td colspan="2" bgcolor="#ad63caff"><b>Reporte Superbloque</b></td></tr>`)

	tableRows := fmt.Sprintf(`
		<tr><td><b>s_filesystem_type</b></td><td>%d</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_inodes_count</b></td><td bgcolor="#edceffff">%d</td></tr>
		<tr><td><b>s_blocks_count</b></td><td>%d</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_free_inodes_count</b></td><td bgcolor="#edceffff">%d</td></tr>
		<tr><td><b>s_free_blocks_count</b></td><td>%d</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_mtime</b></td><td bgcolor="#edceffff">%s</td></tr>
		<tr><td><b>s_umtime</b></td><td>%s</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_mnt_count</b></td><td bgcolor="#edceffff">%d</td></tr>
		<tr><td><b>s_magic</b></td><td>0x%X</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_inode_size</b></td><td bgcolor="#edceffff">%d</td></tr>
		<tr><td><b>s_block_size</b></td><td>%d</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_first_ino</b></td><td bgcolor="#edceffff">%d</td></tr>
		<tr><td><b>s_first_blo</b></td><td>%d</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_bm_inode_start</b></td><td bgcolor="#edceffff">%d</td></tr>
		<tr><td><b>s_bm_block_start</b></td><td>%d</td></tr>
		<tr><td bgcolor="#edceffff"><b>s_inode_start</b></td><td bgcolor="#edceffff">%d</td></tr>
		<tr><td><b>s_block_start</b></td><td>%d</td></tr>
	`,
		s.FilesystemType,
		s.InodesCount,
		s.BlocksCount,
		s.FreeInodesCount,
		s.FreeBlocksCount,
		time.Unix(s.Mtime, 0).Format("2006-01-02 15:04:05"),
		time.Unix(s.Utime, 0).Format("2006-01-02 15:04:05"),
		s.MntCount,
		s.Magic,
		s.InodeSize,
		s.BlockSize,
		s.FirstIno,
		s.FirstBlo,
		s.BmInodeStart,
		s.BmBlockStart,
		s.InodeStart,
		s.BlockStart,
	)
	sb.WriteString(tableRows)
	sb.WriteString("</table>>];}")
	return sb.String()
}
