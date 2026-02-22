package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

type OutputFormatter interface {
	Format(any) error
}

type TableFormatter struct {
	writer *tabwriter.Writer
}

func NewTableFormatter() *TableFormatter {
	return &TableFormatter{
		writer: tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0),
	}
}

func (f *TableFormatter) Format(data any) error {
	switch v := data.(type) {
	case *ShowOutput:
		return f.formatShow(v)
	case *VersionsOutput:
		return f.formatVersions(v)
	case *LatestOutput:
		return f.formatLatest(v)
	case *BrowseOutput:
		return f.formatBrowse(v)
	case *ContainerShowOutput:
		return f.formatContainerShow(v)
	case *ContainerLatestOutput:
		return f.formatContainerLatest(v)
	default:
		return fmt.Errorf("unsupported output type")
	}
}

func (f *TableFormatter) formatShow(data *ShowOutput) error {
	fmt.Fprintf(f.writer, "Name:\t%s\n", data.Name)
	fmt.Fprintf(f.writer, "Version:\t%s\n", data.Version)
	if data.Description != "" {
		fmt.Fprintf(f.writer, "Description:\t%s\n", data.Description)
	}
	if data.Author != "" {
		fmt.Fprintf(f.writer, "Author:\t%s", data.Author)
		if data.AuthorEmail != "" {
			fmt.Fprintf(f.writer, " <%s>", data.AuthorEmail)
		}
		fmt.Fprintln(f.writer)
	}
	if data.License != "" {
		fmt.Fprintf(f.writer, "License:\t%s\n", data.License)
	}
	if data.HomePage != "" {
		fmt.Fprintf(f.writer, "Home Page:\t%s\n", data.HomePage)
	}
	if len(data.Dependencies) > 0 {
		fmt.Fprintf(f.writer, "Dependencies:\t")
		for i, dep := range data.Dependencies {
			if i > 0 {
				fmt.Fprint(f.writer, ", ")
			}
			fmt.Fprint(f.writer, dep)
		}
		fmt.Fprintln(f.writer)
	}
	return f.writer.Flush()
}

func (f *TableFormatter) formatVersions(data *VersionsOutput) error {
	fmt.Fprintln(f.writer, "Version\tUpload Date")
	fmt.Fprintln(f.writer, "-------\t-----------")
	for _, v := range data.Versions {
		fmt.Fprintf(f.writer, "%s\t%s\n", v.Version, v.UploadDate)
	}
	return f.writer.Flush()
}

func (f *TableFormatter) formatLatest(data *LatestOutput) error {
	fmt.Fprintf(f.writer, "Latest Version:\t%s\n", data.Version)
	return f.writer.Flush()
}

func (f *TableFormatter) formatBrowse(data *BrowseOutput) error {
	fmt.Fprintf(f.writer, "Opening:\t%s\n", data.URL)
	return f.writer.Flush()
}

func (f *TableFormatter) formatContainerShow(data *ContainerShowOutput) error {
	fmt.Fprintf(f.writer, "Name:\t%s\n", data.Name)
	if data.Description != "" {
		fmt.Fprintf(f.writer, "Description:\t%s\n", data.Description)
	}
	if data.Tag != "" {
		fmt.Fprintf(f.writer, "Tag:\t%s\n", data.Tag)
	}
	if data.Digest != "" {
		fmt.Fprintf(f.writer, "Digest:\t%s\n", data.Digest)
	}
	if data.TagDate != "" {
		fmt.Fprintf(f.writer, "Tag Date:\t%s\n", data.TagDate)
	}
	if data.TagSize != "" {
		fmt.Fprintf(f.writer, "Tag Size:\t%s\n", data.TagSize)
	}
	fmt.Fprintf(f.writer, "Registry:\t%s\n", data.Registry)
	fmt.Fprintf(f.writer, "Image:\t%s\n", data.FullImageRef)
	return f.writer.Flush()
}

func (f *TableFormatter) formatContainerLatest(data *ContainerLatestOutput) error {
	fmt.Fprintf(f.writer, "Latest Tag:\t%s\n", data.Version)
	return f.writer.Flush()
}

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Format(data any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
