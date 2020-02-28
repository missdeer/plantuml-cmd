package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/google/uuid"
	"github.com/missdeer/golib/fsutil"
)

var (
	remoteService   bool
	javaPath        string
	jarPath         string
	dotPath         string
	outputFormat    string
	outputPath      string
	outputDirectory string
	inputType       string
	sourceFile      string
	serviceURL      string
)

func main() {
	jarPath = os.Getenv(`PLANTUML_PATH`)
	if b, err := fsutil.FileExists(jarPath); err != nil || !b {
		jarPath, _ = exec.LookPath("plantuml.jar")
	}

	dotName := `dot`
	javaName := `java`
	if runtime.GOOS == "windows" {
		dotName = `dot.exe`
		javaName = `java.exe`
	}
	dotPath = os.Getenv(`GRAPHVIZ_DOT`)
	if b, err := fsutil.FileExists(dotPath); err != nil || !b {
		dotPath, _ = exec.LookPath(dotName)
	}

	javaPath = os.Getenv(`JAVA_PATH`)
	if b, err := fsutil.FileExists(javaPath); err != nil || !b {
		javaHome := os.Getenv(`JAVA_HOME`)
		javaPath = filepath.Join(javaHome, "bin", javaName)
	}
	if b, err := fsutil.FileExists(javaPath); err != nil || !b {
		jreHome := os.Getenv(`JRE_HOME`)
		javaPath = filepath.Join(jreHome, "bin", javaName)
	}
	if b, err := fsutil.FileExists(javaPath); err != nil || !b {
		javaPath, _ = exec.LookPath(javaName)
	}

	flag.StringVarP(&javaPath, "java", "j", javaPath, "set local java.exe path")
	flag.StringVarP(&jarPath, "jar", "a", jarPath, "set local plantuml.jar path")
	flag.StringVarP(&dotPath, "dot", "d", dotPath, "set local dot.exe path")
	flag.StringVarP(&outputFormat, "format", "f", "svg", "set output format, png or svg")
	flag.StringVarP(&outputPath, "output", "o", "", "save output file to local path, will ignore path option, if it's not set, will generate a uuid as the file name")
	flag.StringVarP(&outputDirectory, "path", "p", ".", "save output file to local directory path")
	flag.StringVarP(&sourceFile, "input", "i", "", "input source file path, if it's empty, then read source from stdin")
	flag.StringVarP(&inputType, "type", "t", "uml", "set input type, uml/ditto/mindmap/math/latex/dot/gantt")
	flag.BoolVarP(&remoteService, "remote", "r", false, "use remote PlantUML service, must set service URL")
	flag.StringVarP(&serviceURL, "service", "s", "https://www.plantuml.com/plantuml", "set remote PlantUML service url")
	flag.Parse()

	if outputFormat != "svg" && outputFormat != "png" {
		log.Fatal("invalid output format")
	}

	if remoteService && serviceURL == "" {
		log.Fatal("invalid remote service URL")
	}

	input := []string{}
	var fh *os.File
	if b, err := fsutil.FileExists(sourceFile); err == nil && b == true {
		file, err := os.Open(sourceFile)
		if err != nil {
			log.Println(err)
		} else {
			fh = file
			defer file.Close()
		}
	}

	var scanner *bufio.Scanner
	if fh != nil {
		scanner = bufio.NewScanner(fh)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}
	for scanner.Scan() {
		input = append(input, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	content := strings.Join(input, "\n")

	if !strings.HasPrefix(content, "@start"+inputType) &&
		!strings.HasPrefix(content, "@startuml") {
		content = fmt.Sprintf("@start%s\n%s", inputType, content)
	}
	if !strings.HasSuffix(content, "@end"+inputType) &&
		!strings.HasSuffix(content, "@enduml") {
		content = content + "\n@end" + inputType
	}

	output, err := plantuml(content, outputFormat)
	if err != nil {
		log.Fatal(err)
	}
	if outputPath != "" {
		d := filepath.Dir(outputPath)
		if b, err := fsutil.FileExists(d); b == false || err != nil {
			os.MkdirAll(d, 0644)
		}
		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f.Write(output)
		return
	}
	if outputDirectory != "" {
		if b, err := fsutil.FileExists(outputDirectory); b == false || err != nil {
			os.MkdirAll(outputDirectory, 0644)
		}
		id, err := uuid.NewUUID()
		if err != nil {
			log.Fatal(err)
		}
		fn := filepath.Join(outputDirectory, id.String()+"."+outputFormat)
		f, err := os.Create(fn)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f.Write(output)
		return
	}
}

// GetBytes returns content as []byte
func getBytes(u string, headers http.Header, timeout time.Duration, retryCount int) (c []byte, err error) {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	retry := 0
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("Could not parse novel page request:", err)
		return
	}

	req.Header = headers
doRequest:
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send request:", err)
		retry++
		if retry < retryCount {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		log.Println("response not 200:", resp.StatusCode, resp.Status)
		retry++
		if retry < retryCount {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	c, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Println("reading content failed")
		retry++
		if retry < retryCount {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	return
}

func plantumlRemote(content, format string) (b []byte, e error) {
	u := fmt.Sprintf("%s/%s/%s", serviceURL, format, Encode(content))

	return getBytes(u,
		http.Header{
			"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:62.0) Gecko/20100101 Firefox/62.0"},
			"Accept":     []string{"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
		},
		30*time.Second, 3)
}

func plantumlLocal(content, format string) (b []byte, e error) {
	if b, e := fsutil.FileExists(javaPath); e != nil || !b {
		return nil, errors.New("invalid java.exe path")
	}
	if b, e := fsutil.FileExists(jarPath); e != nil || !b {
		return nil, errors.New("invalid plantuml.jar path")
	}
	args := []string{`-Djava.awt.headless=true`, "-jar", jarPath, "-t" + format, "-charset", "UTF-8"}
	if b, _ := fsutil.FileExists(dotPath); b {
		args = append(args, "-graphvizdot", dotPath)
	}

	args = append(args, "-pipe")
	cmd := exec.Command(javaPath, args...)
	stdin, e := cmd.StdinPipe()
	if e != nil {
		return nil, e
	}
	go func() {
		io.WriteString(stdin, content)
		stdin.Close()
	}()

	b, e = cmd.Output()
	if len(b) > 0 {
		e = nil
		index := bytes.Index(b, []byte(`</svg><?xml`))
		if index > 0 {
			b = b[:index+len(`</svg>`)]
		}
	}

	return b, e
}

func plantuml(content, format string) (b []byte, e error) {
	if remoteService {
		b, e = plantumlRemote(content, format)
	} else {
		b, e = plantumlLocal(content, format)
	}

	if e != nil {
		return
	}

	if format == "svg" {
		beginPos := bytes.Index(b, []byte("style=\"width:"))
		if beginPos > 0 {
			t := b[beginPos+1:]
			endPos := bytes.Index(t, []byte(";\""))
			if endPos > 0 {
				b = append(b[:beginPos-1], t[endPos+2:]...)
			}
		}
	}
	return
}
