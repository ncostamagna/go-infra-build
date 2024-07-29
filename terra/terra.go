package terra

import (
	"bufio"
	"fmt"
	"os"
	"slices"

	cp "github.com/otiai10/copy" // When the Go 1.23 version is released, we can use CopyFS: https://pkg.go.dev/os@master#CopyFS
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
		if err := cp.Copy(t.config.CommonFrom, fmt.Sprintf(t.config.CommonTo, t.env)); err != nil {
			return fmt.Errorf("template copy error: %v", err)
		}
	}

	if t.config.IsTemplateAvailable {
		fmt.Println("Copy templates")
		if err := cp.Copy(t.config.TemplateFrom, fmt.Sprintf(t.config.TemplateTo, t.env)); err != nil {
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
		if err := cp.Copy(fmt.Sprintf(t.config.ReusablesFrom, sc.Text()), fmt.Sprintf(t.config.ReusablesTo, t.env, sc.Text())); err != nil {
			return fmt.Errorf("file copy error: %v", err)
		}
	}

	if err := sc.Err(); err != nil {
		return fmt.Errorf("scan file error: %v", err)
	}

	return nil
}