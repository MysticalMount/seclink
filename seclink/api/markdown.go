package api

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func parseMarkdownFile(content []byte) (Page, error) {
	sections := strings.SplitN(string(content), "---", 2)
	if len(sections) < 2 {
		return Page{}, errors.New("invalid markdown format")
	}

	metadata := sections[0]
	mdContent := sections[1]

	// deal with rogue \r's
	metadata = strings.ReplaceAll(metadata, "\r", "")
	mdContent = strings.ReplaceAll(mdContent, "\r", "")

	title, slug, parent, description, metaDescriptionStr,
		metaPropertyTitleStr, metaPropertyDescriptionStr := parseMetadata(metadata)

	htmlContent := mdToHTML([]byte(mdContent))
	headers := extractHeaders([]byte(mdContent))

	return Page{
		Title:                   title,
		Slug:                    slug,
		Parent:                  parent,
		Description:             description,
		Content:                 htmlContent,
		Headers:                 headers,
		MetaDescription:         metaDescriptionStr,
		MetaPropertyTitle:       metaPropertyTitleStr,
		MetaPropertyDescription: metaPropertyDescriptionStr,
	}, nil
}

func extractHeaders(content []byte) []string {
	var headers []string
	//match only level 2 markdown headers
	re := regexp.MustCompile(`(?m)^##\s+(.*)`)
	matches := re.FindAllSubmatch(content, -1)

	for _, match := range matches {
		// match[1] contains header text without the '##'
		headers = append(headers, string(match[1]))
	}

	return headers
}

func mdToHTML(md []byte) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}
	renderer := html.NewRenderer(opts)

	doc := parser.Parse(md)

	output := markdown.Render(doc, renderer)

	return string(output)
}

func parseMetadata(metadata string) (
	title string,
	slug string,
	parent string,
	description string,
	metaDescription string,
	metaPropertyTitle string,
	metaPropertyDescription string,
) {
	re := regexp.MustCompile(`(?m)^(\w+):\s*(.+)`)
	matches := re.FindAllStringSubmatch(metadata, -1)

	metaDataMap := make(map[string]string)
	for _, match := range matches {
		if len(match) == 3 {
			metaDataMap[match[1]] = match[2]
		}
	}

	title = metaDataMap["Title"]
	slug = metaDataMap["Slug"]
	parent = metaDataMap["Parent"]
	description = metaDataMap["Description"]
	metaDescriptionStr := metaDataMap["MetaDescription"]
	metaPropertyTitleStr := metaDataMap["MetaPropertyTitle"]
	metaPropertyDescriptionStr := metaDataMap["MetaPropertyDescription"]

	return title, slug, parent, description, metaDescriptionStr,
		metaPropertyTitleStr, metaPropertyDescriptionStr
}
