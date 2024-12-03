package ranking

import (
	"strings"
	"testing"
)

func Test_saveData(t *testing.T) {
	type args struct {
		filename string
		X        interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := saveData(tt.args.filename, tt.args.X); (err != nil) != tt.wantErr {
				t.Errorf("saveData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validateGeneratedFilename(t *testing.T, filename, base string) {
	// Check that the filename is not empty
	if filename == "" {
		t.Errorf("Generated filename is empty")
	}

	// Check that the filename contains the base string as a prefix
	if !strings.HasPrefix(filename, base+"_") {
		t.Errorf("Filename %s does not have the correct base prefix: %s_", filename, base)
	}

	// Check that the filename ends with the correct extension
	if !strings.HasSuffix(filename, ".gob") {
		t.Errorf("Filename %s does not have the correct .gob extension", filename)
	}
}

func Test_generateUniqueFilename(t *testing.T) {
	tests := []struct {
		name string
		base string
	}{
		{"Base_Testfile", "testfile"},
		{"Base_MyFile", "myfile"},
		{"Base_SpecialChars", "file_with_special_chars"},
	}

	numFiles := 500 // Number of filenames to generate for each test case

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Map to track generated filenames for uniqueness
			filenames := make(map[string]struct{})

			for i := 0; i < numFiles; i++ {
				got := generateUniqueFilename(tt.base)
				validateGeneratedFilename(t, got, tt.base)
				if _, exists := filenames[got]; exists {
					t.Errorf("Duplicate filename generated: %s", got)
				}
				filenames[got] = struct{}{}
			}
		})
	}
}
