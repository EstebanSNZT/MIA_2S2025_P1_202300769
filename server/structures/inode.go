package structures

type Inode struct {
	UID    int32
	GID    int32
	Size   int32
	Atime  float32
	Ctime  float32
	Mtime  float32
	Blocks [15]int32
	Type   [1]byte
	Perm   [3]byte
}
