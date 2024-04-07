package main

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/indiefan/home_assistant_nanit/pkg/app"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
)

func ensureDataDirectories() app.DataDirectories {
	relDataDir := utils.EnvVarStr("NANIT_DATA_DIR", "/data")

	absDataDir, filePathErr := filepath.Abs(relDataDir)
	if filePathErr != nil {
		log.Fatal().Str("path", relDataDir).Err(filePathErr).Msg("Unable to retrieve absolute file path")
	}

	// Create base data directory if it does not exist
	if _, err := os.Stat(absDataDir); os.IsNotExist(err) {
		log.Warn().Str("dir", absDataDir).Msg("Data directory does not exist, creating")
		mkdirErr := os.MkdirAll(absDataDir, 0755)
		if mkdirErr != nil {
			log.Fatal().Str("path", absDataDir).Err(mkdirErr).Msg("Unable to create a directory")
		}
	}

	// Create data dir skeleton
	for _, subdirName := range []string{"video", "log"} {
		absSubdir := filepath.Join(absDataDir, subdirName)

		if _, err := os.Stat(absSubdir); os.IsNotExist(err) {
			mkdirErr := os.Mkdir(absSubdir, 0755)
			if mkdirErr != nil {
				log.Fatal().Str("path", absDataDir).Err(mkdirErr).Msg("Unable to create a directory")
			} else {
				log.Info().Str("dir", absSubdir).Msgf("Directory created ./%v", subdirName)
			}
		}
	}

	return app.DataDirectories{
		BaseDir:  absDataDir,
		VideoDir: filepath.Join(absDataDir, "video"),
		LogDir:   filepath.Join(absDataDir, "log"),
	}
}
