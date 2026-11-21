package web

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ToYoNiX/stash/internal/app"
	"github.com/ToYoNiX/stash/internal/containers"
	"github.com/ToYoNiX/stash/internal/items"
)

//go:embed page.html
var files embed.FS

const maxUpload = 20 << 20

var imageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type Server struct {
	db        *sql.DB
	uploadDir string
	tmpl      *template.Template
}

type page struct {
	Container  *containers.Container
	Path       []containers.Container
	Containers []containers.Container
	Items      []items.Item
}

func New(db *sql.DB, uploadDir string) (*Server, error) {
	err := os.MkdirAll(uploadDir, 0o755)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.ParseFS(files, "page.html")
	if err != nil {
		return nil, err
	}

	return &Server{db: db, uploadDir: uploadDir, tmpl: tmpl}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.root)
	mux.HandleFunc("GET /c/{id}", s.container)
	mux.HandleFunc("POST /containers", s.addContainer)
	mux.HandleFunc("POST /items", s.addItem)
	mux.HandleFunc("POST /containers/{id}/delete", s.deleteContainer)
	mux.HandleFunc("POST /items/{id}/delete", s.deleteItem)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.uploadDir))))
	return mux
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	list, err := containers.ListByParent(s.db, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, page{Containers: list})
}

func (s *Server) container(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	c, err := containers.Get(s.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path, err := s.path(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	children, err := containers.ListByParent(s.db, c.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := items.ListByContainer(s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, page{Container: c, Path: path, Containers: children, Items: list})
}

func (s *Server) path(c *containers.Container) ([]containers.Container, error) {
	var path []containers.Container
	parent := c.Parent
	for parent != nil {
		p, err := containers.Get(s.db, *parent)
		if err != nil {
			return nil, err
		}
		path = append([]containers.Container{*p}, path...)
		parent = p.Parent
	}
	return path, nil
}

func (s *Server) addContainer(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}

	parent, err := optionalInt(r.FormValue("parent"))
	if err != nil {
		http.Error(w, "Invalid parent", http.StatusBadRequest)
		return
	}

	img, err := s.saveImage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := r.FormValue("description")
	c := containers.Container{
		Name:        optionalString(r.FormValue("name")),
		Description: &desc,
		ImgPath:     img,
		Parent:      parent,
	}

	_, err = containers.Insert(s.db, &c)
	if err != nil {
		s.removeImage(img)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, containerURL(parent), http.StatusSeeOther)
}

func (s *Server) addItem(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}

	containerId, err := optionalInt(r.FormValue("container"))
	if err != nil {
		http.Error(w, "Invalid container", http.StatusBadRequest)
		return
	}

	amount, err := optionalInt(r.FormValue("amount"))
	if err != nil {
		http.Error(w, "Amount must be a number", http.StatusBadRequest)
		return
	}

	img, err := s.saveImage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := r.FormValue("description")
	i := items.Item{
		ContainerId: containerId,
		Name:        optionalString(r.FormValue("name")),
		Description: &desc,
		Amount:      amount,
		ImgPath:     img,
	}

	_, err = app.AddItem(s.db, &i)
	if err != nil {
		s.removeImage(img)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, containerURL(containerId), http.StatusSeeOther)
}

func (s *Server) deleteContainer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	c, err := containers.Get(s.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	children, err := containers.ListByParent(s.db, c.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := items.ListByContainer(s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(children) > 0 || len(list) > 0 {
		http.Error(w, "Can't delete this container, empty it first", http.StatusConflict)
		return
	}

	err = containers.Remove(s.db, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.removeImage(c.ImgPath)
	http.Redirect(w, r, containerURL(c.Parent), http.StatusSeeOther)
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	i, err := items.Get(s.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = items.Remove(s.db, i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.removeImage(i.ImgPath)
	http.Redirect(w, r, containerURL(i.ContainerId), http.StatusSeeOther)
}

func (s *Server) render(w http.ResponseWriter, p page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.tmpl.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) saveImage(r *http.Request) (*string, error) {
	file, _, err := r.FormFile("image")
	if errors.Is(err, http.ErrMissingFile) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}

	ext, ok := imageTypes[http.DetectContentType(head[:n])]
	if !ok {
		return nil, errors.New("Image must be jpeg, png, gif or webp")
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	random := make([]byte, 16)
	rand.Read(random)
	name := hex.EncodeToString(random) + ext

	out, err := os.Create(filepath.Join(s.uploadDir, name))
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		os.Remove(out.Name())
		return nil, err
	}

	return &name, nil
}

func (s *Server) removeImage(name *string) {
	if name == nil {
		return
	}
	os.Remove(filepath.Join(s.uploadDir, filepath.Base(*name)))
}

func parseForm(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	err := r.ParseMultipartForm(maxUpload)
	if err != nil {
		http.Error(w, "Upload too large or invalid form", http.StatusBadRequest)
		return false
	}
	return true
}

func optionalInt(v string) (*int, error) {
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func containerURL(id *int) string {
	if id == nil {
		return "/"
	}
	return fmt.Sprintf("/c/%d", *id)
}
