package main

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"x.io/xrpc/pkg/echo"
)

const (
	serverUA      = "Echo/1.0.0"
	fs_maxbufsize = 4096 // 4096 bits = default page size on OSX
	TemplateDir   = "./view/"
	UploadDir     = "./upload"
)

func min(x int64, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func index(c echo.Context) error {
	title := struct {
		Title string
	}{Title: "Echo文件服务器"}
	t, _ := template.ParseFiles(TemplateDir + "index.html")
	return t.Execute(c.Response(), title)
}

func uploadHandler(c echo.Context) (err error) {
	r := c.Request()
	w := c.Response()
	if r.Method == echo.GET {
		t, _ := template.ParseFiles(TemplateDir + "file.html")
		t.Execute(w, "上传文件")
		return nil
	}
	//parse the multipart form in the request
	err = r.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	if _, err := os.Open(UploadDir); err != nil {
		os.MkdirAll(UploadDir, os.ModePerm)
	}

	files := r.MultipartForm.File["uploadfile"]
	for i, _ := range files {
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		dst, err := os.Create("./upload/" + files[i].Filename)
		defer dst.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
	}
	c.String(http.StatusOK, "upload succeed")
	return nil
}

type dirlisting struct {
	Name       string
	ChildDirs  []string
	ChildFiles []string
	ServerUA   string
}

func parseCSV(data string) []string {
	splitted := strings.SplitN(data, ",", -1)
	data_tmp := make([]string, len(splitted))

	for i, val := range splitted {
		data_tmp[i] = strings.TrimSpace(val)
	}

	return data_tmp
}

func parseRange(data string) int64 {
	stop := (int64)(0)
	part := 0
	for i := 0; i < len(data) && part < 2; i = i + 1 {
		if part == 0 { // part = 0 <=> equal isn't met.
			if data[i] == '=' {
				part = 1
			}
			continue
		}
		if part == 1 { // part = 1 <=> we've met the equal, parse beginning
			if data[i] == ',' || data[i] == '-' {
				part = 2 // part = 2 <=> OK DUDE.
			} else {
				if 48 <= data[i] && data[i] <= 57 { // If it's a digit ...
					// ... convert the char to integer and add it!
					stop = (stop * 10) + (((int64)(data[i])) - 48)
				} else {
					part = 2 // Parsing error! No error needed : 0 = from start.
				}
			}
		}
	}
	return stop
}

func serveFile(c echo.Context) (err error) {
	if !strings.HasSuffix(c.Path(), "/") {
		return c.Redirect(http.StatusTemporaryRedirect, c.Path()+"/")
	}
	w := c.Response()
	req := c.Request()
	w.Header().Set("Server", serverUA)

	filepath := path.Join(*rootDir, path.Clean(c.PathParam("path")))
	f, err := os.Open(filepath)
	if err != nil {
		http.Error(w, "404 Not Found : Error while opening the file.", 404)
		return
	}
	defer f.Close()

	statinfo, err := f.Stat()
	if err != nil {
		http.Error(w, "500 Internal Error : stat() failure.", 500)
		return
	}

	if statinfo.IsDir() {
		handleDirectory(f, req, w)
		return
	}

	if (statinfo.Mode() &^ 07777) == os.ModeSocket {
		http.Error(w, "403 Forbidden : you can't access this resource.", 403)
		return
	}

	// Manages If-Modified-Since and add Last-Modified (taken from Golang code)
	if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && statinfo.ModTime().Unix() <= t.Unix() {
		w.WriteHeader(http.StatusNotModified)
		return err
	}
	w.Header().Set("Last-Modified", statinfo.ModTime().Format(http.TimeFormat))

	// Content-Type handling
	query, err := url.ParseQuery(req.URL.RawQuery)

	if err == nil && len(query["dl"]) > 0 { // The user explicitedly wanted to download the file (Dropbox style!)
		w.Header().Set("Content-Type", "application/octet-stream")
	} else {
		if mimeType := mime.TypeByExtension(path.Ext(filepath)); mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
	}

	if req.Header.Get("Range") != "" {
		startByte := parseRange(req.Header.Get("Range"))

		if startByte < statinfo.Size() {
			f.Seek(startByte, 0)
		} else {
			startByte = 0
		}

		w.Header().Set("Content-Range",
			fmt.Sprintf("bytes %d-%d/%d", startByte, statinfo.Size()-1, statinfo.Size()))
	}

	outputWriter := w.(io.Writer)
	isCompressedReply := false

	if (*usesGzip) == true && req.Header.Get("Accept-Encoding") != "" {
		encodings := parseCSV(req.Header.Get("Accept-Encoding"))

		for _, val := range encodings {
			if val == "gzip" {
				w.Header().Set("Content-Encoding", "gzip")
				outputWriter = gzip.NewWriter(w)

				isCompressedReply = true

				break
			} else if val == "deflate" {
				w.Header().Set("Content-Encoding", "deflate")
				outputWriter = zlib.NewWriter(w)

				isCompressedReply = true

				break
			}
		}
	}

	if !isCompressedReply {
		w.Header().Set("Content-Length", strconv.FormatInt(statinfo.Size(), 10))
	}

	buf := make([]byte, min(fs_maxbufsize, statinfo.Size()))
	n := 0
	for err == nil {
		n, err = f.Read(buf)
		outputWriter.Write(buf[0:n])
	}
	switch outputWriter.(type) {
	case *gzip.Writer:
		outputWriter.(*gzip.Writer).Close()
	case *zlib.Writer:
		outputWriter.(*zlib.Writer).Close()
	}

	return f.Close()
}

func handleDirectory(f *os.File, r *http.Request, w http.ResponseWriter) {
	names, _ := f.Readdir(-1)
	var childDirs, childFiles []string
	for _, val := range names {
		if val.Name()[0] == '.' {
			continue
		}

		if val.IsDir() {
			childDirs = append(childDirs, val.Name())
		} else {
			childFiles = append(childFiles, val.Name())
		}
	}

	tpl, err := template.New("tpl").Parse(dirTpl)
	if err != nil {
		http.Error(w, "500 Internal Error : Error while generating directory listing.", 500)
		fmt.Println(err)
		return
	}

	data := dirlisting{Name: r.URL.Path, ServerUA: serverUA,
		ChildDirs: childDirs, ChildFiles: childFiles}

	err = tpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
	return
}

const dirTpl = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
<!-- Modified from lighttpd directory listing -->
<head>
<title>Index of {{.Name}}</title>
<style type="text/css">
a, a:active {text-decoration: none; color: blue;}
a:visited {color: #48468F;}
a:hover, a:focus {text-decoration: underline; color: red;}
body {background-color: #F5F5F5;}
h2 {margin-bottom: 12px;}
table {margin-left: 12px;}
th, td { font: 90% monospace; text-align: left;}
th { font-weight: bold; padding-right: 14px; padding-bottom: 3px;}
td {padding-right: 14px;}
td.s, th.s {text-align: right;}
div.list { background-color: white; border-top: 1px solid #646464; border-bottom: 1px solid #646464; padding-top: 10px; padding-bottom: 14px;}
div.foot { font: 90% monospace; color: #787878; padding-top: 4px;}
</style>
</head>
<body>
<h2>Index of {{.Name}}</h2>
<div class="list">
<table summary="Directory Listing" cellpadding="0" cellspacing="0">
<thead><tr><th class="n">Name</th><th class="t">Type</th><th class="dl">Options</th></tr></thead>
<tbody>
<tr><td class="n"><a href="../">Parent Directory</a>/</td><td class="t">Directory</td><td class="dl"></td></tr>
{{range .ChildDirs}}
<tr><td class="n"><a href="{{.}}/">{{.}}/</a></td><td class="t">Directory</td><td class="dl"></td></tr>
{{end}}
{{range .ChildFiles}}
<tr><td class="n"><a href="{{.}}">{{.}}</a></td><td class="t">&nbsp;</td><td class="dl"><a href="{{.}}?dl">Download</a></td></tr>
{{end}}
</tbody>
</table>
</div>
<div class="foot">{{.ServerUA}}</div>
</body>
</html>`
