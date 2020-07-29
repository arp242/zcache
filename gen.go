// +build go_run_only

package main

import (
	"fmt"
	"log"
	"strings"
	"text/template"
)

var tpl = template.Must(template.New("").Parse(`
// {{if .Inc}}Increment{{else}}Decrement{{end}} an item of type {{.T}} by n. Returns an error if the item's value is
// not an {{.T}}, or if it was not found. If there is no error, the new value is returned.
func (c *cache) {{if .Inc}}Increment{{else}}Decrement{{end}}{{.U}}(k string, n {{.T}}) ({{.T}}, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.{{if .Inc}}Increment{{else}}Decrement{{end}}: item" + k + " not found")
	}
	rv, ok := v.Object.({{.T}})
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an {{.T}}")
	}
	nv := rv {{if .Inc}}+{{else}}-{{end}} n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}
`))

var tplFloat = template.Must(template.New("").Parse(`
// {{if .Inc}}Increment{{else}}Decrement{{end}} an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to {{if .Inc}}increment{{else}}decrement{{end}} it by n.
// To retrieve the {{if .Inc}}incremented{{else}}decremented{{end}} value, use one of the specialized methods,
// e.g. {{if .Inc}}Increment{{else}}Decrement{{end}}Float64.
func (c *cache) {{if .Inc}}Increment{{else}}Decrement{{end}}Float(k string, n float64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return errors.New("zcache.{{if .Inc}}Increment{{else}}Decrement{{end}}Float: item " + k + " not found")
	}
	switch v.Object.(type) {
	case float32:
		v.Object = v.Object.(float32) {{if .Inc}}+{{else}}-{{end}} float32(n)
	case float64:
		v.Object = v.Object.(float64) {{if .Inc}}+{{else}}-{{end}} n
	default:
		c.mu.Unlock()
		return errors.New("zcache.{{if .Inc}}Increment{{else}}Decrement{{end}}Float: the value for " + k + " does not have type float32 or float64")
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}
`))

func main() {
	types := []string{"int", "int8", "int16", "int32", "int64", "uint",
		"uintptr", "uint8", "uint16", "uint32", "uint64", "float32", "float64"}

	out := new(strings.Builder)
	out.WriteString("package cache\n\nimport \"errors\"\n")
	err := tplFloat.Execute(out, struct {
		Inc bool
	}{true})
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range types {
		u := strings.ToUpper(t[:1]) + t[1:]
		err := tpl.Execute(out, struct {
			T, U string
			Inc  bool
		}{t, u, true})
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tplFloat.Execute(out, struct {
		Inc bool
	}{false})
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range types {
		u := strings.ToUpper(t[:1]) + t[1:]
		err := tpl.Execute(out, struct {
			T, U string
			Inc  bool
		}{t, u, false})
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println(strings.TrimSpace(out.String()))
}
