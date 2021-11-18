package ui

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/moorara/promptui"
	"github.com/stretchr/testify/assert"
)

var (
	item1 = Item{
		Key:         "http",
		Name:        "HTTP",
		Description: "Hypertext Transfer Protocol",
		Attributes: []Attribute{
			{
				Key:   "Protocol",
				Value: "HTTP/2",
			},
		},
	}

	item2 = Item{
		Key:         "grpc",
		Name:        "gRPC",
		Description: "Google's Remote Procedure Calls",
		Attributes: []Attribute{
			{
				Key:   "Protocol",
				Value: "Protocol Bufffers",
			},
		},
	}
)

type nopCloser struct {
	io.Writer
}

func (w *nopCloser) Close() error {
	return nil
}

func mockWriter(w io.Writer) io.WriteCloser {
	return &nopCloser{
		Writer: w,
	}
}

// This method implements a hack for testing promptui.
// It pads input strings with a non-zero unused character up to 4096 in length.
// See https://github.com/moorara/promptui/issues/63
func mockReader(inputs ...string) io.ReadCloser {
	buf := new(bytes.Buffer)

	for _, input := range inputs {
		input += "\n"
		padded := make([]byte, 4096)
		copy(padded, []byte(input))
		for i := len(input); i < len(padded); i++ {
			padded[i] = 255
		}

		_, _ = buf.Write(padded)
	}

	return io.NopCloser(buf)
}

func TestNewInteractive(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{
			name:  "Info",
			level: Info,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := NewInteractive(tc.level)
			assert.NotNil(t, u)

			iu, ok := u.(*interactiveUI)
			assert.True(t, ok)
			assert.Equal(t, tc.level, iu.level)
			assert.NotNil(t, iu.reader)
			assert.NotNil(t, iu.writer)
			assert.NotNil(t, iu.errorWriter)
		})
	}
}

func TestInteractiveUI_Printf(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "OK",
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "Hello, World!\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			u := &interactiveUI{
				writer: mockWriter(buf),
			}

			u.Printf(tc.format, tc.args...)

			out, err := io.ReadAll(buf)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, string(out))
		})
	}
}

func TestInteractiveUI_GetLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{
			name:  "Debug",
			level: Debug,
		},
		{
			name:  "Info",
			level: Info,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &interactiveUI{
				level: tc.level,
			}

			level := u.GetLevel()

			assert.Equal(t, tc.level, level)
		})
	}
}

func TestInteractiveUI_SetLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{
			name:  "Debug",
			level: Debug,
		},
		{
			name:  "Info",
			level: Info,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &interactiveUI{}

			u.SetLevel(tc.level)

			assert.Equal(t, tc.level, u.level)
		})
	}
}

func TestInteractiveUI_Tracef(t *testing.T) {
	tests := []struct {
		name           string
		style          Style
		format         string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "OK_Blue",
			style:          Blue,
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[34mHello, World!\x1b[0m\n",
		},
		{
			name:           "OK_Mix",
			style:          Style{BgYellow, FgMagenta, Bold, BlinkSlow},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[43;35;1;5mHello, World!\x1b[0m\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			u := &interactiveUI{
				writer: mockWriter(buf),
			}

			u.Tracef(tc.style, tc.format, tc.args...)

			out, err := io.ReadAll(buf)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, string(out))
		})
	}
}

func TestInteractiveUI_Debugf(t *testing.T) {
	tests := []struct {
		name           string
		style          Style
		format         string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "OK_Cyan",
			style:          Cyan,
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[36mHello, World!\x1b[0m\n",
		},
		{
			name:           "OK_Mix",
			style:          Style{BgYellow, FgMagenta, Bold, BlinkSlow},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[43;35;1;5mHello, World!\x1b[0m\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			u := &interactiveUI{
				writer: mockWriter(buf),
			}

			u.Debugf(tc.style, tc.format, tc.args...)

			out, err := io.ReadAll(buf)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, string(out))
		})
	}
}

func TestInteractiveUI_Infof(t *testing.T) {
	tests := []struct {
		name           string
		style          Style
		format         string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "OK_Green",
			style:          Green,
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[32mHello, World!\x1b[0m\n",
		},
		{
			name:           "OK_Mix",
			style:          Style{BgYellow, FgMagenta, Bold, BlinkSlow},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[43;35;1;5mHello, World!\x1b[0m\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			u := &interactiveUI{
				writer: mockWriter(buf),
			}

			u.Infof(tc.style, tc.format, tc.args...)

			out, err := io.ReadAll(buf)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, string(out))
		})
	}
}

func TestInteractiveUI_Warnf(t *testing.T) {
	tests := []struct {
		name           string
		style          Style
		format         string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "OK_Yellow",
			style:          Yellow,
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[33mHello, World!\x1b[0m\n",
		},
		{
			name:           "OK_Mix",
			style:          Style{BgYellow, FgMagenta, Bold, BlinkSlow},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[43;35;1;5mHello, World!\x1b[0m\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			u := &interactiveUI{
				writer: mockWriter(buf),
			}

			u.Warnf(tc.style, tc.format, tc.args...)

			out, err := io.ReadAll(buf)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, string(out))
		})
	}
}

func TestInteractiveUI_Errorf(t *testing.T) {
	tests := []struct {
		name           string
		style          Style
		format         string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "OK_Red",
			style:          Red,
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[31mHello, World!\x1b[0m\n",
		},
		{
			name:           "OK_Mix",
			style:          Style{BgYellow, FgMagenta, Bold, BlinkSlow},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedOutput: "\x1b[43;35;1;5mHello, World!\x1b[0m\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			u := &interactiveUI{
				errorWriter: mockWriter(buf),
			}

			u.Errorf(tc.style, tc.format, tc.args...)

			out, err := io.ReadAll(buf)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, string(out))
		})
	}
}

func TestInteractiveUI_Confrim(t *testing.T) {
	tests := []struct {
		name           string
		inputs         []string
		prompt         string
		Default        bool
		expectedResult bool
		expectedError  string
	}{
		{
			name:           "DefaultNo_InputYes",
			inputs:         []string{"y"},
			prompt:         "Confirm",
			Default:        false,
			expectedResult: true,
			expectedError:  "",
		},
		{
			name:           "DefaultNo_InputNo",
			inputs:         []string{"n"},
			prompt:         "Confirm",
			Default:        false,
			expectedResult: false,
			expectedError:  "",
		},
		{
			name:           "DefaultNo_InputEmpty",
			inputs:         []string{""},
			prompt:         "Confirm",
			Default:        false,
			expectedResult: false,
			expectedError:  "",
		},
		{
			name:           "DefaultYes_InputNo",
			inputs:         []string{"n"},
			prompt:         "Confirm",
			Default:        true,
			expectedResult: false,
			expectedError:  "",
		},
		{
			name:           "DefaultYes_InputYes",
			inputs:         []string{"y"},
			prompt:         "Confirm",
			Default:        true,
			expectedResult: true,
			expectedError:  "",
		},
		{
			name:           "DefaultYes_InputEmpty",
			inputs:         []string{""},
			prompt:         "Confirm",
			Default:        true,
			expectedResult: true,
			expectedError:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &interactiveUI{
				reader: mockReader(tc.inputs...),
				writer: mockWriter(io.Discard),
			}

			result, err := u.Confrim(tc.prompt, tc.Default)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			} else {
				assert.False(t, result)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestInteractiveUI_Ask(t *testing.T) {
	tests := []struct {
		name          string
		inputs        []string
		prompt        string
		Default       string
		validate      ValidateFunc
		expectedValue string
		expectedError string
	}{
		{
			name:          "OK",
			inputs:        []string{"something"},
			prompt:        "Enter",
			Default:       "",
			validate:      nil,
			expectedValue: "something",
			expectedError: "",
		},
		{
			name:    "ValidationSucceeds",
			inputs:  []string{"something"},
			prompt:  "Enter",
			Default: "",
			validate: func(string) error {
				return nil
			},
			expectedValue: "something",
			expectedError: "",
		},
		{
			name:    "ValidationFails",
			inputs:  []string{"something"},
			prompt:  "Enter",
			Default: "",
			validate: func(string) error {
				return errors.New("invalid input")
			},
			expectedValue: "",
			expectedError: promptui.ErrEOF.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &interactiveUI{
				reader: mockReader(tc.inputs...),
				writer: mockWriter(io.Discard),
			}

			val, err := u.Ask(tc.prompt, tc.Default, tc.validate)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, val)
			} else {
				assert.Empty(t, val)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestInteractiveUI_AskSecret(t *testing.T) {
	tests := []struct {
		name          string
		inputs        []string
		prompt        string
		confirm       bool
		validate      ValidateFunc
		expectedValue string
		expectedError string
	}{
		{
			name:    "FirstValidationFails",
			inputs:  []string{"secret"},
			prompt:  "Enter",
			confirm: false,
			validate: func(string) error {
				return errors.New("invalid secret")
			},
			expectedValue: "",
			expectedError: promptui.ErrEOF.Error(),
		},
		{
			name:    "Success_NoConfirm",
			inputs:  []string{"secret"},
			prompt:  "Enter",
			confirm: false,
			validate: func(string) error {
				return nil
			},
			expectedValue: "secret",
			expectedError: "",
		},
		{
			name:    "SecondValidationFails",
			inputs:  []string{"secret", "cipher"},
			prompt:  "Enter",
			confirm: true,
			validate: func(val string) error {
				if val == "cipher" {
					return errors.New("invalid secret")
				}
				return nil
			},
			expectedValue: "",
			expectedError: promptui.ErrEOF.Error(),
		},
		{
			name:    "ConfirmationNotMatching",
			inputs:  []string{"secret", "cipher"},
			prompt:  "Enter",
			confirm: true,
			validate: func(string) error {
				return nil
			},
			expectedValue: "",
			expectedError: "confirmation not matching",
		},
		{
			name:    "Success_withConfirm",
			inputs:  []string{"secret", "secret"},
			prompt:  "Enter",
			confirm: true,
			validate: func(string) error {
				return nil
			},
			expectedValue: "secret",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &interactiveUI{
				reader: mockReader(tc.inputs...),
				writer: mockWriter(io.Discard),
			}

			val, err := u.AskSecret(tc.prompt, tc.confirm, tc.validate)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, val)
			} else {
				assert.Empty(t, val)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestInteractiveUI_Select(t *testing.T) {
	tests := []struct {
		name          string
		inputs        []string
		prompt        string
		size          int
		items         []Item
		search        SearchFunc
		expectedItem  Item
		expectedError string
	}{
		{
			name:          "NoSelection",
			inputs:        []string{},
			prompt:        "Select",
			size:          4,
			items:         []Item{item1, item2},
			search:        nil,
			expectedError: promptui.ErrEOF.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &interactiveUI{
				reader: mockReader(tc.inputs...),
				writer: mockWriter(io.Discard),
			}

			item, err := u.Select(tc.prompt, tc.size, tc.items, tc.search)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedItem, item)
			} else {
				assert.Empty(t, item)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
