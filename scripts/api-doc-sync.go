//go:build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type route struct {
	Method   string
	Path     string
	Handler  string
	Tag      string
	Auth     bool
	Source   string
	Position int
}

func main() {
	input := flag.String("input", "doc/Pilotwave.v1.yaml", "OpenAPI YAML input path")
	output := flag.String("output", "doc/Pilotwave.v1.yaml", "OpenAPI YAML output path")
	check := flag.Bool("check", false, "fail if generated output differs from output path")
	flag.Parse()

	routes, err := discoverRoutes("pkg/http_server/api")
	if err != nil {
		exitErr(err)
	}

	generated, err := syncOpenAPI(*input, routes)
	if err != nil {
		exitErr(err)
	}

	if *check {
		current, err := os.ReadFile(*output)
		if err != nil {
			exitErr(err)
		}
		if string(current) != string(generated) {
			exitErr(fmt.Errorf("%s is out of date; run make api-docs-sync", *output))
		}
		return
	}

	if err := os.WriteFile(*output, generated, 0644); err != nil {
		exitErr(err)
	}
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func discoverRoutes(dir string) ([]route, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}

	routeRe := regexp.MustCompile(`api\.router\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"(.*)`)
	handlerRe := regexp.MustCompile(`api\.([A-Za-z0-9_]+)\)`)
	var routes []route

	for _, file := range files {
		base := filepath.Base(file)
		if base == "swagger.go" {
			continue
		}

		bodyBytes, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		body := string(bodyBytes)

		tag := strings.TrimSuffix(base, ".go")
		lines := strings.Split(body, "\n")
		for i, line := range lines {
			matches := routeRe.FindStringSubmatch(line)
			if len(matches) == 0 {
				continue
			}

			method := strings.ToLower(matches[1])
			path := normalizeRoutePath(matches[2])
			tail := matches[3]
			handler := ""
			if handlerMatches := handlerRe.FindStringSubmatch(tail); len(handlerMatches) > 1 {
				handler = handlerMatches[1]
			}

			routes = append(routes, route{
				Method:   method,
				Path:     path,
				Handler:  handler,
				Tag:      tag,
				Auth:     strings.Contains(line, "RequiredAuth"),
				Source:   file,
				Position: i + 1,
			})
		}
	}

	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].Path != routes[j].Path {
			return routes[i].Path < routes[j].Path
		}
		return routes[i].Method < routes[j].Method
	})

	return routes, nil
}

func normalizeRoutePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/api/v1")
	if path == "" {
		path = "/"
	}

	paramRe := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	return paramRe.ReplaceAllString(path, `{$1}`)
}

func syncOpenAPI(input string, routes []route) ([]byte, error) {
	inputBytes, err := os.ReadFile(input)
	if err != nil {
		return nil, err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(inputBytes, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s does not contain a YAML mapping document", input)
	}

	root := doc.Content[0]
	paths := mappingValue(root, "paths")
	if paths == nil {
		paths = mappingNode()
		setMappingValue(root, "paths", paths)
	}

	for _, route := range routes {
		pathNode := mappingValue(paths, route.Path)
		if pathNode == nil {
			pathNode = mappingNode()
			setMappingValue(paths, route.Path, pathNode)
		}

		if mappingValue(pathNode, route.Method) == nil {
			setMappingValue(pathNode, route.Method, operationNode(route))
		}
	}

	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(&doc); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func operationNode(route route) *yaml.Node {
	op := mappingNode()
	setScalar(op, "summary", summary(route))
	setScalar(op, "operationId", operationID(route))
	setMappingValue(op, "tags", sequenceOfScalars([]string{route.Tag}))

	if params := pathParameters(route.Path); len(params) > 0 {
		paramNodes := sequenceNode()
		for _, name := range params {
			paramNodes.Content = append(paramNodes.Content, parameterNode(name))
		}
		setMappingValue(op, "parameters", paramNodes)
	}

	if route.Auth {
		security := sequenceNode()
		securityItem := mappingNode()
		setMappingValue(securityItem, "Authentication", sequenceNode())
		security.Content = append(security.Content, securityItem)
		setMappingValue(op, "security", security)
	}

	if route.Method == "post" || route.Method == "put" || route.Method == "patch" {
		setMappingValue(op, "requestBody", requestBodyNode())
	}

	responses := mappingNode()
	setMappingValue(responses, "200", responseNode("OK"))
	setMappingValue(responses, "default", responseNode("Error"))
	setMappingValue(op, "responses", responses)

	return op
}

func summary(route route) string {
	if route.Handler == "" {
		return strings.Title(route.Method) + " " + route.Path
	}
	return splitWords(route.Handler)
}

func operationID(route route) string {
	parts := []string{route.Method}
	for _, part := range strings.Split(strings.Trim(route.Path, "/"), "/") {
		part = strings.Trim(part, "{}")
		if part == "" {
			continue
		}
		parts = append(parts, strings.ToLower(part))
	}
	return strings.Join(parts, "-")
}

func splitWords(value string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	value = re.ReplaceAllString(value, `${1} ${2}`)
	return strings.TrimSpace(value)
}

func pathParameters(path string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)
	params := make([]string, 0, len(matches))
	for _, match := range matches {
		params = append(params, match[1])
	}
	return params
}

func parameterNode(name string) *yaml.Node {
	param := mappingNode()
	setScalar(param, "name", name)
	setScalar(param, "in", "path")
	setScalar(param, "required", "true")
	setScalar(param, "description", name)
	schema := mappingNode()
	setScalar(schema, "type", "string")
	setMappingValue(param, "schema", schema)
	return param
}

func requestBodyNode() *yaml.Node {
	requestBody := mappingNode()
	content := mappingNode()
	media := mappingNode()
	schema := mappingNode()
	setScalar(schema, "type", "object")
	setMappingValue(media, "schema", schema)
	setMappingValue(content, "application/json", media)
	setMappingValue(requestBody, "content", content)
	return requestBody
}

func responseNode(description string) *yaml.Node {
	resp := mappingNode()
	setScalar(resp, "description", description)
	content := mappingNode()
	media := mappingNode()
	schema := mappingNode()
	setScalar(schema, "type", "object")
	setMappingValue(media, "schema", schema)
	setMappingValue(content, "application/json", media)
	setMappingValue(resp, "content", content)
	return resp
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func setScalar(node *yaml.Node, key string, value string) {
	setMappingValue(node, key, scalarNode(value))
}

func setMappingValue(node *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content[i+1] = value
			return
		}
	}
	node.Content = append(node.Content, scalarNode(key), value)
}

func mappingNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode}
}

func sequenceNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode}
}

func sequenceOfScalars(values []string) *yaml.Node {
	seq := sequenceNode()
	for _, value := range values {
		seq.Content = append(seq.Content, scalarNode(value))
	}
	return seq
}

func scalarNode(value string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.ScalarNode, Value: value}
	if value == "true" {
		node.Tag = "!!bool"
	}
	return node
}
