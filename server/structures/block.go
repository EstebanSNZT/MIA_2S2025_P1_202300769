package structures

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
