package structures

import (
	"fmt"
	"os"
	"path"
	"server/utilities"
	"strings"
)

type FileSystem struct {
	File *os.File
	Sb   *SuperBlock
}

func NewFileSystem(file *os.File, superBlock *SuperBlock) *FileSystem {
	return &FileSystem{
		File: file,
		Sb:   superBlock,
	}
}

func (fs *FileSystem) CreateUsersFile() error {
	rootInode := NewInode(1, 1, 0, [1]byte{'0'})
	rootInode.PushBlock(0)
	if err := utilities.WriteObject(fs.File, *rootInode, int64(fs.Sb.InodeStart)); err != nil {
		return err
	}
	if err := fs.Sb.UpdateInodeBitmap(fs.File, 0); err != nil {
		return err
	}

	userInodeIndex := (fs.Sb.FirstIno - fs.Sb.InodeStart) / fs.Sb.InodeSize

	rootBlock := &FolderBlock{
		Content: [4]FolderContent{
			{Name: [12]byte{'.'}, Inode: 0},
			{Name: [12]byte{'.', '.'}, Inode: 0},
			{Name: [12]byte{'u', 's', 'e', 'r', '.', 't', 'x', 't'}, Inode: userInodeIndex},
			{Name: [12]byte{'-'}, Inode: -1},
		},
	}

	if err := utilities.WriteObject(fs.File, *rootBlock, int64(fs.Sb.BlockStart)); err != nil {
		return err
	}
	if err := fs.Sb.UpdateBlockBitmap(fs.File, 0); err != nil {
		return err
	}

	userBlockIndex := (fs.Sb.FirstBlo - fs.Sb.BlockStart) / fs.Sb.BlockSize
	usersText := "1,G,root\n1,U,root,root,123\n"

	userInode := NewInode(1, 1, int32(len(usersText)), [1]byte{'1'})
	userInode.PushBlock(userBlockIndex)
	if err := utilities.WriteObject(fs.File, *userInode, int64(fs.Sb.FirstIno)); err != nil {
		return err
	}
	if err := fs.Sb.UpdateInodeBitmap(fs.File, userInodeIndex); err != nil {
		return err
	}

	usersBlock := &FileBlock{
		Content: [64]byte{},
	}

	copy(usersBlock.Content[:], usersText)

	if err := utilities.WriteObject(fs.File, *usersBlock, int64(fs.Sb.FirstBlo)); err != nil {
		return err
	}
	if err := fs.Sb.UpdateBlockBitmap(fs.File, userBlockIndex); err != nil {
		return err
	}

	return nil
}

func (fs *FileSystem) GetInodeByPath(input string) (*Inode, error) {
	clean := path.Clean(input)
	if clean == "/" || clean == "." || clean == "" {
		var root Inode
		if err := utilities.ReadObject(fs.File, &root, int64(fs.Sb.InodeStart)); err != nil {
			return nil, err
		}
		return &root, nil
	}

	parts := strings.FieldsFunc(clean, func(r rune) bool { return r == '/' })
	currentInodeIndex := int32(0)
	found := false

	for i := range parts {
		offset := int64(fs.Sb.InodeStart + currentInodeIndex*fs.Sb.InodeSize)
		var currentInode Inode
		if err := utilities.ReadObject(fs.File, &currentInode, offset); err != nil {
			return nil, err
		}

		if i < len(parts)-1 && currentInode.Type != [1]byte{'0'} {
			return nil, fmt.Errorf("el componente '%s' no es un directorio", parts[i])
		}

		for j := range currentInode.Blocks {
			if currentInode.Blocks[j] == -1 {
				continue
			}

			var folderBlock FolderBlock
			if err := utilities.ReadObject(fs.File, &folderBlock, int64(fs.Sb.BlockStart+currentInode.Blocks[j]*fs.Sb.BlockSize)); err != nil {
				return nil, err
			}

			for k := range folderBlock.Content {
				if folderBlock.Content[k].Inode == -1 {
					continue
				}

				name := strings.TrimRight(string(folderBlock.Content[k].Name[:]), "\x00")

				if name == parts[i] {
					currentInodeIndex = folderBlock.Content[k].Inode
					found = true
					break
				}
			}

			if found {
				break
			}
		}
	}

	offset := int64(fs.Sb.InodeStart + currentInodeIndex*fs.Sb.InodeSize)
	var targetInode Inode
	if err := utilities.ReadObject(fs.File, &targetInode, offset); err != nil {
		return nil, err
	}

	return &targetInode, nil
}

func (fs *FileSystem) ReadFileContent(inode *Inode) (string, error) {
	if inode.Type != [1]byte{'1'} {
		return "", fmt.Errorf("el inodo no es un archivo regular")
	}

	if inode.Size < 0 {
		return "", fmt.Errorf("el tamaño del archivo es inválido")
	}

	if inode.Size == 0 {
		return "", nil
	}

	var content strings.Builder
	content.Grow(int(inode.Size))

	readBlockContent := func(blockPtr int32) error {
		if int32(content.Len()) >= fs.Sb.BlockSize || blockPtr == -1 {
			return nil
		}

		if blockPtr < 0 || blockPtr >= fs.Sb.BlocksCount {
			return fmt.Errorf("puntero de bloque inválido: %d", blockPtr)
		}

		offset := int64(fs.Sb.BlockStart + blockPtr*fs.Sb.BlockSize)
		var fileBlock FileBlock
		if err := utilities.ReadObject(fs.File, &fileBlock, offset); err != nil {
			return err
		}

		bytesAvailableinBlock := min(int32(len(fileBlock.Content)), fs.Sb.BlockSize)
		bytesToCopy := min(bytesAvailableinBlock, inode.Size-int32(content.Len()))

		if bytesToCopy > 0 {
			written, err := content.Write(fileBlock.Content[:bytesToCopy])
			if err != nil {
				return fmt.Errorf("error escribiendo en buffer: %v", err)
			}
			if int32(written) != bytesToCopy {
				return fmt.Errorf("no se pudo escribir todo el contenido en el buffer")
			}
		}

		return nil
	}

	// Bloques directos
	for i := 0; i < 12; i++ {
		if int32(content.Len()) >= inode.Size {
			break
		}

		if err := readBlockContent(inode.Blocks[i]); err != nil {
			return "", fmt.Errorf("error en bloque directo [%d] (puntero %d): %w", i, inode.Blocks[i], err)
		}
	}

	var readBlockContentRecursive func(level int, blockPtr int32) error
	readBlockContentRecursive = func(level int, blockPtr int32) error {
		if int32(content.Len()) >= inode.Size || blockPtr == -1 {
			return nil
		}

		if blockPtr < 0 || blockPtr >= fs.Sb.BlocksCount {
			return fmt.Errorf("puntero de bloque inválido: %d", blockPtr)
		}

		offset := int64(fs.Sb.BlockStart + blockPtr*fs.Sb.BlockSize)
		var pointerBlock PointerBlock
		if err := utilities.ReadObject(fs.File, &pointerBlock, offset); err != nil {
			return err
		}

		for i := range pointerBlock.Pointers {
			if int32(content.Len()) >= inode.Size {
				break
			}

			nextPtr := pointerBlock.Pointers[i]
			if nextPtr == -1 {
				continue
			}

			if nextPtr < 0 || nextPtr >= fs.Sb.BlocksCount {
				return fmt.Errorf("puntero de bloque inválido: %d", nextPtr)
			}

			if level == 1 {
				if err := readBlockContent(nextPtr); err != nil {
					return fmt.Errorf("error en bloque directo [%d] (puntero %d): %w", i, nextPtr, err)
				}
			} else {
				if err := readBlockContentRecursive(level-1, nextPtr); err != nil {
					return fmt.Errorf("error procesando bloque de punteros %d (Nivel%d[%d] en bloque de punteros %d): %w", nextPtr, level, i, blockPtr, err)
				}
			}
		}

		return nil
	}

	// Apuntador indirecto simple (Nivel 1)
	if err := readBlockContentRecursive(1, inode.Blocks[12]); err != nil {
		return "", fmt.Errorf("error en indirección simple: %w", err)
	}

	// Apuntador indirecto doble (Nivel 2)
	if err := readBlockContentRecursive(2, inode.Blocks[13]); err != nil {
		return "", fmt.Errorf("error en indirección doble: %w", err)
	}

	// Apuntador indirecto triple (Nivel 3)
	if err := readBlockContentRecursive(3, inode.Blocks[14]); err != nil {
		return "", fmt.Errorf("error en indirección triple: %w", err)
	}

	if int32(content.Len()) != inode.Size {
		return "", fmt.Errorf("lectura incompleta, se esperaban %d bytes pero se leyeron %d", inode.Size, content.Len())
	}

	return content.String(), nil
}
