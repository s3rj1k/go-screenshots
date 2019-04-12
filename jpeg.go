package screenshot

import (
	"bytes"

	jseg "github.com/garyhouston/jpegsegs"
)

// addCOMtoJPEG - adds comment to JPEG stream.
func addCOMtoJPEG(in []byte, comment []byte) ([]byte, error) {

	reader := bytes.NewReader(in)

	writer := new(bytes.Buffer)

	scanner, err := jseg.NewScanner(reader)
	if err != nil {
		return nil, err
	}

	dumper, err := jseg.NewDumper(writer)
	if err != nil {
		return nil, err
	}

	jsegComEnd := false

	for {

		marker, buf, err := scanner.Scan()
		if err != nil {
			return nil, err
		}

		if !jsegComEnd &&
			(marker >= jseg.SOF0 &&
				marker <= jseg.SOF15 &&
				marker != jseg.DHT &&
				marker != jseg.JPG &&
				marker != jseg.DAC) {

			err := dumper.Dump(jseg.COM, comment)
			if err != nil {
				return nil, err
			}

			jsegComEnd = true
		}

		if marker == jseg.COM {
			continue
		}

		if err := dumper.Dump(marker, buf); err != nil {
			return nil, err
		}

		if marker == jseg.EOI {
			break
		}

	}

	return writer.Bytes(), nil
}
