package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/indiefan/home_assistant_nanit/pkg/baby"
)

// Revision - marks the version of the structure of a session file. Only files with equal revision will be loaded
// Note: you should increment this whenever you change the Session structure
const Revision = 3

// Session - application session data container
type Session struct {
	Revision            int         `json:"revision"`
	AuthToken           string      `json:"authToken"`
	AuthTime            time.Time   `json:"authTime"`
	Babies              []baby.Baby `json:"babies"`
	RefreshToken        string      `json:"refreshToken"`
	LastSeenMessageTime time.Time   `json:"lastSeenMessageTime"`
}

// Store - application session store context
type Store struct {
	Filename string
	Session  *Session
}

// NewSessionStore - constructor
func NewSessionStore() *Store {
	return &Store{
		Session: &Session{Revision: Revision},
	}
}

// Load - loads previous state from a file
func (store *Store) Load() {
	if _, err := os.Stat(store.Filename); os.IsNotExist(err) {
		log.Info().Str("filename", store.Filename).Msg("No app session file found")
		return
	}

	f, err := os.Open(store.Filename)
	if err != nil {
		log.Fatal().Str("filename", store.Filename).Err(err).Msg("Unable to open app session file")
	}

	defer f.Close()

	session := &Session{}
	jsonErr := json.NewDecoder(f).Decode(session)
	if jsonErr != nil {
		log.Fatal().Str("filename", store.Filename).Err(jsonErr).Msg("Unable to decode app session file")
	}

	if session.Revision == Revision {
		store.Session = session
		log.Info().Str("filename", store.Filename).Msg("Loaded app session from the file")
	} else {
		log.Warn().Str("filename", store.Filename).Msg("App session file contains older revision of the state, ignoring")
	}

}

// Save - stores current data in a file
func (store *Store) Save() {
	if store.Filename == "" {
		return
	}

	log.Trace().Str("filename", store.Filename).Msg("Storing app session to the file")

	f, err := os.OpenFile(store.Filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal().Str("filename", store.Filename).Err(err).Msg("Unable to open app session file for writing")
	}

	defer f.Close()

	data, jsonErr := json.Marshal(store.Session)
	if jsonErr != nil {
		log.Fatal().Str("filename", store.Filename).Err(jsonErr).Msg("Unable to marshal contents of app session file")
	}

	_, writeErr := f.Write(data)
	if writeErr != nil {
		log.Fatal().Str("filename", store.Filename).Err(writeErr).Msg("Unable to wrote to app session file")
	}
}

// InitSessionStore - Initializes new application session store
func InitSessionStore(sessionFile string) *Store {
	sessionStore := NewSessionStore()

	// Load previous state of the application from session file
	if sessionFile != "" {

		absFileName, filePathErr := filepath.Abs(sessionFile)
		if filePathErr != nil {
			log.Fatal().Str("path", sessionFile).Err(filePathErr).Msg("Unable to retrieve absolute file path")
		}

		sessionStore.Filename = absFileName
		sessionStore.Load()
	}

	return sessionStore
}
