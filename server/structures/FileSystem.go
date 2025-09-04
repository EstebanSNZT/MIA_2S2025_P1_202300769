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
	rootInodeIndex, err := fs.Sb.GetFreeInodeIndex(fs.File)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar un inodo libre para la raíz: %w", err)
	}

	rootBlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar un bloque libre para la raíz: %w", err)
	}

	usersInodeIndex, err := fs.Sb.GetFreeInodeIndex(fs.File)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar un inodo libre para users.txt: %w", err)
	}

	usersBlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar un bloque libre para users.txt: %w", err)
	}

	rootInode := NewInode(1, 1, 0, [1]byte{'0'}, [3]byte{'7', '7', '7'})
	rootInode.PushBlock(rootBlockIndex)

	usersText := "1,G,root\n1,U,root,root,123\n"
	userInode := NewInode(1, 1, int32(len(usersText)), [1]byte{'1'}, [3]byte{'7', '7', '7'})
	userInode.PushBlock(usersBlockIndex)

	rootBlock := &FolderBlock{
		Content: [4]FolderContent{
			{Name: [12]byte{'.'}, Inode: rootInodeIndex},
			{Name: [12]byte{'.', '.'}, Inode: rootInodeIndex},
			{Name: [12]byte{'u', 's', 'e', 'r', '.', 't', 'x', 't'}, Inode: usersInodeIndex},
			{Name: [12]byte{'-'}, Inode: -1},
		},
	}

	usersBlock := &FileBlock{}
	copy(usersBlock.Content[:], usersText)

	rootInodeOffset := int64(fs.Sb.InodeStart) + int64(rootInodeIndex)*int64(fs.Sb.InodeSize)
	if err := utilities.WriteObject(fs.File, *rootInode, rootInodeOffset); err != nil {
		return err
	}

	userInodeOffset := int64(fs.Sb.InodeStart) + int64(usersInodeIndex)*int64(fs.Sb.InodeSize)
	if err := utilities.WriteObject(fs.File, *userInode, userInodeOffset); err != nil {
		return err
	}

	rootBlockOffset := int64(fs.Sb.BlockStart) + int64(rootBlockIndex)*int64(fs.Sb.BlockSize)
	if err := utilities.WriteObject(fs.File, *rootBlock, rootBlockOffset); err != nil {
		return err
	}

	usersBlockOffset := int64(fs.Sb.BlockStart) + int64(usersBlockIndex)*int64(fs.Sb.BlockSize)
	if err := utilities.WriteObject(fs.File, *usersBlock, usersBlockOffset); err != nil {
		return err
	}

	if err := fs.Sb.UpdateInodeBitmap(rootInodeIndex, [1]byte{'1'}, fs.File); err != nil {
		return err
	}

	if err := fs.Sb.UpdateBlockBitmap(rootBlockIndex, [1]byte{'1'}, fs.File); err != nil {
		return err
	}

	if err := fs.Sb.UpdateInodeBitmap(usersInodeIndex, [1]byte{'1'}, fs.File); err != nil {
		return err
	}

	if err := fs.Sb.UpdateBlockBitmap(usersBlockIndex, [1]byte{'1'}, fs.File); err != nil {
		return err
	}

	return nil
}

func (fs *FileSystem) GetInodeByPath(input string) (*Inode, int32, error) {
	clean := path.Clean(input)
	if clean == "/" || clean == "." || clean == "" {
		var root Inode
		if err := utilities.ReadObject(fs.File, &root, int64(fs.Sb.InodeStart)); err != nil {
			return nil, -1, err
		}
		return &root, 0, nil
	}

	parts := strings.FieldsFunc(clean, func(r rune) bool { return r == '/' })
	currentInodeIndex := int32(0)

	for _, part := range parts {
		var currentInode Inode
		if err := utilities.ReadObject(fs.File, &currentInode, int64(fs.Sb.InodeStart+currentInodeIndex*fs.Sb.InodeSize)); err != nil {
			return nil, -1, err
		}

		nextInodeIndex, err := fs.GetInodeIndexByName(&currentInode, part)
		if err != nil {
			return nil, -1, err
		}

		if nextInodeIndex == -1 {
			return nil, -1, fmt.Errorf("el componente '%s' no se encontró", part)
		}
		currentInodeIndex = nextInodeIndex
	}

	offset := int64(fs.Sb.InodeStart + currentInodeIndex*fs.Sb.InodeSize)
	var targetInode Inode
	if err := utilities.ReadObject(fs.File, &targetInode, offset); err != nil {
		return nil, -1, err
	}

	return &targetInode, currentInodeIndex, nil
}

func (fs *FileSystem) GetInodeIndexByName(inode *Inode, name string) (int32, error) {
	for _, blockIndex := range inode.Blocks {
		if blockIndex == -1 {
			continue
		}

		var folderBlock FolderBlock
		if err := utilities.ReadObject(fs.File, &folderBlock, int64(fs.Sb.BlockStart+blockIndex*fs.Sb.BlockSize)); err != nil {
			return -1, err
		}

		for j := range folderBlock.Content {
			if folderBlock.Content[j].Inode == -1 {
				continue
			}

			inode_name := strings.TrimRight(string(folderBlock.Content[j].Name[:]), "\x00")
			if inode_name == name {
				return folderBlock.Content[j].Inode, nil
			}
		}
	}

	return -1, nil
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

	readBlockContent := func(blockIndex int32) error {
		if blockIndex == -1 {
			return nil
		}

		if blockIndex < 0 || blockIndex >= fs.Sb.BlocksCount {
			return fmt.Errorf("puntero de bloque inválido: %d", blockIndex)
		}

		offset := int64(fs.Sb.BlockStart + blockIndex*fs.Sb.BlockSize)
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
		if int32(content.Len()) >= inode.Size || blockPtr == -1 || level < 1 || level > 3 {
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

func (fs *FileSystem) FreeFileInode(inode *Inode) error {
	freeFileBlock := func(blockPtr int32) error {
		if blockPtr < 0 {
			return nil
		}

		if err := fs.Sb.UpdateBlockBitmap(blockPtr, [1]byte{'0'}, fs.File); err != nil {
			return fmt.Errorf("error al liberar bloque de archivo %d: %v", blockPtr, err)
		}

		return nil
	}

	// Liberar bloques directos
	for i := 0; i < 12; i++ {
		if err := freeFileBlock(inode.Blocks[i]); err != nil {
			return err
		}
		inode.Blocks[i] = -1
	}

	var freeFileBlockRecursive func(level int, blockPtr int32) error
	freeFileBlockRecursive = func(level int, blockPtr int32) error {
		if blockPtr == -1 || level < 1 || level > 3 {
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
			nextPtr := pointerBlock.Pointers[i]
			if nextPtr == -1 {
				continue
			}

			if nextPtr < 0 || nextPtr >= fs.Sb.BlocksCount {
				return fmt.Errorf("puntero de bloque inválido: %d", nextPtr)
			}

			if level == 1 {
				if err := freeFileBlock(nextPtr); err != nil {
					return fmt.Errorf("error liberando bloque directo [%d] (puntero %d): %w", i, nextPtr, err)
				}
			} else {
				if err := freeFileBlockRecursive(level-1, nextPtr); err != nil {
					return fmt.Errorf("error liberando bloque de punteros %d (Nivel%d[%d] en bloque de punteros %d): %w", nextPtr, level, i, blockPtr, err)
				}
			}
		}

		return freeFileBlock(blockPtr)
	}

	// Liberar bloques de punteros
	if err := freeFileBlockRecursive(1, inode.Blocks[12]); err != nil {
		return fmt.Errorf("error liberando indirección simple: %w", err)
	}
	inode.Blocks[12] = -1

	// Liberar bloques de punteros dobles
	if err := freeFileBlockRecursive(2, inode.Blocks[13]); err != nil {
		return fmt.Errorf("error liberando indirección doble: %w", err)
	}
	inode.Blocks[13] = -1

	// Liberar bloques de punteros triples
	if err := freeFileBlockRecursive(3, inode.Blocks[14]); err != nil {
		return fmt.Errorf("error liberando indirección triple: %w", err)
	}
	inode.Blocks[14] = -1

	inode.Size = 0
	return nil
}

func (fs *FileSystem) AllocateFileBlocks(content []byte) ([15]int32, error) {
	allocatedBlocks := [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	numBlocksNeeded := (int32(len(content)) + fs.Sb.BlockSize - 1) / fs.Sb.BlockSize
	pointerPerBlock := int32(len(PointerBlock{}.Pointers))

	if numBlocksNeeded > fs.Sb.FreeBlocksCount {
		return allocatedBlocks, fmt.Errorf("no hay suficientes bloques libres para escribir el archivo")
	}

	directLimit := int32(12)
	simpleLimit := directLimit + pointerPerBlock
	doubleLimit := simpleLimit + pointerPerBlock*pointerPerBlock
	tripleLimit := doubleLimit + pointerPerBlock*pointerPerBlock*pointerPerBlock
	var PointerBlockCache = make(map[int32]*PointerBlock)

	for i := int32(0); i < numBlocksNeeded; i++ {
		fileBlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
		if err != nil {
			return allocatedBlocks, fmt.Errorf("error al obtener bloque libre: %w", err)
		}

		fileBlock, err := NewFileBlock(content, i*fs.Sb.BlockSize, fs.Sb.BlockSize)
		if err != nil {
			return allocatedBlocks, fmt.Errorf("error al crear bloque de archivo: %w", err)
		}

		fileBlockOffset := int64(fs.Sb.BlockStart + fileBlockIndex*fs.Sb.BlockSize)
		if err := utilities.WriteObject(fs.File, *fileBlock, fileBlockOffset); err != nil {
			return allocatedBlocks, fmt.Errorf("error al escribir bloque de archivo: %w", err)
		}

		if err := fs.Sb.UpdateBlockBitmap(fileBlockIndex, [1]byte{'1'}, fs.File); err != nil {
			return allocatedBlocks, fmt.Errorf("error al actualizar bitmap de bloques: %w", err)
		}

		if i < directLimit {
			allocatedBlocks[i] = fileBlockIndex
			continue
		}

		if i < simpleLimit {
			indexInSimple := i - directLimit

			if allocatedBlocks[12] == -1 {
				simpleBlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
				if err != nil {
					return allocatedBlocks, fmt.Errorf("error al obtener bloque libre para punteros simples: %w", err)
				}
				allocatedBlocks[12] = simpleBlockIndex
				PointerBlockCache[simpleBlockIndex] = NewPointerBlock()
				if err := fs.Sb.UpdateBlockBitmap(simpleBlockIndex, [1]byte{'1'}, fs.File); err != nil {
					return allocatedBlocks, fmt.Errorf("error al actualizar bitmap de bloques: %w", err)
				}
			}

			PointerBlockCache[allocatedBlocks[12]].Pointers[indexInSimple] = fileBlockIndex
			continue
		}

		if i < doubleLimit {
			indexInDouble := i - simpleLimit
			level1Index := indexInDouble / pointerPerBlock
			level2Index := indexInDouble % pointerPerBlock

			if allocatedBlocks[13] == -1 {
				level1BlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
				if err != nil {
					return allocatedBlocks, fmt.Errorf("error al obtener bloque libre para punteros dobles: %w", err)
				}
				allocatedBlocks[13] = level1BlockIndex
				PointerBlockCache[level1BlockIndex] = NewPointerBlock()
				if err := fs.Sb.UpdateBlockBitmap(level1BlockIndex, [1]byte{'1'}, fs.File); err != nil {
					return allocatedBlocks, fmt.Errorf("error al actualizar bitmap de bloques: %w", err)
				}
			}

			level1PointerBlock := PointerBlockCache[allocatedBlocks[13]]

			if level1PointerBlock.Pointers[level1Index] == -1 {
				level2BlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
				if err != nil {
					return allocatedBlocks, fmt.Errorf("error al obtener bloque libre para punteros simples de punteros dobles: %w", err)
				}
				level1PointerBlock.Pointers[level1Index] = level2BlockIndex
				PointerBlockCache[level2BlockIndex] = NewPointerBlock()
				if err := fs.Sb.UpdateBlockBitmap(level2BlockIndex, [1]byte{'1'}, fs.File); err != nil {
					return allocatedBlocks, fmt.Errorf("error al actualizar bitmap de bloques: %w", err)
				}
			}

			level2PointerBlock := PointerBlockCache[level1PointerBlock.Pointers[level1Index]]
			level2PointerBlock.Pointers[level2Index] = fileBlockIndex
			continue
		}

		if i < tripleLimit {
			indexInTriple := i - doubleLimit
			level1Index := indexInTriple / (pointerPerBlock * pointerPerBlock)
			level2Index := (indexInTriple / pointerPerBlock) % pointerPerBlock
			level3Index := indexInTriple % pointerPerBlock

			if allocatedBlocks[14] == -1 {
				level1BlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
				if err != nil {
					return allocatedBlocks, fmt.Errorf("error al obtener bloque libre para punteros triples: %w", err)
				}
				allocatedBlocks[14] = level1BlockIndex
				PointerBlockCache[level1BlockIndex] = NewPointerBlock()
				if err := fs.Sb.UpdateBlockBitmap(level1BlockIndex, [1]byte{'1'}, fs.File); err != nil {
					return allocatedBlocks, fmt.Errorf("error al actualizar bitmap de bloques: %w", err)
				}
			}

			level1PointerBlock := PointerBlockCache[allocatedBlocks[14]]

			if level1PointerBlock.Pointers[level1Index] == -1 {
				level2BlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
				if err != nil {
					return allocatedBlocks, fmt.Errorf("error al obtener bloque libre para punteros dobles de punteros triples: %w", err)
				}
				level1PointerBlock.Pointers[level1Index] = level2BlockIndex
				PointerBlockCache[level2BlockIndex] = NewPointerBlock()
				if err := fs.Sb.UpdateBlockBitmap(level2BlockIndex, [1]byte{'1'}, fs.File); err != nil {
					return allocatedBlocks, fmt.Errorf("error al actualizar bitmap de bloques: %w", err)
				}
			}

			level2PointerBlock := PointerBlockCache[level1PointerBlock.Pointers[level1Index]]
			if level2PointerBlock.Pointers[level2Index] == -1 {
				level3BlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
				if err != nil {
					return allocatedBlocks, fmt.Errorf("error al obtener bloque libre para punteros triples: %w", err)
				}
				level2PointerBlock.Pointers[level2Index] = level3BlockIndex
				PointerBlockCache[level3BlockIndex] = NewPointerBlock()
				if err := fs.Sb.UpdateBlockBitmap(level3BlockIndex, [1]byte{'1'}, fs.File); err != nil {
					return allocatedBlocks, fmt.Errorf("error al actualizar bitmap de bloques: %w", err)
				}
			}

			level3PointerBlock := PointerBlockCache[level2PointerBlock.Pointers[level2Index]]
			level3PointerBlock.Pointers[level3Index] = fileBlockIndex
		}
	}

	for pointerBlockIndex, pointerBlock := range PointerBlockCache {
		offset := int64(fs.Sb.BlockStart + pointerBlockIndex*fs.Sb.BlockSize)
		if err := utilities.WriteObject(fs.File, *pointerBlock, offset); err != nil {
			return allocatedBlocks, fmt.Errorf("error al escribir bloque de punteros: %w", err)
		}
	}

	return allocatedBlocks, nil
}

func (fs *FileSystem) AddEntryToParent(parentInode *Inode, parentInodeIndex int32, entryName string, entryInodeIndex int32) error {
	firstFreeSlot := -1

	for i, blockIndex := range parentInode.Blocks {
		if blockIndex == -1 {
			if firstFreeSlot == -1 {
				firstFreeSlot = i
			}
			continue
		}

		offset := int64(fs.Sb.BlockStart + blockIndex*fs.Sb.BlockSize)
		var folderBlock FolderBlock
		if err := utilities.ReadObject(fs.File, &folderBlock, offset); err != nil {
			return fmt.Errorf("error al leer bloque de carpeta: %w", err)
		}

		for j := range folderBlock.Content {
			if folderBlock.Content[j].Inode == -1 {
				folderBlock.Content[j].Inode = entryInodeIndex
				copy(folderBlock.Content[j].Name[:], entryName)

				if err := utilities.WriteObject(fs.File, folderBlock, offset); err != nil {
					return fmt.Errorf("no se pudo escribir el bloque de directorio modificado %d: %w", blockIndex, err)
				}
				return nil
			}
		}
	}

	if firstFreeSlot == -1 {
		return fmt.Errorf("no hay espacio libre en el directorio")

	}

	newBlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
	if err != nil {
		return fmt.Errorf("no se pudo obtener un bloque libre para el nuevo bloque de carpeta: %w", err)
	}

	parentInode.Blocks[firstFreeSlot] = newBlockIndex
	newFolderBlock := NewFolderBlock()
	newFolderBlock.Content[0].Inode = entryInodeIndex
	copy(newFolderBlock.Content[0].Name[:], entryName)

	newBlockOffset := int64(fs.Sb.BlockStart + newBlockIndex*fs.Sb.BlockSize)
	if err := utilities.WriteObject(fs.File, newFolderBlock, newBlockOffset); err != nil {
		return fmt.Errorf("no se pudo escribir el nuevo bloque de carpeta: %w", err)
	}

	if err := fs.Sb.UpdateBlockBitmap(newBlockIndex, [1]byte{'1'}, fs.File); err != nil {
		return err
	}

	parentInode.UpdateModificationTime()
	parentOffset := int64(fs.Sb.InodeStart + parentInodeIndex*fs.Sb.InodeSize)
	if err := utilities.WriteObject(fs.File, *parentInode, parentOffset); err != nil {
		return err
	}

	return nil
}

func (fs *FileSystem) CreateNewFolder(parentIndex int32, UID int32, GID int32) (int32, error) {
	folderInodeIndex, err := fs.Sb.GetFreeInodeIndex(fs.File)
	if err != nil {
		return -1, fmt.Errorf("no se pudo encontrar un inodo libre para la nueva carpeta: %w", err)
	}

	folderBlockIndex, err := fs.Sb.GetFreeBlockIndex(fs.File)
	if err != nil {
		return -1, fmt.Errorf("no se pudo encontrar un bloque libre para la nueva carpeta: %w", err)
	}

	folderInode := NewInode(UID, GID, 0, [1]byte{'0'}, [3]byte{'6', '6', '4'})
	folderInode.PushBlock(folderBlockIndex)

	folderBlock := NewFolderBlock()
	folderBlock.Content[0] = FolderContent{Name: [12]byte{'.'}, Inode: folderInodeIndex}
	folderBlock.Content[1] = FolderContent{Name: [12]byte{'.', '.'}, Inode: parentIndex}

	folderInodeOffset := int64(fs.Sb.InodeStart) + int64(folderInodeIndex)*int64(fs.Sb.InodeSize)
	if err := utilities.WriteObject(fs.File, *folderInode, folderInodeOffset); err != nil {
		return -1, err
	}

	folderBlockOffset := int64(fs.Sb.BlockStart) + int64(folderBlockIndex)*int64(fs.Sb.BlockSize)
	if err := utilities.WriteObject(fs.File, folderBlock, folderBlockOffset); err != nil {
		return -1, err
	}

	if err := fs.Sb.UpdateInodeBitmap(folderInodeIndex, [1]byte{'1'}, fs.File); err != nil {
		return -1, err
	}

	if err := fs.Sb.UpdateBlockBitmap(folderBlockIndex, [1]byte{'1'}, fs.File); err != nil {
		return -1, err
	}

	return folderInodeIndex, nil
}

func (fs *FileSystem) EnsurePathExist(path string, UID int32, GID int32) (*Inode, int32, error) {
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
	currentInodeIndex := int32(0)

	for _, part := range parts {
		var currentInode Inode
		if err := utilities.ReadObject(fs.File, &currentInode, int64(fs.Sb.InodeStart+currentInodeIndex*fs.Sb.InodeSize)); err != nil {
			return nil, -1, err
		}

		if currentInode.Type != [1]byte{'0'} {
			return nil, -1, fmt.Errorf("no se puede crear: '%s' no es un directorio en la ruta '%s'", part, path)
		}

		nextInodeIndex, err := fs.GetInodeIndexByName(&currentInode, part)
		if err != nil {
			return nil, -1, err
		}

		if nextInodeIndex == -1 {
			newFolderInodeIndex, err := fs.CreateNewFolder(currentInodeIndex, UID, GID)
			if err != nil {
				return nil, -1, err
			}
			if err := fs.AddEntryToParent(&currentInode, currentInodeIndex, part, newFolderInodeIndex); err != nil {
				return nil, -1, err
			}
			nextInodeIndex = newFolderInodeIndex
		}

		currentInodeIndex = nextInodeIndex
	}

	finalOffset := int64(fs.Sb.InodeStart + currentInodeIndex*fs.Sb.InodeSize)
	var lastFolderInode Inode
	if err := utilities.ReadObject(fs.File, &lastFolderInode, finalOffset); err != nil {
		return nil, -1, err
	}

	return &lastFolderInode, currentInodeIndex, nil
}
