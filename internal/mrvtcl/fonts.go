//go:build windows && 386

package mrvtcl

import (
	"fmt"
	"path/filepath"
)

func LoadFonts(fontDir string) (int, error) {
	loaded := 0

	jtfPattern := fmt.Sprintf("%s\\*.jtf", fontDir)
	hFind := FindFirstFileA(jtfPattern, &WIN32_FIND_DATAA{})
	if hFind != INVALID_HANDLE_VALUE {
		defer FindClose(hFind)

		findData := WIN32_FIND_DATAA{}
		for {
			filename := nullTerminatedString(findData.FileName[:])
			if filename != "" {
				fontPath := fmt.Sprintf("%s\\%s", fontDir, filename)
				if AddFontResourceExA(fontPath, FR_PRIVATE, 0) > 0 {
					loadedFonts = append(loadedFonts, fontPath)
					loaded++
				}
			}

			if !FindNextFileA(hFind, &findData) {
				break
			}
		}
	}

	ttfPattern := fmt.Sprintf("%s\\*.ttf", fontDir)
	hFind = FindFirstFileA(ttfPattern, &WIN32_FIND_DATAA{})
	if hFind != INVALID_HANDLE_VALUE {
		defer FindClose(hFind)

		findData := WIN32_FIND_DATAA{}
		for {
			filename := nullTerminatedString(findData.FileName[:])
			if filename != "" {
				fontPath := fmt.Sprintf("%s\\%s", fontDir, filename)
				if AddFontResourceExA(fontPath, FR_PRIVATE, 0) > 0 {
					loadedFonts = append(loadedFonts, fontPath)
					loaded++
				}
			}

			if !FindNextFileA(hFind, &findData) {
				break
			}
		}
	}

	if loaded > 0 {
		BroadcastFontChange()
	}

	return loaded, nil
}

func UnloadFonts() {
	for _, fontPath := range loadedFonts {
		RemoveFontResourceExA(fontPath, FR_PRIVATE, 0)
	}
	loadedFonts = nil

	if len(loadedFonts) > 0 {
		BroadcastFontChange()
	}
}

func nullTerminatedString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

func FindFontFiles(dir string) ([]string, error) {
	var fonts []string

	pattern := filepath.Join(dir, "*.jtf")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		fonts = append(fonts, matches...)
	}

	pattern = filepath.Join(dir, "*.ttf")
	matches, err = filepath.Glob(pattern)
	if err == nil {
		fonts = append(fonts, matches...)
	}

	return fonts, nil
}
