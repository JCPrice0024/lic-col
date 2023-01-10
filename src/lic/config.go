package lic

import (
	"encoding/json"
	"fmt"
	"os"
)

// initJsonConfigs decodes filename into the interface i. This makes it easier to
// get the required information from the Config files.
func initJsonConfigs(filename string, i interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file:  file: %s err: %w", filename, err)
	}
	fs, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error stat file:  file: %s err: %w", filename, err)
	}
	// an empty file is not an error
	if fs.Size() < 2 {
		return nil
	}
	dec := json.NewDecoder(file)
	err = dec.Decode(i)
	file.Close()
	return err
}
