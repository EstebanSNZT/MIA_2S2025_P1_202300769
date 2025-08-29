package structures

import "fmt"

type FolderBlock struct {
	Content [4]FolderContent
}

type FolderContent struct {
	Name  [12]byte
	Inode int32
}

type FileBlock struct {
	Content [64]byte
}

type PointerBlock struct {
	Pointers [16]int32
}

func NewFileBlock(data []byte, start int32, blockSize int32) (*FileBlock, error) {
	fileSize := int32(len(data))
	if start >= fileSize {
		return nil, fmt.Errorf("inicio del índice fuera de rango")
	}

	end := min(start+blockSize, fileSize)
	fileBlock := &FileBlock{}
	copy(fileBlock.Content[:], data[start:end])
	return fileBlock, nil
}

func NewPointerBlock() *PointerBlock {
	return &PointerBlock{
		Pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
	}
}
