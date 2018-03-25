package fb2_test

import (
	"io/ioutil"
	"os"
	"testing"

	"bitbucket.org/enlab/modules/fb2"
)

func TestParser(t *testing.T) {
	var (
		file     *os.File
		data     []byte
		result   fb2.FB2
		err      error
		filename = "test_books/455068.fb2"
	)

	if file, err = os.OpenFile(filename, os.O_RDONLY, 0666); err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	if data, err = ioutil.ReadAll(file); err != nil {
		t.Fatal(err)
	}

	p := fb2.New(data)

	if result, err = p.Unmarshal(); err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", result.Description.TitleInfo.Annotation)
}
