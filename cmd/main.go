package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	vbutton "github.com/xoltia/v-button"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "vbutton.db")

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer db.Close()

	repo := vbutton.NewVoiceClipDB(db)

	if err = repo.Create(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	storage := vbutton.NewFileSystemStorage("storage")

	if err = storage.Init(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	encoder := vbutton.NewOpusOggEncoder()
	altEncoder := vbutton.NewOpusCafEncoder()

	vc := vbutton.NewVoiceClipService(repo, storage, []vbutton.AudioEncoder{encoder, altEncoder})

	http.Handle("/", vbutton.NewIndexHandler(vc))
	http.Handle("/submit", vbutton.NewSubmitHandler(vc))
	http.Handle("/tos", vbutton.NewTOSHandler())
	http.Handle("/update", vbutton.NewUpdateHandler(vc))
	http.Handle("/audio/", http.StripPrefix("/audio/", http.HandlerFunc(storage.ServeFile)))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/style/", http.StripPrefix("/style/", http.FileServer(http.Dir("style/dist"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
