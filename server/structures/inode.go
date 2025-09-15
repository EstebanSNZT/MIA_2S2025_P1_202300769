package structures

import (
	"fmt"
	"strings"
	"time"
)

type Inode struct {
	UID    int32
	GID    int32
	Size   int32
	Atime  int64
	Ctime  int64
	Mtime  int64
	Blocks [15]int32
	Type   [1]byte
	Perm   [3]byte
}

func NewInode(uid int32, gid int32, size int32, blockType [1]byte, perm [3]byte) *Inode {
	return &Inode{
		UID:    uid,
		GID:    gid,
		Size:   size,
		Atime:  time.Now().Unix(),
		Ctime:  time.Now().Unix(),
		Mtime:  time.Now().Unix(),
		Blocks: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Type:   blockType,
		Perm:   perm,
	}
}

func (i *Inode) PushBlock(blockIndex int32) error {
	for j := 0; j < len(i.Blocks); j++ {
		if i.Blocks[j] == -1 {
			i.Blocks[j] = blockIndex
			return nil
		}
	}
	return fmt.Errorf("no hay espacio en el inodo para más bloques")
}

func (i *Inode) String() string {
	aTime := time.Unix(i.Atime, 0).Format("2006-01-02 15:04:05")
	cTime := time.Unix(i.Ctime, 0).Format("2006-01-02 15:04:05")
	mTime := time.Unix(i.Mtime, 0).Format("2006-01-02 15:04:05")

	return fmt.Sprintf("------ Inode ------\n- Inode:\n- UID: %d\n- GID: %d\n- Size: %d\n- Atime: %s\n- Ctime: %s\n- Mtime: %s\n- Blocks: %v\n- Type: %s\n- Perm: %s\n",
		i.UID,
		i.GID,
		i.Size,
		aTime,
		cTime,
		mTime,
		i.Blocks,
		string(i.Type[:]),
		string(i.Perm[:]),
	)
}

func (i *Inode) UpdateAccessTime() {
	i.Atime = time.Now().Unix()
}

func (i *Inode) UpdateModificationTime() {
	i.Mtime = time.Now().Unix()
}

func (i *Inode) GenerateTable(index int32) string {
	var sb strings.Builder

	for j, block := range i.Blocks {
		color := "#ffffff"
		label := ""
		if j < 12 {
			label = "Directo"
			color = "#bee7ffff"
		} else if j < 13 {
			label = "Indirecto Simple"
			color = "#4bdffdff"
		} else if j < 14 {
			label = "Indirecto Doble"
			color = "#52ffc5ff"
		} else if j < 15 {
			label = "Indirecto Triple"
			color = "#83ff6aff"
		}
		sb.WriteString(fmt.Sprintf(`<tr><td bgcolor="%s"><b>%s [%d]</b></td><td PORT="p%d">%d</td></tr>`, color, label, j, j, block))
	}

	return fmt.Sprintf(`
		inode%d [label=<
			<table border="0" cellborder="1" cellspacing="0">
			<tr><td PORT="top" colspan="2" bgcolor="#f0e050ff"><b>Inodo %d</b></td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_uid</b></td><td>%d</td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_gid</b></td><td>%d</td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_size</b></td><td>%d</td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_atime</b></td><td>%s</td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_ctime</b></td><td>%s</td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_mtime</b></td><td>%s</td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_type</b></td><td>%c</td></tr>
			<tr><td bgcolor="#fff8beff"><b>i_perm</b></td><td>%s</td></tr>
			<tr><td colspan="2" bgcolor="#27b0ffff"><b>Bloques</b></td></tr>
			%s
			</table>
		>];
	`,
		index,
		index,
		i.UID,
		i.GID,
		i.Size,
		time.Unix(i.Atime, 0).Format("2006-01-02 15:04:05"),
		time.Unix(i.Ctime, 0).Format("2006-01-02 15:04:05"),
		time.Unix(i.Mtime, 0).Format("2006-01-02 15:04:05"),
		i.Type[0],
		string(i.Perm[:]),
		sb.String(),
	)
}

func (i *Inode) GetPermissionsString() string {
	var sb strings.Builder
	permMap := map[byte]string{
		'0': "---", '1': "--x", '2': "-w-", '3': "-wx",
		'4': "r--", '5': "r-x", '6': "rw-", '7': "rwx",
	}

	typePrefix := "-"
	if i.Type[0] == '0' {
		typePrefix = "d"
	}
	sb.WriteString(typePrefix)

	for _, p := range i.Perm {
		sb.WriteString(permMap[p])
	}
	return sb.String()
}
