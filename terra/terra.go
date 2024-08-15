package terra

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"slices"

)

type (
	Config struct {
		Envs []string
		IsReusableAvailable bool
		ReusablesFileSource string
		ReusablesFrom string
		ReusablesTo string
		IsCommonAvailable bool
		CommonTo string
		CommonFrom string
		IsTemplateAvailable bool
		TemplateTo string
		TemplateFrom string
	}

	Terra interface {
		Build() error
	}

	terra struct {
		env string
		config Config
	}
)

func New(env string, config Config) Terra {

	if config.ReusablesFileSource == "" {
		config.ReusablesFileSource = "./environments/%s/.reusables"
	}

	if config.ReusablesFrom == "" {
		config.ReusablesFrom = "./reusables/%s"
	}

	if config.ReusablesTo == "" {
		config.ReusablesTo = "./environments/%s/%s"
	}

	if config.CommonTo == "" {
		config.CommonTo = "./environments/%s"
	}

	if config.TemplateTo == "" {
		config.TemplateTo = "./environments/%s/templates"
	}

	if config.TemplateFrom == "" {
		config.TemplateFrom = "./templates"
	}

	if config.CommonFrom == "" {
		config.CommonFrom = "./common"
	}

	return &terra{
		env: env,
		config: config,
	}
}

func (t *terra) Build() error {

	if v := slices.Index(t.config.Envs, t.env); v < 0 {
		return &InvalidEnvError{t.env}
	}

	if t.config.IsReusableAvailable {
		fmt.Println("Copy Reusable")
		if err := t.copyReusables(); err != nil {
			return err
		}
	}
	
	if t.config.IsCommonAvailable {
		fmt.Println("Copy common")
		if err := os.CopyFS(fmt.Sprintf(t.config.CommonTo, t.env), os.DirFS(t.config.CommonFrom)); err != nil {
			return fmt.Errorf("template copy error: %v", err)
		}
	}

	if t.config.IsTemplateAvailable {
		fmt.Println("Copy templates")
		if err := os.CopyFS(fmt.Sprintf(t.config.TemplateTo, t.env), os.DirFS(t.config.TemplateFrom)); err != nil {
			return fmt.Errorf("template copy error: %v", err)
		}
	}
	return nil
}

func (t *terra) copyReusables() error { 
	f, err := os.OpenFile(fmt.Sprintf(t.config.ReusablesFileSource, t.env), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if sc.Text() == "" {
			continue
		}
		fmt.Printf("Copy %s\n", sc.Text())
		if err := copyFile(fmt.Sprintf( t.config.ReusablesFrom, sc.Text()), fmt.Sprintf( t.config.ReusablesTo, t.env, sc.Text())); err != nil {
			return fmt.Errorf("file copy error: %v", err)
		}
	}

	if err := sc.Err(); err != nil {
		return fmt.Errorf("scan file error: %v", err)
	}

	return nil
}


func copyFile(src, dst string) (err error) {
    sfi, err := os.Stat(src)
    if err != nil {
        return
    }
    if !sfi.Mode().IsRegular() {
        // cannot copy non-regular files (e.g., directories,
        // symlinks, devices, etc.)
        return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
    }
    dfi, err := os.Stat(dst)
    if err != nil {
        if !os.IsNotExist(err) {
            return
        }
    } else {
        if !(dfi.Mode().IsRegular()) {
            return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
        }
        if os.SameFile(sfi, dfi) {
            return
        }
    }
    if err = os.Link(src, dst); err == nil {
        return
    }
    err = copyFileContents(src, dst)
    return
}

func copyFileContents(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()
    return
}