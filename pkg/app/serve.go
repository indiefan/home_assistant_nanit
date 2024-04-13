package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/indiefan/home_assistant_nanit/pkg/baby"
)

func serve(babies []baby.Baby, dataDir DataDirectories) {
	const port = 8080

	// Index handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		for _, baby := range babies {
			fmt.Fprintf(w, "<video src=\"/video/%v.m3u8\" controls autoplay width=\"1280\" height=\"960\"></video>", baby.UID)
		}
	})

	// Video files
	http.Handle("/video/", http.StripPrefix("/video/", http.FileServer(http.Dir(dataDir.VideoDir))))

	// Dummy log handler - useful for receiving logs from cam
	// Note: Cam is sending tared archive through curl as binary file
	// TODO: proper handling of Expect: 100-continue
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		filename := filepath.Join(dataDir.LogDir, fmt.Sprintf("camlogs-%v.tar.gz", time.Now().Format(time.RFC3339)))

		log.Info().Str("file", filename).Msg("Saving log to file")
		defer r.Body.Close()

		out, err := os.Create(filename)
		if err != nil {
			log.Error().Str("file", filename).Err(err).Msg("Unable to create file")
		}

		defer out.Close()

		_, err = io.Copy(out, r.Body)

		if err != nil {
			log.Error().Str("file", filename).Err(err).Msg("Unable to save received log file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	log.Info().Int("port", port).Msg("Starting HTTP server")
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
