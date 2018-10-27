// Package utl (convert.go) :
// This is a convereter to send GAS script to Google.
package utl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli"
)

// const :
const (
	deffuncwith = "main"
)

// chkScript : Check the imported script.
func chkScript(c *cli.Context) (string, error) {
	var scriptfile string
	fext := filepath.Ext(c.String("scriptfile"))
	if fext == ".coffee" {
		var cmd *exec.Cmd
		cmd = exec.Command("coffee", "-cb", c.String("scriptfile"))
		if err := cmd.Run(); err != nil {
			return "", err
		}
		scriptfile = strings.Replace(filepath.Base(c.String("scriptfile")), fext, ".js", -1)
	} else {
		scriptfile = c.String("scriptfile")
	}
	return scriptfile, nil
}

// ConvGasToPut : Reads GAS source and formats it.
func ConvGasToPut(c *cli.Context) string {
	scriptfile, err := chkScript(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Compile error of CoffeeScript. Please check the script.\n%s\n", err.Error())
		os.Exit(1)
	}
	var res string
	if len(scriptfile) > 0 {
		fp, err := os.Open(scriptfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Script '%s' is not found. ", scriptfile)
			os.Exit(1)
		}
		defer fp.Close()
		scripts := []string{}
		s := bufio.NewScanner(fp)
		for s.Scan() {
			dat := s.Text()
			dat += "\n"
			scripts = append(scripts, dat)
		}
		if s.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error: %v .", s.Err())
			os.Exit(1)
		}
		mem := make([]byte, 0, 100)
		for _, v := range scripts {
			mem = append(mem, v...)
		}
		res = string(mem)
	} else {
		res = ""
	}
	return res
}

// ConvGasToRun : Reads GAS source and formats it.
func ConvGasToRun(c *cli.Context) string {
	scriptfile, err := chkScript(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Compile error of CoffeeScript. Please check the script.\n%s\n", err.Error())
		os.Exit(1)
	}
	senddata := c.String("value")
	var res string
	if len(scriptfile) > 0 {
		fp, err := os.Open(scriptfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Script '%s' is not found. ", scriptfile)
			os.Exit(1)
		}
		defer fp.Close()
		scripts := []string{}
		s := bufio.NewScanner(fp)
		for s.Scan() {
			dat := s.Text()
			if strings.Contains(dat, "//") {
				if !strings.Contains(dat, "://") || strings.Contains(dat, "// ") || strings.Contains(dat, " //") {
					dat = dat[:strings.Index(dat, "//")]
				}
			}
			if dat != "" {
				conv := strings.Replace(strings.TrimSpace(dat), "\\", "\\\\", -1)
				conv = strings.Replace(strings.TrimSpace(conv), "\"", "\\\"", -1)
				conv = strings.Replace(conv, "'", "\\'", -1)
				scripts = append(scripts, conv)
			}
		}
		if s.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error: %v .", s.Err())
			os.Exit(1)
		}
		mem := make([]byte, 0, 100)
		for _, v := range scripts {
			mem = append(mem, v...)
		}
		st := string(mem)
		if len(senddata) > 0 {
			if regexp.MustCompile("^[+-]?[0-9]*[\\.]?[0-9]+$").Match([]byte(senddata)) {
				st += "var defdata=" + senddata + ";"
			} else if regexp.MustCompile("^\\[|\\]$").Match([]byte(senddata)) || regexp.MustCompile("^{|}$").Match([]byte(senddata)) {
				dat := strings.Replace(strings.TrimSpace(senddata), "\"", "\\\"", -1)
				senddata = strings.Replace(dat, "'", "\\'", -1)
				st += "var defdata=" + senddata + ";"
			} else {
				if regexp.MustCompile("^\"|\"$").Match([]byte(senddata)) || regexp.MustCompile("^\\'|\\'$").Match([]byte(senddata)) {
					dat := strings.Replace(strings.TrimSpace(senddata), "\"", "\\\"", -1)
					senddata = strings.Replace(dat, "'", "\\'", -1)
					st += "var defdata=" + senddata + ";"
				} else {
					st += "var defdata=\\\"" + senddata + "\\\";"
				}
			}
			st += deffuncwith + "(defdata)"
		} else {
			st += deffuncwith + "()"
		}
		res = "\"" + st + "\""
	} else {
		res = ""
	}
	return res
}

// ConvStringToRun : Reads GAS source and formats it as a string script.
func ConvStringToRun(c *cli.Context, stringscript string) string {
	senddata := c.String("value")
	var res string
	if len(stringscript) > 0 {
		scripts := []string{}
		s := bufio.NewScanner(strings.NewReader(stringscript))
		for s.Scan() {
			dat := s.Text()
			if strings.Contains(dat, "//") {
				if !strings.Contains(dat, "://") || strings.Contains(dat, "// ") || strings.Contains(dat, " //") {
					dat = dat[:strings.Index(dat, "//")]
				}
			}
			if dat != "" {
				conv := strings.Replace(strings.TrimSpace(dat), "\\", "\\\\", -1)
				conv = strings.Replace(strings.TrimSpace(conv), "\"", "\\\"", -1)
				conv = strings.Replace(conv, "'", "\\'", -1)
				scripts = append(scripts, conv)
			}
		}
		if s.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error: %v .", s.Err())
			os.Exit(1)
		}
		mem := make([]byte, 0, 100)
		for _, v := range scripts {
			mem = append(mem, v...)
		}
		st := string(mem)
		if len(senddata) > 0 {
			if regexp.MustCompile("^[+-]?[0-9]*[\\.]?[0-9]+$").Match([]byte(senddata)) {
				st += "var defdata=" + senddata + ";"
			} else if regexp.MustCompile("^\\[|\\]$").Match([]byte(senddata)) || regexp.MustCompile("^{|}$").Match([]byte(senddata)) {
				dat := strings.Replace(strings.TrimSpace(senddata), "\"", "\\\"", -1)
				senddata = strings.Replace(dat, "'", "\\'", -1)
				st += "var defdata=" + senddata + ";"
			} else {
				if regexp.MustCompile("^\"|\"$").Match([]byte(senddata)) || regexp.MustCompile("^\\'|\\'$").Match([]byte(senddata)) {
					dat := strings.Replace(strings.TrimSpace(senddata), "\"", "\\\"", -1)
					senddata = strings.Replace(dat, "'", "\\'", -1)
					st += "var defdata=" + senddata + ";"
				} else {
					st += "var defdata=\\\"" + senddata + "\\\";"
				}
			}
			st += deffuncwith + "(defdata)"
		} else {
			st += deffuncwith + "()"
		}
		res = "\"" + st + "\""
	} else {
		res = ""
	}
	return res
}

// ConvGasToUpload :
func ConvGasToUpload(scriptfile string) string {
	var res string
	if len(scriptfile) > 0 {
		fp, err := os.Open(scriptfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Script '%s' is not found. ", scriptfile)
			os.Exit(1)
		}
		defer fp.Close()
		scripts := []string{}
		s := bufio.NewScanner(fp)
		for s.Scan() {
			dat := s.Text()
			dat += "\n"
			scripts = append(scripts, dat)
		}
		if s.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error: %v .", s.Err())
			os.Exit(1)
		}
		mem := make([]byte, 0, 100)
		for _, v := range scripts {
			mem = append(mem, v...)
		}
		res = string(mem)
	} else {
		res = ""
	}
	return res
}
