package build

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/suse/elemental/v3/internal/image"
	"github.com/suse/elemental/v3/pkg/log"
)

type Runner interface {
	Run(cmd string, args ...string) ([]byte, error)
	RunContextParseOutput(ctx context.Context, stdoutH, stderrH func(line string), cmd string, args ...string) error
}

//go:embed config.sh.tpl
var configScript string

func Run(definition *image.Definition, logger log.Logger, runner Runner) error {
	logger.Info("Pulling RKE2 image")

	kubernetesImagePath, err := getFilenameFromURL(definition.Release.KubernetesImage)
	if err != nil {
		return err
	}

	if err = downloadFile(kubernetesImagePath, definition.Release.KubernetesImage); err != nil {
		logger.Error("Failed to download RKE2 image")
		return err
	}

	logger.Info("Preparing configuration script")

	config := struct {
		User            string
		Password        string
		KubernetesImage string
	}{
		User:            definition.OperatingSystem.Users[0].Username,
		Password:        definition.OperatingSystem.Users[0].Password,
		KubernetesImage: kubernetesImagePath,
	}

	tmpl, err := template.New("script").Parse(configScript)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	if err = os.WriteFile("config.sh", buf.Bytes(), 0755); err != nil {
		return fmt.Errorf("writing config.sh: %w", err)
	}

	logOutput := func(line string) {
		level, message := parseLogrusLine(line)
		logger.Log(level, message)
	}

	logger.Info("Creating RAW image")

	ctx := context.Background()

	if err = runner.RunContextParseOutput(ctx, logOutput, nil,
		"qemu-img", "create", "-f", "raw", definition.Image.OutputImageName, "10G"); err != nil {
		return fmt.Errorf("running qemu-img create: %w", err)
	}

	device, err := runner.Run("losetup", "-f", "--show", definition.Image.OutputImageName)
	if err != nil {
		return fmt.Errorf("running losetup -f: %w", err)
	}

	attachedDevice := strings.TrimSpace(string(device))
	logger.Info("Attached device: %s", attachedDevice)

	defer func() {
		if dErr := runner.RunContextParseOutput(ctx, logOutput, nil,
			"losetup", "-d", attachedDevice); dErr != nil {
			logger.Error("Detaching device failed: %v", dErr)
		}
	}()

	err = runner.RunContextParseOutput(ctx, logOutput, logOutput,
		"./elemental-toolkit", "--debug", "install", "--os-image", definition.Release.OperatingSystemImage, "--target", attachedDevice)
	if err != nil {
		return fmt.Errorf("running elemental-toolkit install: %w", err)
	}

	return nil
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func getFilenameFromURL(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return filepath.Base(parsedURL.Path), nil
}

func parseLogrusLine(line string) (logrus.Level, string) {
	re := regexp.MustCompile(`time="([^"]+)" level=([^ ]+) msg="([^"]+)"(.*)`)
	match := re.FindStringSubmatch(line)
	if len(match) < 4 {
		return logrus.InfoLevel, line
	}

	levelStr := match[2]
	msg := match[3]

	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return logrus.InfoLevel, line
	}

	return level, msg
}
