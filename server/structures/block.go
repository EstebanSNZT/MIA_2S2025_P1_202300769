package structures

type FolderBlock struct {
	content [4]FolderContent
}

type FolderContent struct {
	name  [12]byte
	inode int32
}

type FileBlock struct {
	content [64]byte
}

type PointerBlock struct {
	Pointers [16]int32
}
