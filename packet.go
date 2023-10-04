package sftpsender

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/yudgxe/beans-go-test/model"
)

// getFiles возвращает слайсл всех файлов подходящих под паттерн.
func getFiles(pattern string) ([]string, error) {
	response := make([]string, 0)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			filepath.Walk(match, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					response = append(response, path)
				}
				return nil
			})
		} else {
			response = append(response, match)
		}
	}

	return response, nil
}

// Create собирает все необходимые файлы отправляет на сервер.
func Create(sc *sftp.Client, packet model.Create) error {
	files := make([]string, 0)
	for _, target := range packet.Targets {
		f, err := getFiles(target.Path)
		if err != nil {
			return err
		}
		files = append(files, f...)
	}

	remotePath := sc.Join(packet.Name, packet.Version)
	if err := uploadZip(sc, files, remotePath, PacketName); err != nil {
		return err
	}

	if packet.Packets != nil {
		if err := uploadDeps(sc, remotePath, DependenciesName, packet.Packets); err != nil {
			return err
		}
	} else {
		if err := removeDeps(sc, remotePath, DependenciesName); err != nil {
			return err
		}
	}

	return nil
}

// Update скачивает необходиые пакеты и распаковывает их.
// Если версия не указана будет скачан пакет с самой новой версией.
func Update(sc *sftp.Client, packages model.Update) error {
	var uerr UpdateError
	for _, p := range packages.Packets {
		if p.Version == "" {
			ver, err := GetLastVersion(sc, p.Name)
			if err != nil {
				if errors.Is(err, ErrNoVerison) {
					uerr.Errs = append(uerr.Errs, fmt.Errorf("Packet: %s error: %w", p.Name, err))
					continue
				}
				if errors.Is(err, os.ErrNotExist) {
					uerr.Errs = append(uerr.Errs, fmt.Errorf("Packet: %s version: %s error: %w", p.Name, p.Version, err))
					continue
				}
				uerr.Errs = append(uerr.Errs, err)
				continue
			}
			p.Version = ver
		}
		remotePath := sc.Join(p.Name, p.Version)
		if err := downloadFile(sc, remotePath, p.Name+"@"+p.Version, PacketName); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				uerr.Errs = append(uerr.Errs, fmt.Errorf("Packet: %s version: %s error: %w", p.Name, p.Version, err))
				continue
			}
			uerr.Errs = append(uerr.Errs, err)
			continue
		}

		deps, err := getDeps(sc, remotePath, DependenciesName)
		if err != nil {
			uerr.Errs = append(uerr.Errs, err)
			continue
		}

		for _, deps := range deps {
			remotePath := sc.Join(deps.Name, deps.Version)
			if err := downloadFile(sc, remotePath, deps.Name+"@"+deps.Version, PacketName); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					uerr.Errs = append(uerr.Errs,
						fmt.Errorf("For packet: %s version: %s not found deps packet: %s version: %s", p.Name, p.Version, deps.Name, deps.Version))
					continue
				}
				uerr.Errs = append(uerr.Errs, err)
				continue
			}
		}
	}

	if len(uerr.Errs) > 0 {
		return uerr
	}

	return nil
}
