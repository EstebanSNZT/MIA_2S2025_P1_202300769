package structures

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"server/utilities"
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
		InodesCount:     0,
		BlocksCount:     0,
		FreeInodesCount: n,
		FreeBlocksCount: 3 * n,
		Mtime:           int64(time.Now().Unix()),
		Utime:           0,
		MntCount:        1,
		Magic:           0xEF53,
		InodeSize:       int32(inodeSize),
		BlockSize:       int32(blockSize),
		FirstIno:        inodeStart,
		FirstBlo:        blockStart,
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

func (s *SuperBlock) UpdateInodeBitmap(file *os.File, index int32) error {
	offset := int64(s.BmInodeStart) + int64(index)
	if err := utilities.WriteBytes(file, []byte{'1'}, offset); err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos: %v", err)
	}
	s.InodesCount++
	s.FreeInodesCount--
	s.FirstIno += s.InodeSize
	return nil
}

func (s *SuperBlock) UpdateBlockBitmap(file *os.File, index int32) error {
	offset := int64(s.BmBlockStart) + int64(index)
	if err := utilities.WriteBytes(file, []byte{'1'}, offset); err != nil {
		return fmt.Errorf("error al actualizar bitmap de bloques: %v", err)
	}
	s.BlocksCount++
	s.FreeBlocksCount--
	s.FirstBlo += s.BlockSize
	return nil
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
