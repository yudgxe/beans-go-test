package sftpsender

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"github.com/yudgxe/beans-go-test/model"
)

const (
	PacketName       = "packet.zip"
	DependenciesName = "deps"
)

var ErrNoVerison = errors.New("no one version does not exist")

// GetLastVersion возвращает последнию версию, если она существует.
func GetLastVersion(sc *sftp.Client, dir string) (string, error) {
	files, err := sc.ReadDir(dir)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", ErrNoVerison
	}

	max := files[0].Name()
	for i := 1; i < len(files); i++ {
		name := files[i].Name()
		if max < name {
			max = name
		}
	}

	return max, nil
}

// UploadZip упаковка в zip и отправка.
func uploadZip(sc *sftp.Client, local []string, remotePath string, name string) error {
	if err := sc.MkdirAll(remotePath); err != nil {
		return err
	}

	zipfile, err := sc.OpenFile(sc.Join(remotePath, name), (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for _, source := range local {
		file, err := os.Open(source)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := archive.Create(source)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func uploadDeps(sc *sftp.Client, remotePath string, name string, packets []model.Packet) error {
	if err := sc.MkdirAll(remotePath); err != nil {
		return err
	}

	dstfile, err := sc.OpenFile(sc.Join(remotePath, name), (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return err
	}
	defer dstfile.Close()

	var sb strings.Builder
	for _, packet := range packets {
		sb.WriteString(fmt.Sprintf("%s %s\n", packet.Name, packet.Version))
	}

	if _, err := io.Copy(dstfile, strings.NewReader(sb.String())); err != nil {
		return err
	}

	return nil
}

func getDeps(sc *sftp.Client, remotePath, name string) ([]model.Packet, error) {
	remoteFile, err := sc.OpenFile(sc.Join(remotePath, name), (os.O_RDONLY))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	b, err := io.ReadAll(remoteFile)
	if err != nil {
		return nil, err
	}

	strs := strings.Fields(string(b))
	packets := make([]model.Packet, 0, len(strs)/2)
	for i := 0; i < len(strs); i += 2 {
		packets = append(packets, model.Packet{
			Name:    strs[i],
			Version: strs[i+1],
		})
	}

	return packets, nil
}

// RemoveDeps удаляет зависимости, если они существуют.
func removeDeps(sc *sftp.Client, remotePath, name string) error {
	if err := sc.Remove(sc.Join(remotePath, name)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return nil
}

func downloadFile(sc *sftp.Client, remote, local, name string) error {
	remoteFile, err := sc.OpenFile(sc.Join(remote, name), (os.O_RDONLY))
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	b, err := io.ReadAll(remoteFile)
	if err != nil {
		return err
	}
	reader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}

	destination, err := filepath.Abs(local)
	for _, f := range reader.File {
		destPath := filepath.Join(destination, f.Name)

		// Проверка на Zip Slip Vulnerability уязвимосить https://security.snyk.io/research/zip-slip-vulnerability
		if !strings.HasPrefix(destPath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", destPath)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}

		destFile, err := os.OpenFile(destPath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC), f.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		zipFile, err := f.Open()
		if err != nil {
			return err
		}
		defer zipFile.Close()

		if _, err := io.Copy(destFile, zipFile); err != nil {
			return err
		}
	}

	return nil
}
