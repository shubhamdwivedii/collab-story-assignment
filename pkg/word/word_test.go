package word

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateWord(t *testing.T) {
	err := ValidateWord("validword")
	require.NoError(t, err)

	err = ValidateWord("invalid word")
	require.Error(t, err)

	err = ValidateWord(" leadingspace")
	require.Error(t, err)

	err = ValidateWord("trailingspace ")
	require.Error(t, err)

	err = ValidateWord("multiple number of spaces")
	require.Error(t, err)

	err = ValidateWord("Num63r$&&SYmb0l$")
	require.NoError(t, err)

	err = ValidateWord("InvalidLengthOfTheWord")
	require.Error(t, err)
}
