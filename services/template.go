package services

import (
	"log"
	"path/filepath"
	"sync"
	"text/template"
	"time"
)

type TemplateService struct {
	templates map[string]*template.Template
	location  *time.Location
	mu        sync.RWMutex
}

func NewTemplateService(location *time.Location) *TemplateService {
	return &TemplateService{
		templates: make(map[string]*template.Template),
		location:  location,
	}
}

// GetTemplate 按需加载、缓存并返回模板
func (s *TemplateService) GetTemplate(templatePath string) (*template.Template, error) {
	s.mu.RLock()
	tmpl, ok := s.templates[templatePath]
	s.mu.RUnlock()
	if ok {
		return tmpl, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// 再次检查以防并发加载
	if tmpl, ok = s.templates[templatePath]; ok {
		return tmpl, nil
	}

	// 加载新模板
	templateName := filepath.Base(templatePath)
	newTmpl, err := template.New(templateName).Funcs(template.FuncMap{
		"getCSTtime": s.getCSTtime,
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}).ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	s.templates[templatePath] = newTmpl
	log.Printf("模板 %s 加载成功", templatePath)
	return newTmpl, nil
}

func (s *TemplateService) getCSTtime(t time.Time) string {
	return t.In(s.location).Format("2006-01-02 15:04:05")
}

func SetTimezone(timezone string) (*time.Location, error) {
	return time.LoadLocation(timezone)
}
