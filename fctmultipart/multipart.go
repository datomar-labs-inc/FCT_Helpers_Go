package fctmultipart

import (
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/ferr"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/textproto"
	"strings"
	"sync"
)

type MultipartForm struct {
	pipeReader  *io.PipeReader
	pipeWriter  *io.PipeWriter
	fw          *multipart.Writer
	wg          *sync.WaitGroup
	writeMe     chan func() error
	waitStarted bool
}

func NewMultipartForm() *MultipartForm {
	pipeReader, pipeWriter := io.Pipe()

	fw := multipart.NewWriter(pipeWriter)

	mpf := &MultipartForm{
		pipeReader:  pipeReader,
		pipeWriter:  pipeWriter,
		fw:          fw,
		wg:          &sync.WaitGroup{},
		waitStarted: false,
		writeMe:     make(chan func() error, 2000),
	}

	return mpf
}

func (m *MultipartForm) startWait() {
	go func() {
		for {
			select {
			case writeFn, ok := <-m.writeMe:
				if writeFn != nil {
					err := writeFn()
					if err != nil {
						lggr.GetDetached("multipart-write").Error("failed to write to multipart form body", zap.Error(err))
					}
				}

				if !ok {
					m.writeMe = nil
				}
			}

			if m.writeMe == nil {
				break
			}
		}
	}()

	go func() {
		m.wg.Wait()

		close(m.writeMe)
		m.close()
	}()
}

func (m *MultipartForm) AddField(name string, value string) error {
	m.wg.Add(1)

	if !m.waitStarted {
		m.waitStarted = true
		m.startWait()
	}

	m.writeMe <- func() error {
		defer m.wg.Done()

		err := m.fw.WriteField(name, value)
		if err != nil {
			return ferr.Wrap(err)
		}

		return nil
	}

	return nil
}

func (m *MultipartForm) MustAddField(name string, value string) *MultipartForm {
	err := m.AddField(name, value)
	if err != nil {
		panic(err)
	}

	return m
}

func (m *MultipartForm) AddFile(name string, filename string, value io.Reader) error {
	m.wg.Add(1)

	if !m.waitStarted {
		m.waitStarted = true
		m.startWait()
	}

	m.writeMe <- func() error {
		defer m.wg.Done()

		writer, err := m.fw.CreateFormFile(name, filename)
		if err != nil {
			return ferr.Wrap(err)
		}

		_, err = io.Copy(writer, value)
		if err != nil {
			return ferr.Wrap(err)
		}

		return nil
	}

	return nil
}

func (m *MultipartForm) AddFileExtra(name string, filename string, contentType string, value io.Reader) error {
	m.wg.Add(1)

	if !m.waitStarted {
		m.waitStarted = true
		m.startWait()
	}

	header := make(textproto.MIMEHeader)

	header.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(name), escapeQuotes(filename)))

	header.Set("Content-Type", contentType)

	m.writeMe <- func() error {
		defer m.wg.Done()

		writer, err := m.fw.CreatePart(header)
		if err != nil {
			return ferr.Wrap(err)
		}

		_, err = io.Copy(writer, value)
		if err != nil {
			return ferr.Wrap(err)
		}

		return nil
	}

	return nil
}

func (m *MultipartForm) MustAddFile(name string, filename string, value io.Reader) *MultipartForm {
	err := m.AddFile(name, filename, value)
	if err != nil {
		panic(err)
	}

	return m
}

func (m *MultipartForm) MustAddFileExtra(name string, filename string, contentType string, value io.Reader) *MultipartForm {
	err := m.AddFileExtra(name, filename, contentType, value)
	if err != nil {
		panic(err)
	}

	return m
}

func (m *MultipartForm) GetReader() io.ReadCloser {
	return m.pipeReader
}

func (m *MultipartForm) GetBytes() ([]byte, error) {
	reader := m.GetReader()
	defer reader.Close()
	defer m.close()

	byteSlice, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	return byteSlice, nil
}

func (m *MultipartForm) FormDataContentType() string {
	return m.fw.FormDataContentType()
}

func (m *MultipartForm) close() error {
	fwErr := m.fw.Close()
	pwErr := m.pipeWriter.Close()

	var err error

	if pwErr != nil {
		err = fmt.Errorf("failed to close pipe writer: %w", pwErr)
	}

	if fwErr != nil {
		if err != nil {
			err = fmt.Errorf("failed to close form writer: %w, (%v)", fwErr, err)
		} else {
			err = fmt.Errorf("failed to close form writer: %w", fwErr)
		}
	}

	return ferr.Wrap(err)
}

func BuildMultipartForm(stringParams [][]string, fileID string, fileBody io.Reader) (*MultipartForm, error) {
	form := NewMultipartForm()

	for _, parameter := range stringParams {
		err := form.AddField(parameter[0], parameter[1])
		if err != nil {
			return nil, ferr.Wrap(err)
		}
	}

	err := form.AddFile("file", fileID, fileBody)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	return form, nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
