package bringauto_testing

import (
	"os"
	"fmt"
	"path/filepath"
)

const (
	fileExt = "_file"
	Pack1Name = "pack1"
	Pack2Name = "pack2"
	Pack3Name = "pack3"
	Pack1FileName = Pack1Name + fileExt
	Pack2FileName = Pack2Name + fileExt
	Pack3FileName = Pack3Name + fileExt
)

func SetupPackageFiles() error {
	err := os.Mkdir(Pack1Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}
	err = os.Mkdir(Pack2Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}
	err = os.Mkdir(Pack3Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}

	file1, err := os.Create(filepath.Join(Pack1Name, Pack1FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file1.Close()
	file2, err := os.Create(filepath.Join(Pack2Name, Pack2FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file2.Close()
	file3, err := os.Create(filepath.Join(Pack3Name, Pack3FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file3.Close()

	_, err = file1.WriteString("file1 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}
	_, err = file2.WriteString("file2 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}
	_, err = file3.WriteString("file3 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}

	return nil
}

func DeletePackageFiles() error {
	err := os.RemoveAll(Pack1Name)
	if err != nil {
		return err
	}
	err = os.RemoveAll(Pack2Name)
	if err != nil {
		return err
	}
	err = os.RemoveAll(Pack3Name)
	if err != nil {
		return err
	}
	return nil
}
