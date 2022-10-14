package ustripe

import (
	"fmt"
	"os/exec"
	"bytes"
	"strings"
	"regexp"
)

// PasswordHash calculates the MD5 BSD hash.
func PasswordHash(password string) (hash string, err error) {
	cmd := exec.Command("openssl", "passwd", "-1", "-salt", "pstripe", "-stdin")
	out := bytes.Buffer{}
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}
	hash, _, _ = strings.Cut(out.String(), "\n")
	return
}

// PasswordCheck checks the password is usable.
func PasswordCheck(password string) (err error) {
	var cmd *exec.Cmd
	var out  bytes.Buffer
	var reg *regexp.Regexp
	
	cmd = exec.Command("cracklib-check")
	out = bytes.Buffer{}
	reg, _ = regexp.CompilePOSIX(": *OK")
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = &out

	err = cmd.Run()
	
	switch {
	case err != nil:                    return
	case out.Len() == 0:                return fmt.Errorf("password check failed")
	case len(reg.Find(out.Bytes()))!=0: return nil
	default:                            return fmt.Errorf("the password, " + out.String())	
	}
}

func SendMail(mail string) (err error) {
	cmd    := exec.Command("sh", "-e", "-c", SendmailCommand)
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdin  = strings.NewReader(mail)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	if stderr.Len() > 0 {
		err = fmt.Errorf("%s", stderr.String())
		return err
	}
	return nil	
}

func OpenBrowser(url string) (err error) {
	return exec.Command("xdg-open", url).Run()
}
