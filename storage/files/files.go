package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"gogobot/events/telegram"
	"gogobot/lib/e"
	"gogobot/storage"
	"math/rand"
	"os"
	"path/filepath"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774

func NewStorage(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) Save(page *storage.Page) (err error) {

	defer func() { err = e.WrapIfErr("can`t save", err) }()

	filePath := filepath.Join(s.basePath, page.UserName)

	if err := os.MkdirAll(filePath, defaultPerm); err != nil {
		return err
	}
	fName, err := fileName(page)
	if err != nil {
		return err
	}

	filePath = filepath.Join(filePath, fName)

	file, err := os.Create(filePath)

	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}
	return nil

}

func (s Storage) PickRandom(userName string, mode string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can`t pick random page", err) }()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	n := rand.Intn(len(files))
	file := files[n]

	switch mode {
	case telegram.PickMode:

		filePath := filepath.Join(path, file.Name())

		defer func() {
			if removeErr := os.Remove(filePath); removeErr != nil {
				err = e.WrapIfErr("failed to remove file", removeErr)
			}
		}()

		return s.decodePage(filePath)
	case telegram.PeekMode:
		return s.decodePage(filepath.Join(path, file.Name()))

	}

	return nil, storage.ErrNoSavedPages
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can`t check if file exists", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	switch _, err := os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can`t check if file exists %s", path)
		return false, e.Wrap(msg, err)
	}
	return true, nil
}

func (s Storage) Remove(p *storage.Page) (err error) {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("can`t remove file", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can`t remove file %s", path)
		return e.Wrap(msg, err)
	}
	return nil
}

func (s Storage) decodePage(filePath string) (page *storage.Page, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode file", err)
	}
	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode file", err)
	}
	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}

func (s Storage) Count(userName string) (int, error) {
	dir := filepath.Join(s.basePath, userName)
	files, err := os.ReadDir(dir)

	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}

	if err != nil {
		return 0, e.Wrap("can't read dir", err)
	}

	return len(files), nil

}
